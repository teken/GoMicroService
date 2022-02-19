package chassis

import (
	"context"
	"encoding/json"
)

type EventContext struct {
	context.Context
}

func (r EventContext) EventId() string {
	return r.Value("event-id").(string)
}

func (r EventContext) SourceUserId() string {
	return r.Value("source-user-id").(string)
}

func (r EventContext) Payload() []byte {
	return r.Value("payload").([]byte)
}

func (r EventContext) PayloadType() string {
	return r.Value("payload-type").(string)
}

func (r EventContext[T]) FromJson() (*T, error) {
	payload := r.Value("payload").([]byte)
	content := new(T)
	err := json.Unmarshal(payload, content)
	if err != nil {
		return content, err
	}
	return content, nil
}

func NewEventContext(ctx context.Context) *EventContext {
	return &EventContext{
		ctx,
	}
}
