package bridge

import (
	"context"
	"testing"
)

func TestBuildHandleLifecycleRulesPublishesBeforeAndAfterEvents(t *testing.T) {
	publisher := &planFakePublisher{}

	rules, err := BuildHandleLifecycleRules(publisher, "cmd.app.billing.invoice.create")
	if err != nil {
		t.Fatalf("build rules: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}

	beforeCtx := &Context{
		Stage:   StageBeforeHandle,
		Subject: "cmd.app.billing.invoice.create",
		Request: []byte(`{"id":"r1"}`),
	}
	if err := rules[0].Run(context.Background(), beforeCtx); err != nil {
		t.Fatalf("run before rule: %v", err)
	}

	afterCtx := &Context{
		Stage:    StageAfterHandle,
		Subject:  "cmd.app.billing.invoice.create",
		Request:  []byte(`{"id":"r1"}`),
		Response: []byte(`{"ok":true}`),
	}
	if err := rules[1].Run(context.Background(), afterCtx); err != nil {
		t.Fatalf("run after rule: %v", err)
	}

	if len(publisher.publishSubjects) != 2 {
		t.Fatalf("expected 2 publish calls, got %d", len(publisher.publishSubjects))
	}
	if publisher.publishSubjects[0] != "run.app.billing.execution.started" {
		t.Fatalf("unexpected before subject: %s", publisher.publishSubjects[0])
	}
	if publisher.publishSubjects[1] != "evt.app.billing.invoice.create" {
		t.Fatalf("unexpected after subject: %s", publisher.publishSubjects[1])
	}
}

func TestBuildHandleLifecycleRulesRejectsInvalidHandleSubject(t *testing.T) {
	_, err := BuildHandleLifecycleRules(&planFakePublisher{}, "invalid-subject")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBuildHandleLifecycleRulesRejectsNonAppScope(t *testing.T) {
	_, err := BuildHandleLifecycleRules(&planFakePublisher{}, "cmd.user.alice.invoice.create")
	if err == nil {
		t.Fatal("expected error")
	}
}

type planFakePublisher struct {
	publishSubjects []string
}

func (f *planFakePublisher) Publish(ctx context.Context, subject string, data []byte) error {
	f.publishSubjects = append(f.publishSubjects, subject)
	return nil
}
