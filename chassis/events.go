package chassis

import "encoding/json"

type Events struct {
	eventManager *EventManager
}

func (e Events) Subscribe(id string, action EventFunction) {
	e.eventManager.Subscribe(id, action)
}

func (e Events) EventPanicChannel() <-chan EventContext {
	return e.eventManager.eventPanicChannel
}

func (e Events) Serve() error {
	return e.eventManager.Serve()
}

func SendEvent(sourceUserId string, eventId string, contentType string, body []byte) error {
	return defaultEventManager.SendEvent(sourceUserId, eventId, contentType, body)
}

func SendJsonEvent(sourceUserId string, eventId string, body any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return defaultEventManager.SendEvent(sourceUserId, eventId, "application/json", data)
}
