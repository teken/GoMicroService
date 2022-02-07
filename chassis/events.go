package chassis

type Events struct {
	eventManager *EventManager
}

func (e Events) Subscribe(id string, action EventFunction) {
	e.eventManager.Subscribe(id, action)
}

func (e Events) EventPanicChannel() <-chan EventContext {
	return e.eventManager.eventPanicChannel
}
