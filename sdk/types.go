package sdk

import "context"

type Command struct {
	Name          string
	Payload       any
	CorrelationID string
}

type Event struct {
	Type          string
	Payload       any
	CorrelationID string
}

type Handler func(ctx context.Context, payload []byte) error
