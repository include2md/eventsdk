package sdk

import (
	"context"
	"time"
)

type Delivery struct {
	Subject string
	Data    []byte
	Ack     func() error
}

type Transport interface {
	Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error)
	Publish(ctx context.Context, subject string, data []byte) error
	Subscribe(ctx context.Context, subject string, durable string, handler func(context.Context, Delivery) error) error
	HandleRequest(ctx context.Context, subject string, handler RequestHandler) error
}
