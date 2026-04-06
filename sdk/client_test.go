package sdk_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
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

func TestClientPublishEventBridgeSuccessDefaultMode(t *testing.T) {
	tr := &fakeTransport{requestResp: []byte(`{"ok":true}`)}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)
	c.RegisterBridgeRule(sdk.BridgeRule{
		EventType:   "UserRegistered",
		CommandName: "CreateMessage",
		MapPayload: func(event sdk.Event) (any, error) {
			return map[string]any{
				"userId":      "u1",
				"messageId":   "m1",
				"title":       "hello",
				"description": "world",
				"category":    "billing",
				"box":         "primary",
			}, nil
		},
	})

	err := c.PublishEvent(context.Background(), sdk.Event{
		Type:    "UserRegistered",
		Payload: map[string]any{"id": "u1"},
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tr.requestSubject != "TW.XX.inbox.command.create" {
		t.Fatalf("unexpected bridge request subject: %s", tr.requestSubject)
	}
}

func TestClientPublishEventBridgeFailureDefaultModeStillReturnsNil(t *testing.T) {
	tr := &fakeTransport{requestErr: errors.New("bridge request failed")}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)
	c.RegisterBridgeRule(sdk.BridgeRule{
		EventType:   "UserRegistered",
		CommandName: "CreateMessage",
		MapPayload: func(event sdk.Event) (any, error) {
			return map[string]any{
				"userId":      "u1",
				"messageId":   "m1",
				"title":       "hello",
				"description": "world",
				"category":    "billing",
				"box":         "primary",
			}, nil
		},
	})

	err := c.PublishEvent(context.Background(), sdk.Event{
		Type:    "UserRegistered",
		Payload: map[string]any{"id": "u1"},
	})
	if err != nil {
		t.Fatalf("expected nil in default mode, got %v", err)
	}
}

func TestClientPublishEventBridgeFailureStrictModeReturnsError(t *testing.T) {
	tr := &fakeTransport{requestErr: errors.New("bridge request failed")}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)
	c.SetBridgeMode(sdk.BridgeModeStrict)
	c.RegisterBridgeRule(sdk.BridgeRule{
		EventType:   "UserRegistered",
		CommandName: "CreateMessage",
		MapPayload: func(event sdk.Event) (any, error) {
			return map[string]any{
				"userId":      "u1",
				"messageId":   "m1",
				"title":       "hello",
				"description": "world",
				"category":    "billing",
				"box":         "primary",
			}, nil
		},
	})

	err := c.PublishEvent(context.Background(), sdk.Event{
		Type:    "UserRegistered",
		Payload: map[string]any{"id": "u1"},
	})
	if err == nil {
		t.Fatal("expected error in strict mode")
	}
}

func TestClientPublishEventFailsWhenDomainPublishFails(t *testing.T) {
	tr := &fakeTransport{publishErr: errors.New("publish failed")}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)
	c.RegisterBridgeRule(sdk.BridgeRule{
		EventType:   "UserRegistered",
		CommandName: "CreateMessage",
		MapPayload: func(event sdk.Event) (any, error) {
			return map[string]any{
				"userId":      "u1",
				"messageId":   "m1",
				"title":       "hello",
				"description": "world",
				"category":    "billing",
				"box":         "primary",
			}, nil
		},
	})

	err := c.PublishEvent(context.Background(), sdk.Event{
		Type:    "UserRegistered",
		Payload: map[string]any{"id": "u1"},
	})
	if err == nil {
		t.Fatal("expected publish error")
	}
	if tr.requestSubject != "" {
		t.Fatalf("bridge should not execute when publish fails, got request to %s", tr.requestSubject)
	}
}

