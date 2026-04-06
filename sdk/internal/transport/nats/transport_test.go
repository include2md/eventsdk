package nats_test

import (
	"context"
	"testing"
	"time"

	natslib "github.com/nats-io/nats.go"

	transportnats "github.com/include2md/eventsdk/sdk/internal/transport/nats"
)

func TestNewTransportValidateNil(t *testing.T) {
	_, err := transportnats.New(nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}

	_, err = transportnats.New(&natslib.Conn{}, nil)
	if err == nil {
		t.Fatal("expected js error")
	}
}

func TestRequestWithClosedConn(t *testing.T) {
	tr := transportnats.UnsafeForTest(&natslib.Conn{}, &fakeJS{})
	_, err := tr.Request(context.Background(), "sub", []byte("x"), time.Second)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPublishWithFailingJS(t *testing.T) {
	tr := transportnats.UnsafeForTest(&natslib.Conn{}, &fakeJS{publishErr: context.DeadlineExceeded})
	err := tr.Publish(context.Background(), "sub", []byte("x"))
	if err == nil {
		t.Fatal("expected error")
	}
}

type fakeJS struct {
	natslib.JetStreamContext
	publishErr error
}

func (f *fakeJS) Publish(subj string, data []byte, opts ...natslib.PubOpt) (*natslib.PubAck, error) {
	return nil, f.publishErr
}
