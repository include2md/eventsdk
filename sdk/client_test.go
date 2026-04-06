package sdk_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/include2md/eventsdk/sdk"
	"github.com/include2md/eventsdk/sdk/internal/envelope"
	"github.com/include2md/eventsdk/sdk/internal/subject"
)

func TestClientSendCommand(t *testing.T) {
	tr := &fakeTransport{requestResp: []byte(`{"ok":true}`)}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)

	_, err := c.SendCommand(context.Background(), sdk.Command{Name: "CreateMessage", Payload: map[string]any{"title": "hi"}})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if tr.requestSubject != "TW.XX.inbox.command.create" {
		t.Fatalf("unexpected subject: %s", tr.requestSubject)
	}
}

func TestClientPublishEvent(t *testing.T) {
	tr := &fakeTransport{}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)

	err := c.PublishEvent(context.Background(), sdk.Event{Type: "UserRegistered", Payload: map[string]any{"id": "u1"}})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if tr.publishSubject != "TW.XX.sdk.event.UserRegistered" {
		t.Fatalf("unexpected subject: %s", tr.publishSubject)
	}

	var env envelope.EventEnvelope
	if err := json.Unmarshal(tr.publishData, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.EventType != "UserRegistered" || env.EventID == "" || env.CorrelationID == "" {
		t.Fatalf("unexpected envelope: %+v", env)
	}
}

type fakeTransport struct {
	requestSubject string
	requestData    []byte
	requestResp    []byte

	publishSubject string
	publishData    []byte
}

func (f *fakeTransport) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	f.requestSubject = subject
	f.requestData = data
	return f.requestResp, nil
}

func (f *fakeTransport) Publish(ctx context.Context, subject string, data []byte) error {
	f.publishSubject = subject
	f.publishData = data
	return nil
}

func (f *fakeTransport) Subscribe(ctx context.Context, subject string, durable string, handler func(context.Context, sdk.Delivery) error) error {
	return nil
}
