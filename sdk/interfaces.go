package sdk

import (
	"context"
	"time"
)

type Client interface {
	SendCommand(ctx context.Context, cmd Command) ([]byte, error)
	PublishEvent(ctx context.Context, event Event) error
}

type Service interface {
	RegisterHandler(eventType string, handler Handler)
	Run(ctx context.Context, cfg RunConfig) error
}

type Delivery struct {
	Data []byte
	Ack  func() error
}

type Transport interface {
	Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error)
	Publish(ctx context.Context, subject string, data []byte) error
	Subscribe(ctx context.Context, subject string, durable string, handler func(context.Context, Delivery) error) error
}
