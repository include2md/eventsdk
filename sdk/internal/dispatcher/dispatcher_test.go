package dispatcher_test

import (
	"context"
	"errors"
	"testing"

	"github.com/include2md/eventsdk/sdk/internal/dispatcher"
	"github.com/include2md/eventsdk/sdk/internal/envelope"
	"github.com/include2md/eventsdk/sdk/internal/registry"
	"github.com/include2md/eventsdk/sdk/internal/retry"
)

func TestHandleDecodeError(t *testing.T) {
	d := dispatcher.New(registry.New(), retry.NewPolicy(3), &fakeRetryPublisher{})

	if err := d.Handle(context.Background(), []byte("not-json")); err == nil {
		t.Fatal("expected decode error")
	}
}

func TestHandleNoHandler(t *testing.T) {
	d := dispatcher.New(registry.New(), retry.NewPolicy(3), &fakeRetryPublisher{})
	raw, _ := envelope.Marshal(envelope.EventEnvelope{EventType: "Unknown", Payload: map[string]any{"x": 1}, Attempt: 1})

	if err := d.Handle(context.Background(), raw); err == nil {
		t.Fatal("expected no handler error")
	}
}

func TestHandleSuccess(t *testing.T) {
	r := registry.New()
	called := false
	r.Register("UserRegistered", func(ctx context.Context, payload []byte) error {
		called = true
		return nil
	})

	d := dispatcher.New(r, retry.NewPolicy(3), &fakeRetryPublisher{})
	raw, _ := envelope.Marshal(envelope.EventEnvelope{EventType: "UserRegistered", Payload: map[string]any{"id": "u1"}, Attempt: 1})

	if err := d.Handle(context.Background(), raw); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !called {
		t.Fatal("expected handler called")
	}
}

func TestHandleFailureTriggersRetry(t *testing.T) {
	r := registry.New()
	r.Register("UserRegistered", func(ctx context.Context, payload []byte) error {
		return errors.New("boom")
	})

	retryPub := &fakeRetryPublisher{}
	d := dispatcher.New(r, retry.NewPolicy(3), retryPub)
	raw, _ := envelope.Marshal(envelope.EventEnvelope{EventType: "UserRegistered", Payload: map[string]any{"id": "u1"}, Attempt: 1})

	if err := d.Handle(context.Background(), raw); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if retryPub.calls != 1 {
		t.Fatalf("expected retry republish, got %d", retryPub.calls)
	}
	if retryPub.last.Attempt != 2 {
		t.Fatalf("expected attempt 2, got %d", retryPub.last.Attempt)
	}
}

func TestHandleFailureNoRetryAfterMax(t *testing.T) {
	r := registry.New()
	r.Register("UserRegistered", func(ctx context.Context, payload []byte) error {
		return errors.New("boom")
	})

	retryPub := &fakeRetryPublisher{}
	d := dispatcher.New(r, retry.NewPolicy(3), retryPub)
	raw, _ := envelope.Marshal(envelope.EventEnvelope{EventType: "UserRegistered", Payload: map[string]any{"id": "u1"}, Attempt: 3})

	if err := d.Handle(context.Background(), raw); err == nil {
		t.Fatal("expected terminal error")
	}
	if retryPub.calls != 0 {
		t.Fatalf("did not expect retry call, got %d", retryPub.calls)
	}
}

type fakeRetryPublisher struct {
	calls int
	last  envelope.EventEnvelope
}

func (f *fakeRetryPublisher) Republish(ctx context.Context, env envelope.EventEnvelope) error {
	f.calls++
	f.last = env
	return nil
}
