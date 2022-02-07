package chassis

import (
	"context"
	"fmt"
	"time"
)

type EventFunction func(eventContext EventContext)
type EventContext context.Context

type EventManager struct {
	registeredEvents  []RegisteredEvent
	eventPanicChannel chan EventContext

	options *EventManagerOptions
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

func NewEventManager(options *EventManagerOptions) *EventManager {
	if options == nil {
		options = DefaultEventManagerOptions
	}
	return &EventManager{
		[]RegisteredEvent{},
		make(chan EventContext, options.eventPanicChannelSize),
		options,
	}
}

func (em *EventManager) Subscribe(id string, action EventFunction) {
	em.registeredEvents = append(em.registeredEvents, RegisteredEvent{
		id:     id,
		action: action,
	})
}

func (em EventManager) NewEvent(id string, payload []byte) error {
	c := context.WithValue(context.Background(), "id", id)
	c = context.WithValue(c, "body", payload)

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

	finalC, canFunc := context.WithTimeout(c, em.options.eventTimeOut)

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
