package chassis

import (
	"context"
	"encoding/json"
)

type MessageContext[T any] struct {
	context.Context
}

func (r MessageContext[T]) EventId() string {
	return r.Value("event-id").(string)
}

func (r MessageContext[T]) SourceUserId() string {
	return r.Value("source-user-id").(string)
}

func (r MessageContext[T]) Payload() []byte {
	return r.Value("payload").([]byte)
}

func (r MessageContext[T]) PayloadType() string {
	return r.Value("payload-type").(string)
}

func (r MessageContext[T]) FromJson() (*T, error) {
	payload := r.Value("payload").([]byte)
	content := new(T)
	err := json.Unmarshal(payload, content)
	if err != nil {
		return content, err
	}
	return content, nil
}

func FromContext[T any](ctx context.Context) MessageContext[T] {
	return MessageContext[T]{
		ctx,
	}
}
