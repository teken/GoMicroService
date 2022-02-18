package chassis

import (
	"context"
	"fmt"
	"github.com/streadway/amqp"
	"time"
)

type EventFunction func(eventContext EventContext)
type EventContext context.Context

type EventManager struct {
	communication *RabbitCommunication

	registeredEvents  []RegisteredEvent
	eventPanicChannel chan EventContext

	options     *EventManagerOptions
	serviceInfo *ServiceInfo
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

func NewEventManager(com *RabbitCommunication, info *ServiceInfo, options *EventManagerOptions) *EventManager {
	if options == nil {
		options = DefaultEventManagerOptions
	}
	return &EventManager{
		com,
		[]RegisteredEvent{},
		make(chan EventContext, options.eventPanicChannelSize),
		options,
		info,
	}
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
		defer func(context RequestContext) {
			if r := recover(); r != nil {
				em.eventPanicChannel <- context
				fmt.Println("Recovering from panic:", r)
			}
		}(finalC)
		defer canFunc()
		handler(finalC)
	}()
	return nil
}

func (em *EventManager) Serve() error {
	ch, err := em.communication.Connection.Channel()
	if err != nil {
		return err
	}

	err = ch.ExchangeDeclare("events", "fanout", false, false, false, false, nil)
	if err != nil {
		return err
	}

	serviceLabel := em.serviceInfo.serviceName + ".events"

	q, err := ch.QueueDeclare(serviceLabel, false, false, true, false, nil)
	if err != nil {
		return err
	}

	err = ch.QueueBind(q.Name, serviceLabel, "events", false, nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(q.Name, serviceLabel, true, false, false, false, nil)
	if err != nil {
		return err
	}

	go em.consumeRabbit(msgs)

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
