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
	"strings"
	"sync"
)

type RouteManager struct {
	serviceId            *uuid.UUID
	communication        *chassis.RabbitCommunication
	routes               []Route
	outstandingResponses map[string]chan *chassis.RequestResponse
	serviceChannels      sync.Map
}

func NewRouteManager(id *uuid.UUID, comm *chassis.RabbitCommunication) *RouteManager {
	return &RouteManager{
		id,
		comm,
		make([]Route, 0),
		make(map[string]chan *chassis.RequestResponse),
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

	errChan := chann.NotifyClose(make(chan *amqp.Error))
	go func() {
		err, more := <-errChan
		if more {
			fmt.Println("Channel Closed due to: " + err.Reason)
			err := r.ConnectRoute()
			if err != nil {
				fmt.Println("Channel reconnect failed: " + err.Error())
			}
		}
	}()

	return nil
}

func (r Route) SendRequest(req *http.Request) (<-chan *chassis.RequestResponse, error) {
	err := r.ConnectRoute()
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	headers := amqp.Table{}
	headers["request-path"] = req.URL.Path
	headers["request-method"] = req.Method
	headers["request-matched-path"] = r.path

	correlationId := uuid.NewRandom().String()
	chann, _ := r.routeManager.serviceChannels.Load(r.service)
	err = chann.(*amqp.Channel).Publish("requests", r.service+".requests", false, false, amqp.Publishing{
		MessageId:     uuid.NewRandom().String(),
		ReplyTo:       "requests.responses." + r.routeManager.serviceId.String(),
		CorrelationId: correlationId,
		Headers:       headers,
		DeliveryMode:  amqp.Persistent,
		ContentType:   req.Header.Get("Content-Type"),
		Body:          b,
	})
	if err != nil {
		return nil, err
	}

	return r.routeManager.AwaitResponseOfRequest(correlationId), err
}

func (rm *RouteManager) GetRoute(method string, path string) *Route {
	for _, registered := range rm.routes {
		if strings.EqualFold(registered.method, method) && registered.matcher.MatchString(path) {
			return &registered
		}
	}
	return nil
}

func (rm *RouteManager) AwaitResponseOfRequest(correlationId string) <-chan *chassis.RequestResponse {
	responseChannel := make(chan *chassis.RequestResponse, 1)
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

	err = chann.QueueBind(q.Name, q.Name, "requests.responses", false, nil)
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

	resp := chassis.RequestResponse{
		StatusCode:  statusCode,
		Body:        body,
		ContentType: contentType,
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

func (rm *RouteManager) AddRoute(method string, path string, service string) error {
	for _, route := range rm.routes {
		if route.path == path && route.method == method {
			return errors.New("already existing route for " + method + ":" + path)
		}
	}

	compPath, err := regexp.Compile(chassis.PathToRegex(path))
	if err != nil {
		return err
	}
	rm.routes = append(rm.routes, Route{
		rm,
		path,
		method,
		service,
		compPath,
	})
	return nil
}
