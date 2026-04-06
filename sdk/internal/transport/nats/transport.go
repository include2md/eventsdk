package nats

import (
	"context"
	"fmt"
	"time"

	natslib "github.com/nats-io/nats.go"

	"github.com/include2md/eventsdk/sdk"
)

type Transport struct {
	conn *natslib.Conn
	js   natslib.JetStreamContext
}

func New(conn *natslib.Conn, js natslib.JetStreamContext) (*Transport, error) {
	if conn == nil {
		return nil, fmt.Errorf("nats conn is nil")
	}
	if js == nil {
		return nil, fmt.Errorf("jetstream context is nil")
	}
	return &Transport{conn: conn, js: js}, nil
}

func UnsafeForTest(conn *natslib.Conn, js natslib.JetStreamContext) *Transport {
	return &Transport{conn: conn, js: js}
}

func (t *Transport) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	if t.conn == nil {
		return nil, fmt.Errorf("nats conn is nil")
	}
	if t.conn.Status() != natslib.CONNECTED {
		return nil, fmt.Errorf("nats conn is not connected")
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	msg, err := t.conn.RequestWithContext(reqCtx, subject, data)
	if err != nil {
		return nil, fmt.Errorf("nats request failed: %w", err)
	}
	return msg.Data, nil
}

func (t *Transport) Publish(ctx context.Context, subject string, data []byte) error {
	if t.js == nil {
		return fmt.Errorf("jetstream context is nil")
	}

	_, err := t.js.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("jetstream publish failed: %w", err)
	}
	return nil
}

func (t *Transport) Subscribe(ctx context.Context, subject string, durable string, handler func(context.Context, sdk.Delivery) error) error {
	if t.js == nil {
		return fmt.Errorf("jetstream context is nil")
	}

	_, err := t.js.Subscribe(subject, func(msg *natslib.Msg) {
		delivery := sdk.Delivery{
			Data: msg.Data,
			Ack: func() error {
				return msg.Ack()
			},
		}
		_ = handler(ctx, delivery)
	}, natslib.Durable(durable), natslib.ManualAck())
	if err != nil {
		return fmt.Errorf("jetstream subscribe failed: %w", err)
	}

	return nil
}
