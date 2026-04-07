package sdk

import (
	"context"
	"time"
)

type Client interface {
	Publish(ctx context.Context, subject string, payload any) error
	Request(ctx context.Context, subject string, payload any) ([]byte, error)
}

type Service interface {
	Subscribe(ctx context.Context, subject string, consumerName string, handler Handler) error
	HandleRequest(ctx context.Context, subject string, handler RequestHandler) error
}

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
