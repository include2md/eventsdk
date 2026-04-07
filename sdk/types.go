package sdk

import (
	"context"
	"time"
)

type Handler func(ctx context.Context, msg Message) error
type RequestHandler func(ctx context.Context, request []byte) ([]byte, error)

type Message struct {
	Subject       string
	EventID       string
	CorrelationID string
	Timestamp     time.Time
	Attempt       int
	Payload       []byte
}
