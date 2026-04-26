package sdk_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/include2md/eventsdk/sdk"
	"github.com/include2md/eventsdk/sdk/internal/envelope"
)

func TestClientRequest(t *testing.T) {
	tr := &fakeTransport{requestResp: []byte(`{"ok":true}`)}
	c := sdk.NewClient(tr, time.Second)

	_, err := c.Request(context.Background(), "TW.XX.user.command.create", map[string]any{"title": "hi"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if tr.requestSubject != "TW.XX.user.command.create" {
		t.Fatalf("unexpected subject: %s", tr.requestSubject)
	}
}

func TestClientPublishWrapsEnvelope(t *testing.T) {
	tr := &fakeTransport{}
	c := sdk.NewClient(tr, time.Second)

	err := c.Emit(context.Background(), "TW.XX.user.event.created", map[string]any{"id": "u1"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if tr.publishSubject != "TW.XX.user.event.created" {
		t.Fatalf("unexpected subject: %s", tr.publishSubject)
	}

	var env envelope.EventEnvelope
	if err := json.Unmarshal(tr.publishData, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	if env.EventType != "TW.XX.user.event.created" || env.EventID == "" || env.CorrelationID == "" {
		t.Fatalf("unexpected envelope: %+v", env)
	}
}

func TestClientPublishDoesNotAutoBridgeWhenPayloadMatchesInboxCreate(t *testing.T) {
	tr := &fakeTransport{}
	c := sdk.NewClient(tr, time.Second)

	err := c.Emit(context.Background(), "TW.XX.user.event.created", map[string]any{
		"userId":      "u1",
		"messageId":   "m1",
		"title":       "hello",
		"description": "world",
		"category":    "billing",
		"box":         "primary",
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tr.requestSubject != "" {
		t.Fatalf("expected no bridge request, got %s", tr.requestSubject)
	}
}

func TestClientPublishDoesNotBridgeWhenPayloadMissingRequiredFields(t *testing.T) {
	tr := &fakeTransport{}
	c := sdk.NewClient(tr, time.Second)

	err := c.Emit(context.Background(), "TW.XX.user.event.created", map[string]any{"title": "hello"})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tr.requestSubject != "" {
		t.Fatalf("expected no bridge request, got: %s", tr.requestSubject)
	}
}

func TestClientPublishFailsWhenDomainPublishFails(t *testing.T) {
	tr := &fakeTransport{publishErr: errors.New("publish failed")}
	c := sdk.NewClient(tr, time.Second)

	err := c.Emit(context.Background(), "TW.XX.user.event.created", map[string]any{
		"userId":      "u1",
		"messageId":   "m1",
		"title":       "hello",
		"description": "world",
		"category":    "billing",
		"box":         "primary",
	})
	if err == nil {
		t.Fatal("expected publish error")
	}
	if tr.requestSubject != "" {
		t.Fatalf("request should not run after publish failure, got %s", tr.requestSubject)
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

func (f *fakeTransport) HandleRequest(ctx context.Context, subject string, handler sdk.RequestHandler) error {
	return nil
}
