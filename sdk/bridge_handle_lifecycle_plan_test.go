package sdk

import (
	"context"
	"testing"
	"time"

	bridgeint "github.com/include2md/eventsdk/sdk/internal/bridge"
)

func TestBuildHandleLifecycleRulesPublishesBeforeAndAfterEvents(t *testing.T) {
	tr := &planFakeTransport{nextRequest: []byte(`{"id":"r1"}`)}

	rules, err := bridgeint.BuildHandleLifecycleRules(tr, "cmd.app.billing.invoice.create")
	if err != nil {
		t.Fatalf("build rules: %v", err)
	}

	c := newClientWithOptions(tr, time.Second, bridgeint.Options{
		Rules: rules,
	})
	err = c.Handle(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		return []byte(`{"ok":true}`), nil
	})
	if err != nil {
		t.Fatalf("handle: %v", err)
	}

	if len(tr.publishSubjects) != 2 {
		t.Fatalf("expected 2 publish calls, got %d", len(tr.publishSubjects))
	}
	if tr.publishSubjects[0] != "run.app.billing.execution.started" {
		t.Fatalf("unexpected before subject: %s", tr.publishSubjects[0])
	}
	if tr.publishSubjects[1] != "evt.app.billing.invoice.create" {
		t.Fatalf("unexpected after subject: %s", tr.publishSubjects[1])
	}
}

func TestBuildHandleLifecycleRulesRejectsInvalidHandleSubject(t *testing.T) {
	tr := &planFakeTransport{}
	_, err := bridgeint.BuildHandleLifecycleRules(tr, "invalid-subject")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildHandleLifecycleRulesRejectsNonAppScope(t *testing.T) {
	tr := &planFakeTransport{}
	_, err := bridgeint.BuildHandleLifecycleRules(tr, "cmd.user.alice.invoice.create")
	if err == nil {
		t.Fatal("expected error")
	}
}

type planFakeTransport struct {
	nextRequest     []byte
	publishSubjects []string
}

func (f *planFakeTransport) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	return nil, nil
}

func (f *planFakeTransport) Publish(ctx context.Context, subject string, data []byte) error {
	f.publishSubjects = append(f.publishSubjects, subject)
	return nil
}

func (f *planFakeTransport) Subscribe(ctx context.Context, subject string, durable string, handler func(context.Context, Delivery) error) error {
	return nil
}

func (f *planFakeTransport) HandleRequest(ctx context.Context, subject string, handler RequestHandler) error {
	_, err := handler(ctx, f.nextRequest)
	return err
}
