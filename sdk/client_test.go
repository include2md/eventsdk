package sdk_test

import (
	"context"
	"encoding/json"
	"errors"
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

func TestClientPublishEventPublishesEnvelope(t *testing.T) {
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

func TestClientPublishEventInboxBridgeAutoTriggeredWhenPayloadMatches(t *testing.T) {
	tr := &fakeTransport{requestResp: []byte(`{"ok":true}`)}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)

	err := c.PublishEvent(context.Background(), sdk.Event{
		Type: "AnyBusinessEvent",
		Payload: map[string]any{
			"userId":      "u1",
			"messageId":   "m1",
			"title":       "hello",
			"description": "world",
			"category":    "billing",
			"box":         "primary",
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tr.requestSubject != "TW.XX.inbox.command.create" {
		t.Fatalf("unexpected bridge request subject: %s", tr.requestSubject)
	}
}

func TestClientPublishEventInboxBridgeSkippedWhenPayloadMissingRequiredFields(t *testing.T) {
	tr := &fakeTransport{}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)

	err := c.PublishEvent(context.Background(), sdk.Event{
		Type:    "AnyBusinessEvent",
		Payload: map[string]any{"title": "hello"},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tr.requestSubject != "" {
		t.Fatalf("bridge should be skipped, got request subject: %s", tr.requestSubject)
	}
}

func TestClientPublishEventFailsWhenDomainPublishFails(t *testing.T) {
	tr := &fakeTransport{publishErr: errors.New("publish failed")}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)

	err := c.PublishEvent(context.Background(), sdk.Event{
		Type: "AnyBusinessEvent",
		Payload: map[string]any{
			"userId":      "u1",
			"messageId":   "m1",
			"title":       "hello",
			"description": "world",
			"category":    "billing",
			"box":         "primary",
		},
	})
	if err == nil {
		t.Fatal("expected publish error")
	}
	if tr.requestSubject != "" {
		t.Fatalf("bridge should not execute when publish fails, got request to %s", tr.requestSubject)
	}
}

func TestClientPublishEventBridgeReplyNotOKDoesNotFailPublishResult(t *testing.T) {
	tr := &fakeTransport{requestResp: []byte(`{"ok":false,"code":"BAD_REQUEST","message":"bad request"}`)}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)

	err := c.PublishEvent(context.Background(), sdk.Event{
		Type: "AnyBusinessEvent",
		Payload: map[string]any{
			"userId":      "u1",
			"messageId":   "m1",
			"title":       "hello",
			"description": "world",
			"category":    "billing",
			"box":         "primary",
		},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

type fakeTransport struct {
	requestSubject string
	requestData    []byte
	requestResp    []byte
	requestErr     error

	publishSubject string
	publishData    []byte
	publishErr     error
}

func (f *fakeTransport) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	f.requestSubject = subject
	f.requestData = data
	if f.requestErr != nil {
		return nil, f.requestErr
	}
	return f.requestResp, nil
}

func (f *fakeTransport) Publish(ctx context.Context, subject string, data []byte) error {
	f.publishSubject = subject
	f.publishData = data
	return f.publishErr
}

func (f *fakeTransport) Subscribe(ctx context.Context, subject string, durable string, handler func(context.Context, sdk.Delivery) error) error {
	return nil
}
