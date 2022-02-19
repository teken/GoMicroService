package chassis

import (
	"context"
	"fmt"
	"github.com/pborman/uuid"
	"github.com/streadway/amqp"
	"time"
)

type EventFunction func(eventContext *EventContext)

type EventManager struct {
	communication *RabbitCommunication

	registeredEvents  []RegisteredEvent
	eventPanicChannel chan *EventContext

	options            *EventManagerOptions
	serviceInfo        *ServiceInfo
	eventRabbitChannel *amqp.Channel
}

type EventManagerOptions struct {
	eventPanicChannelSize int
	eventTimeOut          time.Duration
}

var DefaultEventManagerOptions = &EventManagerOptions{
	10, time.Minute,
}

type RegisteredEvent struct {
	id     string
	action EventFunction
}

var defaultEventManager *EventManager

func NewEventManager(com *RabbitCommunication, info *ServiceInfo, options *EventManagerOptions) *EventManager {
	if options == nil {
		options = DefaultEventManagerOptions
	}

	defaultEventManager = &EventManager{
		com,
		[]RegisteredEvent{},
		make(chan *EventContext, options.eventPanicChannelSize),
		options,
		info,
		nil,
	}
	return defaultEventManager
}

func (em *EventManager) Subscribe(id string, action EventFunction) {
	em.registeredEvents = append(em.registeredEvents, RegisteredEvent{
		id:     id,
		action: action,
	})
}

func (em EventManager) NewEvent(id string, payload []byte, contentType string) error {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "event-id", id)
	ctx = context.WithValue(ctx, "payload", payload)
	ctx = context.WithValue(ctx, "payload-type", contentType)

	var handler EventFunction
	for _, registered := range em.registeredEvents {
		if registered.id == id {
			handler = registered.action
			break
		}
	}

	if handler == nil {
		return nil
	}

	finalC, canFunc := context.WithTimeout(ctx, em.options.eventTimeOut)

	go func() {
		ectx := NewEventContext(finalC)
		defer func(context *EventContext) {
			if r := recover(); r != nil {
				em.eventPanicChannel <- context
				fmt.Println("Recovering from panic:", r)
			}
		}(ectx)
		defer canFunc()
		handler(ectx)
	}()
	return nil
}

func (em *EventManager) Serve() error {
	var err error
	em.eventRabbitChannel, err = em.communication.Connection.Channel()
	if err != nil {
		return err
	}

	err = em.eventRabbitChannel.ExchangeDeclare("events", "fanout", false, false, false, false, nil)
	if err != nil {
		return err
	}

	serviceLabel := em.serviceInfo.serviceName + ".events"

	q, err := em.eventRabbitChannel.QueueDeclare(serviceLabel, false, false, true, false, nil)
	if err != nil {
		return err
	}

	err = em.eventRabbitChannel.QueueBind(q.Name, serviceLabel, "events", false, nil)
	if err != nil {
		return err
	}

	msgs, err := em.eventRabbitChannel.Consume(q.Name, serviceLabel, true, false, false, false, nil)
	if err != nil {
		return err
	}

	go em.consumeRabbit(msgs)

	errChan := em.eventRabbitChannel.NotifyClose(make(chan *amqp.Error))
	go func() {
		err, more := <-errChan
		if more {
			fmt.Println("Channel Closed due to: " + err.Reason)
			err := em.Serve()
			if err != nil {
				fmt.Println("Channel reconnect failed: " + err.Error())
			}
		}
	}()

	return nil
}

func (em *EventManager) consumeRabbit(msgs <-chan amqp.Delivery) {
	for msg := range msgs {
		eventId, exists := msg.Headers["event-id"].(string)
		if !exists {
			fmt.Println("EventManager: consumeRabbit: Path not provided")
			continue
		}
		if err := em.NewEvent(eventId, msg.Body, msg.ContentType); err != nil {
			fmt.Println("EventManager: consumeRabbit: " + err.Error())
		}
	}
}

func (em *EventManager) SendEvent(sourceUserId string, eventId string, contentType string, body []byte) error {
	headers := amqp.Table{}
	headers["event-id"] = eventId
	headers["source-user-id"] = sourceUserId

	err := em.eventRabbitChannel.Publish("events", eventId, false, false, amqp.Publishing{
		MessageId:    uuid.NewRandom().String(),
		Headers:      headers,
		DeliveryMode: amqp.Persistent,
		ContentType:  contentType,
		Body:         body,
	})
	if err != nil {
		return err
	}
	return nil
}
