package sdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	bridgeint "github.com/include2md/eventsdk/sdk/internal/bridge"
	"github.com/include2md/eventsdk/sdk/internal/envelope"
)

func TestServiceSubscribeDispatchesEnvelopeAndAck(t *testing.T) {
	tr := &fakeServiceTransport{}
	svc := NewClient(tr, time.Second)

	handled := false
	raw, _ := envelope.Marshal(envelope.EventEnvelope{EventType: "TW.XX.user.event.created", Payload: map[string]any{"id": "u1"}, Attempt: 1, EventID: "e1", CorrelationID: "c1", Timestamp: time.Now().UTC()})
	tr.nextDelivery = Delivery{
		Subject: "TW.XX.user.event.created",
		Data:    raw,
		Ack: func() error {
			tr.acked = true
			return nil
		},
	}

	err := svc.Listen(context.Background(), "TW.XX.user.event.*", "consumer-a", func(ctx context.Context, msg Message) error {
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
	svc := NewClient(tr, time.Second)

	err := svc.Handle(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
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

func TestServiceHandleRequestDoesNotAutoBridgeAfterSuccess(t *testing.T) {
	tr := &fakeServiceTransport{nextRequest: []byte(`{"userId":"u1","messageId":"m1","title":"hello","description":"world","category":"billing","box":"primary"}`), requestResp: []byte(`{"ok":true}`)}
	svc := NewClient(tr, time.Second)

	err := svc.Handle(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		return []byte(`{"ok":true}`), nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tr.requestSubject != "" {
		t.Fatalf("did not expect inbox request, got %s", tr.requestSubject)
	}
}

func TestServiceHandleRequestNoBridgeWhenHandlerFails(t *testing.T) {
	tr := &fakeServiceTransport{nextRequest: []byte(`{"userId":"u1","messageId":"m1","title":"hello","description":"world","category":"billing","box":"primary"}`)}
	svc := NewClient(tr, time.Second)

	err := svc.Handle(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		return nil, errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected handler error")
	}
	if tr.requestSubject != "" {
		t.Fatalf("did not expect inbox request, got %s", tr.requestSubject)
	}
}

func TestServiceHandleBeforeHookFailBlocksHandlerByDefault(t *testing.T) {
	tr := &fakeServiceTransport{nextRequest: []byte(`{"name":"x"}`)}
	before := &testBridgeRule{
		name:     "before-send",
		stage:    bridgeint.StageBeforeHandle,
		priority: 10,
		run: func(context.Context, *bridgeint.Context) error {
			return errors.New("before failed")
		},
	}
	svc := newClientWithOptions(tr, time.Second, bridgeint.Options{
		Rules: []bridgeint.Rule{before},
	})

	handlerCalled := false
	err := svc.Handle(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		handlerCalled = true
		return []byte(`{"ok":true}`), nil
	})
	if err == nil {
		t.Fatal("expected before hook error")
	}
	if handlerCalled {
		t.Fatal("handler should not run when before hook fails with default fail policy")
	}
}

func TestServiceHandleBeforeHookIgnoreAllowsHandler(t *testing.T) {
	tr := &fakeServiceTransport{nextRequest: []byte(`{"name":"x"}`)}
	mode := bridgeint.PolicyIgnore
	before := &testBridgeRule{
		name:     "before-send",
		stage:    bridgeint.StageBeforeHandle,
		priority: 10,
		policy:   &mode,
		run: func(context.Context, *bridgeint.Context) error {
			return errors.New("before failed")
		},
	}
	svc := newClientWithOptions(tr, time.Second, bridgeint.Options{
		Rules: []bridgeint.Rule{before},
	})

	handlerCalled := false
	err := svc.Handle(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		handlerCalled = true
		return []byte(`{"ok":true}`), nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !handlerCalled {
		t.Fatal("handler should run when before hook is ignore")
	}
}

func TestServiceHandleAfterHookFailDoesNotOverrideSuccessResponse(t *testing.T) {
	tr := &fakeServiceTransport{nextRequest: []byte(`{"name":"x"}`)}
	after := &testBridgeRule{
		name:     "after-send",
		stage:    bridgeint.StageAfterHandle,
		priority: 1,
		run: func(context.Context, *bridgeint.Context) error {
			return errors.New("after failed")
		},
	}
	svc := newClientWithOptions(tr, time.Second, bridgeint.Options{
		Rules: []bridgeint.Rule{after},
	})

	err := svc.Handle(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		return []byte(`{"ok":true}`), nil
	})
	if err != nil {
		t.Fatalf("after fail should not override handler success response: %v", err)
	}
	if string(tr.lastResponse) != `{"ok":true}` {
		t.Fatalf("unexpected response: %s", string(tr.lastResponse))
	}
}

func TestServiceHandleAfterHooksRunByPriority(t *testing.T) {
	tr := &fakeServiceTransport{nextRequest: []byte(`{"name":"x"}`)}
	order := make([]string, 0, 2)
	r1 := &testBridgeRule{
		name:     "late",
		stage:    bridgeint.StageAfterHandle,
		priority: 20,
		run: func(context.Context, *bridgeint.Context) error {
			order = append(order, "late")
			return nil
		},
	}
	r2 := &testBridgeRule{
		name:     "early",
		stage:    bridgeint.StageAfterHandle,
		priority: 5,
		run: func(context.Context, *bridgeint.Context) error {
			order = append(order, "early")
			return nil
		},
	}
	svc := newClientWithOptions(tr, time.Second, bridgeint.Options{
		Rules: []bridgeint.Rule{r1, r2},
	})

	err := svc.Handle(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		return []byte(`{"ok":true}`), nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if fmt.Sprint(order) != "[early late]" {
		t.Fatalf("unexpected run order: %v", order)
	}
}

func TestServiceHandleBeforePublishRulePublishesEvent(t *testing.T) {
	tr := &fakeServiceTransport{nextRequest: []byte(`{"name":"x"}`)}
	rule := bridgeint.NewHandlePublishRule(tr, bridgeint.HandlePublishRuleOptions{
		Name:     "before-publish",
		Stage:    bridgeint.StageBeforeHandle,
		Priority: 1,
		Subject:  "TW.XX.bridge.event.before",
		BuildPayload: func(bc *bridgeint.Context) (any, bool, error) {
			return map[string]any{"phase": "before", "source": bc.Subject}, true, nil
		},
	})
	svc := newClientWithOptions(tr, time.Second, bridgeint.Options{
		Rules: []bridgeint.Rule{rule},
	})

	err := svc.Handle(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		return []byte(`{"ok":true}`), nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(tr.publishCalls) != 1 {
		t.Fatalf("expected 1 publish call, got %d", len(tr.publishCalls))
	}
	if tr.publishCalls[0].subject != "TW.XX.bridge.event.before" {
		t.Fatalf("unexpected publish subject: %s", tr.publishCalls[0].subject)
	}
	var env envelope.EventEnvelope
	if err := json.Unmarshal(tr.publishCalls[0].data, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	payload, _ := env.Payload.(map[string]any)
	if payload["phase"] != "before" || payload["source"] != "TW.XX.user.command.create" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}

func TestServiceHandleAfterPublishRuleOnlyRunsOnHandlerSuccess(t *testing.T) {
	tr := &fakeServiceTransport{nextRequest: []byte(`{"name":"x"}`)}
	rule := bridgeint.NewHandlePublishRule(tr, bridgeint.HandlePublishRuleOptions{
		Name:     "after-publish",
		Stage:    bridgeint.StageAfterHandle,
		Priority: 1,
		Subject:  "TW.XX.bridge.event.after",
		BuildPayload: func(bc *bridgeint.Context) (any, bool, error) {
			return map[string]any{"phase": "after", "response": string(bc.Response)}, true, nil
		},
	})
	svc := newClientWithOptions(tr, time.Second, bridgeint.Options{
		Rules: []bridgeint.Rule{rule},
	})

	err := svc.Handle(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		return nil, errors.New("boom")
	})
	if err == nil {
		t.Fatal("expected handler error")
	}
	if len(tr.publishCalls) != 0 {
		t.Fatalf("expected no publish on handler failure, got %d", len(tr.publishCalls))
	}

	err = svc.Handle(context.Background(), "TW.XX.user.command.create", func(ctx context.Context, request []byte) ([]byte, error) {
		return []byte(`{"ok":true}`), nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(tr.publishCalls) != 1 {
		t.Fatalf("expected 1 publish call after success, got %d", len(tr.publishCalls))
	}
	if tr.publishCalls[0].subject != "TW.XX.bridge.event.after" {
		t.Fatalf("unexpected publish subject: %s", tr.publishCalls[0].subject)
	}
}

func TestServiceHandleAutoLifecycleRulesWithNewClient(t *testing.T) {
	tr := &fakeServiceTransport{nextRequest: []byte(`{"id":"r1"}`)}
	svc := NewClient(tr, time.Second)

	err := svc.Handle(context.Background(), "cmd.app.billing.invoice.create", func(ctx context.Context, request []byte) ([]byte, error) {
		return []byte(`{"ok":true}`), nil
	})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if len(tr.publishCalls) != 2 {
		t.Fatalf("expected 2 publish calls, got %d", len(tr.publishCalls))
	}
	if tr.publishCalls[0].subject != "run.app.billing.execution.started" {
		t.Fatalf("unexpected before lifecycle subject: %s", tr.publishCalls[0].subject)
	}
	if tr.publishCalls[1].subject != "evt.app.billing.invoice.create" {
		t.Fatalf("unexpected after lifecycle subject: %s", tr.publishCalls[1].subject)
	}
}

type fakeServiceTransport struct {
	subscribeSubject     string
	subscribeDurable     string
	nextDelivery         Delivery
	acked                bool
	handleRequestSubject string
	nextRequest          []byte
	lastResponse         []byte

	requestSubject string
	requestData    []byte
	requestResp    []byte
	publishCalls   []publishCall
}

func (f *fakeServiceTransport) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	f.requestSubject = subject
	f.requestData = data
	return f.requestResp, nil
}

func (f *fakeServiceTransport) Publish(ctx context.Context, subject string, data []byte) error {
	f.publishCalls = append(f.publishCalls, publishCall{subject: subject, data: data})
	return nil
}

func (f *fakeServiceTransport) Subscribe(ctx context.Context, subject string, durable string, handler func(context.Context, Delivery) error) error {
	f.subscribeSubject = subject
	f.subscribeDurable = durable
	return handler(ctx, f.nextDelivery)
}

func (f *fakeServiceTransport) HandleRequest(ctx context.Context, subject string, handler RequestHandler) error {
	f.handleRequestSubject = subject
	response, err := handler(ctx, f.nextRequest)
	if err != nil {
		return err
	}
	f.lastResponse = response
	return nil
}

type testBridgeRule struct {
	name     string
	stage    bridgeint.Stage
	priority int
	policy   *bridgeint.PolicyMode
	match    func(*bridgeint.Context) bool
	run      func(context.Context, *bridgeint.Context) error
}

func (r *testBridgeRule) Name() string { return r.name }

func (r *testBridgeRule) Stage() bridgeint.Stage { return r.stage }

func (r *testBridgeRule) Priority() int { return r.priority }

func (r *testBridgeRule) Match(bc *bridgeint.Context) bool {
	if r.match == nil {
		return true
	}
	return r.match(bc)
}

func (r *testBridgeRule) Run(ctx context.Context, bc *bridgeint.Context) error {
	if r.run == nil {
		return nil
	}
	return r.run(ctx, bc)
}

func (r *testBridgeRule) Policy() *bridgeint.PolicyMode { return r.policy }

var _ bridgeint.Rule = (*testBridgeRule)(nil)

type publishCall struct {
	subject string
	data    []byte
}
