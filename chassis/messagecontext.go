package chassis

import (
	"context"
	"encoding/json"
	"errors"
	"google.golang.org/protobuf/proto"
)

type MessageContext[T any] struct {
	context.Context
}

func (r MessageContext[T]) EventId() string {
	return r.Value("event-id").(string)
}

func (r MessageContext[T]) Path() string {
	return r.Value("path").(string)
}

func (r MessageContext[T]) Method() string {
	return r.Value("method").(string)
}

func (r MessageContext[T]) Payload() []byte {
	return r.Value("payload").([]byte)
}

func (r MessageContext[T]) PayloadType() string {
	return r.Value("payload-type").(string)
}

func (r MessageContext[T]) FromBody() (T, error) {
	switch r.PayloadType() {
	case "application/x-protobuf":
		return r.FromProto()
	case "application/json":
		return r.FromJson()
	}
	return nil, errors.New("unknown mime type: " + r.PayloadType())
}

func (r MessageContext[T]) FromJson() (T, error) {
	payload := r.Value("payload").([]byte)
	content := T{}
	err := json.Unmarshal(payload, content)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func (r MessageContext[T]) FromProto() (T, error) {
	payload := r.Value("payload").([]byte)
	content := T{}
	err := proto.Unmarshal(payload, content)
	if err != nil {
		return nil, err
	}
	return content, nil
}

func FromContext[T any](ctx context.Context) MessageContext[T] {
	return MessageContext[T]{
		ctx,
	}
}
