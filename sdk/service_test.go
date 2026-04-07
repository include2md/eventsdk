package sdk_test

import (
	"context"
	"testing"
	"time"

	"github.com/include2md/eventsdk/sdk"
	"github.com/include2md/eventsdk/sdk/internal/envelope"
)

func TestServiceSubscribeDispatchesEnvelopeAndAck(t *testing.T) {
	tr := &fakeServiceTransport{}
	svc := sdk.NewService(tr)

	handled := false
	raw, _ := envelope.Marshal(envelope.EventEnvelope{EventType: "TW.XX.user.event.created", Payload: map[string]any{"id": "u1"}, Attempt: 1, EventID: "e1", CorrelationID: "c1", Timestamp: time.Now().UTC()})
	tr.nextDelivery = sdk.Delivery{
		Subject: "TW.XX.user.event.created",
		Data:    raw,
		Ack: func() error {
			tr.acked = true
			return nil
		},
	}

	err := svc.Subscribe(context.Background(), "TW.XX.user.event.*", "consumer-a", func(ctx context.Context, msg sdk.Message) error {
		handled = true
		if msg.Subject != "TW.XX.user.event.created" {
			t.Fatalf("unexpected subject: %s", msg.Subject)
		}
		if msg.EventID != "e1" || msg.CorrelationID != "c1" {
			t.Fatalf("unexpected metadata: %+v", msg)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tr.subscribeSubject != "TW.XX.user.event.*" {
		t.Fatalf("unexpected subscribe subject: %s", tr.subscribeSubject)
	}
	if !handled {
		t.Fatal("expected handler called")
	}
	if !tr.acked {
		t.Fatal("expected ack")
	}
}

func TestServiceHandleRequest(t *testing.T) {
	tr := &fakeServiceTransport{}
	svc := sdk.NewService(tr)

	err := svc.HandleRequest(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		return []byte(`{"ok":true}`), nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tr.handleRequestSubject != "TW.XX.user.command.create" {
		t.Fatalf("unexpected request subject: %s", tr.handleRequestSubject)
	}
}

type fakeServiceTransport struct {
	subscribeSubject     string
	subscribeDurable     string
	nextDelivery         sdk.Delivery
	acked                bool
	handleRequestSubject string
}

func (f *fakeServiceTransport) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	return nil, nil
}

func (f *fakeServiceTransport) Publish(ctx context.Context, subject string, data []byte) error {
	return nil
}

func (f *fakeServiceTransport) Subscribe(ctx context.Context, subject string, durable string, handler func(context.Context, sdk.Delivery) error) error {
	f.subscribeSubject = subject
	f.subscribeDurable = durable
	return handler(ctx, f.nextDelivery)
}

func (f *fakeServiceTransport) HandleRequest(ctx context.Context, subject string, handler sdk.RequestHandler) error {
	f.handleRequestSubject = subject
	return nil
}
