package sdk_test

import (
	"context"
	"testing"
	"time"

	"github.com/include2md/eventsdk/sdk"
	"github.com/include2md/eventsdk/sdk/internal/envelope"
	"github.com/include2md/eventsdk/sdk/internal/subject"
)

func TestServiceRunDispatchesEventAndAck(t *testing.T) {
	tr := &fakeServiceTransport{}
	svc := sdk.NewService(tr, subject.NewResolver("TW.XX"), 3)

	handled := false
	svc.RegisterHandler("UserRegistered", func(ctx context.Context, payload []byte) error {
		handled = true
		return nil
	})

	raw, _ := envelope.Marshal(envelope.EventEnvelope{EventType: "UserRegistered", Payload: map[string]any{"id": "u1"}, Attempt: 1})
	tr.nextDelivery = sdk.Delivery{
		Data: raw,
		Ack: func() error {
			tr.acked = true
			return nil
		},
	}

	err := svc.Run(context.Background(), sdk.RunConfig{Namespace: "TW.XX", ConsumerName: "consumer-a"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if tr.subscribeSubject != "TW.XX.sdk.event.*" {
		t.Fatalf("unexpected subscribe subject: %s", tr.subscribeSubject)
	}
	if !handled {
		t.Fatal("expected handler called")
	}
	if !tr.acked {
		t.Fatal("expected ack")
	}
}

type fakeServiceTransport struct {
	requestSubject string
	requestData    []byte
	requestResp    []byte

	publishSubject string
	publishData    []byte

	subscribeSubject string
	subscribeDurable string
	nextDelivery     sdk.Delivery
	acked            bool
}

func (f *fakeServiceTransport) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	f.requestSubject = subject
	f.requestData = data
	return f.requestResp, nil
}

func (f *fakeServiceTransport) Publish(ctx context.Context, subject string, data []byte) error {
	f.publishSubject = subject
	f.publishData = data
	return nil
}

func (f *fakeServiceTransport) Subscribe(ctx context.Context, subject string, durable string, handler func(context.Context, sdk.Delivery) error) error {
	f.subscribeSubject = subject
	f.subscribeDurable = durable
	return handler(ctx, f.nextDelivery)
}