func TestClientPublishEventBridgeObserverNotifiedOnFailure(t *testing.T) {
	tr := &fakeTransport{requestErr: errors.New("bridge request failed")}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)
	obs := &fakeBridgeObserver{}
	c.SetBridgeObserver(obs)
	c.RegisterBridgeRule(sdk.BridgeRule{
		EventType:   "UserRegistered",
		CommandName: "CreateMessage",
		MapPayload: func(event sdk.Event) (any, error) {
			return map[string]any{
				"userId":      "u1",
				"messageId":   "m1",
				"title":       "hello",
				"description": "world",
				"category":    "billing",
				"box":         "primary",
			}, nil
		},
	})

	err := c.PublishEvent(context.Background(), sdk.Event{
		Type:    "UserRegistered",
		Payload: map[string]any{"id": "u1"},
	})
	if err != nil {
		t.Fatalf("expected nil in default mode, got %v", err)
	}
	if obs.failureCalls != 1 {
		t.Fatalf("expected failure observer call, got %d", obs.failureCalls)
	}
}

func TestClientPublishEventBridgeCreatePayloadValidationDefaultMode(t *testing.T) {
	tr := &fakeTransport{}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)
	obs := &fakeBridgeObserver{}
	c.SetBridgeObserver(obs)
	c.RegisterBridgeRule(sdk.BridgeRule{
		EventType:   "UserRegistered",
		CommandName: "CreateMessage",
		MapPayload: func(event sdk.Event) (any, error) {
			return map[string]any{
				"title": "hello",
			}, nil
		},
	})

	err := c.PublishEvent(context.Background(), sdk.Event{Type: "UserRegistered", Payload: map[string]any{"id": "u1"}})
	if err != nil {
		t.Fatalf("expected nil in default mode, got %v", err)
	}
	if tr.requestSubject != "" {
		t.Fatal("bridge request should not be sent for invalid payload")
	}
	if obs.failureCalls != 1 {
		t.Fatalf("expected observer failure call, got %d", obs.failureCalls)
	}
}

func TestClientPublishEventBridgeCreatePayloadValidationStrictMode(t *testing.T) {
	tr := &fakeTransport{}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)
	c.SetBridgeMode(sdk.BridgeModeStrict)
	c.RegisterBridgeRule(sdk.BridgeRule{
		EventType:   "UserRegistered",
		CommandName: "CreateMessage",
		MapPayload: func(event sdk.Event) (any, error) {
			return map[string]any{"title": "hello"}, nil
		},
	})

	err := c.PublishEvent(context.Background(), sdk.Event{Type: "UserRegistered", Payload: map[string]any{"id": "u1"}})
	if err == nil {
		t.Fatal("expected error in strict mode")
	}
	if !strings.Contains(err.Error(), "missing required field") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClientPublishEventBridgeReplyNotOK(t *testing.T) {
	tr := &fakeTransport{requestResp: []byte(`{"ok":false,"code":"BAD_REQUEST","message":"bad request"}`)}
	c := sdk.NewClient(tr, subject.NewResolver("TW.XX"), time.Second)
	c.SetBridgeMode(sdk.BridgeModeStrict)
	c.RegisterBridgeRule(sdk.BridgeRule{
		EventType:   "UserRegistered",
		CommandName: "CreateMessage",
		MapPayload: func(event sdk.Event) (any, error) {
			return map[string]any{
				"userId":      "u1",
				"messageId":   "m1",
				"title":       "hello",
				"description": "world",
				"category":    "billing",
				"box":         "primary",
			}, nil
		},
	})

	err := c.PublishEvent(context.Background(), sdk.Event{Type: "UserRegistered", Payload: map[string]any{"id": "u1"}})
	if err == nil {
		t.Fatal("expected strict mode error when bridge reply is not ok")
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

type fakeBridgeObserver struct {
	successCalls int
	failureCalls int
}

func (f *fakeBridgeObserver) OnBridgeSuccess(ctx context.Context, eventType string, correlationID string, commandName string) {
	f.successCalls++
}

func (f *fakeBridgeObserver) OnBridgeFailure(ctx context.Context, eventType string, correlationID string, commandName string, err error) {
	f.failureCalls++
}
