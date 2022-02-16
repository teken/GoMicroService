package main

import (
	"errors"
	"fmt"
	"github.com/pborman/uuid"
	"github.com/streadway/amqp"
	"github.com/teken/GoMicroService/chassis"
	"io"
	"net/http"
	"regexp"
	"sync"
)

type RouteManager struct {
	serviceId            *uuid.UUID
	communication        *chassis.RabbitCommunication
	routes               []Route
	outstandingResponses map[string]chan *RequestResponse
	serviceChannels      sync.Map
}

func NewRouteManager(id *uuid.UUID, comm *chassis.RabbitCommunication) *RouteManager {
	return &RouteManager{
		id,
		comm,
		make([]Route, 0),
		make(map[string]chan *RequestResponse),
		sync.Map{},
	}
}

type Route struct {
	routeManager *RouteManager
	path         string
	method       string
	service      string
	matcher      *regexp.Regexp
}

func (r Route) ConnectRoute() error {

	if _, exists := r.routeManager.serviceChannels.Load(r.service); exists {
		return nil
	}

	chann, err := r.routeManager.communication.Connection.Channel()
	if err != nil {
		return err
	}

	r.routeManager.serviceChannels.Store(r.service, chann)

	err = chann.ExchangeDeclare("requests", "topic", false, false, false, false, nil)
	if err != nil {
		return err
	}

	q, err := chann.QueueDeclare(r.service, false, false, true, false, nil)
	if err != nil {
		return err
	}

	err = chann.QueueBind(q.Name, r.service, "requests", false, nil)
	if err != nil {
		return err
	}

	errChan := chann.NotifyClose(make(chan *amqp.Error))
	go func() {
		err, more := <-errChan
		if more {
			fmt.Println("Channel Closed due to: " + err.Reason)
		}
	}()

	return nil
}

func (r Route) SendRequest(req *http.Request) (<-chan *RequestResponse, error) {
	err := r.ConnectRoute()
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	correlationId := uuid.NewRandom().String()
	chann, _ := r.routeManager.serviceChannels.Load(r.service)
	err = chann.(*amqp.Channel).Publish("requests", r.service, false, false, amqp.Publishing{
		MessageId:     uuid.NewRandom().String(),
		ReplyTo:       "requests.responses." + r.routeManager.serviceId.String(),
		CorrelationId: correlationId,
		DeliveryMode:  amqp.Persistent,
		ContentType:   req.Header.Get("Content-Type"),
		Body:          b,
	})
	if err != nil {
		return nil, err
	}

	return r.routeManager.AwaitResponseOfRequest(correlationId), err
}

type RequestResponse struct {
	StatusCode  int
	Body        []byte
	ContentType string
}

func (rm *RouteManager) GetRoute(method string, path string) *Route {
	for _, registered := range rm.routes {
		if registered.method == method && registered.matcher.MatchString(path) {
			return &registered
		}
	}
	return nil
}

func (rm *RouteManager) AwaitResponseOfRequest(correlationId string) <-chan *RequestResponse {
	responseChannel := make(chan *RequestResponse, 1)
	rm.outstandingResponses[correlationId] = responseChannel

	return responseChannel
}

func (rm *RouteManager) StartResponseListening() error {

	chann, err := rm.communication.Connection.Channel()
	if err != nil {
		return err
	}

	err = chann.ExchangeDeclare("requests.responses", "direct", false, false, false, false, nil)
	if err != nil {
		return err
	}

	q, err := chann.QueueDeclare("requests.responses."+rm.serviceId.String(), false, false, true, false, nil)
	if err != nil {
		return err
	}

	err = chann.QueueBind(q.Name, q.Name, "requests", false, nil)
	if err != nil {
		return err
	}

	msgs, err := chann.Consume(q.Name, q.Name, true, false, false, false, nil)
	if err != nil {
		return err
	}

	go rm.consumeResponses(msgs)

	return nil
}

func (rm *RouteManager) NewResponse(correlationId string, statusCode int, contentType string, body []byte) error {
	respChan, exists := rm.outstandingResponses[correlationId]
	if !exists {
		return errors.New("responses channel does not exist")
	}

	resp := RequestResponse{
		statusCode,
		body,
		contentType,
	}

	respChan <- &resp
	return nil
}

func (rm *RouteManager) consumeResponses(msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		statusCode, exists := msg.Headers["status-code"].(int)
		if !exists {
			fmt.Println("RouteManager: consumeResponses: Status Code not provided")
			statusCode = 200
		}
		if err := rm.NewResponse(msg.CorrelationId, statusCode, msg.ContentType, msg.Body); err != nil {
			fmt.Println("RouteManager: consumeResponses: " + err.Error())
		}
	}
}
