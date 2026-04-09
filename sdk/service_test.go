package sdk_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/include2md/eventsdk/sdk"
	"github.com/include2md/eventsdk/sdk/internal/envelope"
)

func TestServiceSubscribeDispatchesEnvelopeAndAck(t *testing.T) {
	tr := &fakeServiceTransport{}
	svc := sdk.NewClient(tr, time.Second)

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

	err := svc.Listen(context.Background(), "TW.XX.user.event.*", "consumer-a", func(ctx context.Context, msg sdk.Message) error {
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
	tr := &fakeServiceTransport{nextRequest: []byte(`{"name":"x"}`)}
	svc := sdk.NewClient(tr, time.Second)

	err := svc.Respond(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		return []byte(`{"ok":true}`), nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tr.handleRequestSubject != "TW.XX.user.command.create" {
		t.Fatalf("unexpected request subject: %s", tr.handleRequestSubject)
	}
	if string(tr.lastResponse) != `{"ok":true}` {
		t.Fatalf("unexpected response: %s", string(tr.lastResponse))
	}
}

func TestServiceHandleRequestAutoBridgeAfterSuccess(t *testing.T) {
	tr := &fakeServiceTransport{nextRequest: []byte(`{"userId":"u1","messageId":"m1","title":"hello","description":"world","category":"billing","box":"primary"}`), requestResp: []byte(`{"ok":true}`)}
	svc := sdk.NewClient(tr, time.Second)

	err := svc.Respond(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		return []byte(`{"ok":true}`), nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tr.requestSubject != "TW.XX.inbox.command.create" {
		t.Fatalf("expected inbox request, got %s", tr.requestSubject)
	}
}

func TestServiceHandleRequestNoBridgeWhenHandlerFails(t *testing.T) {
	tr := &fakeServiceTransport{nextRequest: []byte(`{"userId":"u1","messageId":"m1","title":"hello","description":"world","category":"billing","box":"primary"}`)}
	svc := sdk.NewClient(tr, time.Second)

	err := svc.Respond(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		return nil, errors.New("boom")
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tr.requestSubject != "" {
		t.Fatalf("did not expect inbox request, got %s", tr.requestSubject)
	}
}

type fakeServiceTransport struct {
	subscribeSubject     string
	subscribeDurable     string
	nextDelivery         sdk.Delivery
	acked                bool
	handleRequestSubject string
	nextRequest          []byte
	lastResponse         []byte

	requestSubject string
	requestData    []byte
	requestResp    []byte
}

func (f *fakeServiceTransport) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	f.requestSubject = subject
	f.requestData = data
	return f.requestResp, nil
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
	response, err := handler(ctx, f.nextRequest)
	if err != nil {
		return nil
	}
	f.lastResponse = response
	return nil
}
