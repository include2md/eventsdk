package bridge

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/include2md/eventsdk/sdk/internal/envelope"
)

type Publisher interface {
	Publish(ctx context.Context, subject string, data []byte) error
}

type HandlePublishRuleOptions struct {
	Name         string
	Stage        Stage
	Priority     int
	Subject      string
	Policy       *PolicyMode
	BuildPayload func(*Context) (any, bool, error)
}

type handlePublishRule struct {
	publisher    Publisher
	name         string
	stage        Stage
	priority     int
	subject      string
	policy       *PolicyMode
	buildPayload func(*Context) (any, bool, error)
}

func NewHandlePublishRule(publisher Publisher, opts HandlePublishRuleOptions) Rule {
	return &handlePublishRule{
		publisher:    publisher,
		name:         opts.Name,
		stage:        opts.Stage,
		priority:     opts.Priority,
		subject:      opts.Subject,
		policy:       opts.Policy,
		buildPayload: opts.BuildPayload,
	}
}

func (r *handlePublishRule) Name() string { return r.name }

func (r *handlePublishRule) Stage() Stage { return r.stage }

func (r *handlePublishRule) Priority() int { return r.priority }

func (r *handlePublishRule) Match(bc *Context) bool {
	return bc != nil && bc.Stage == r.stage
}

func (r *handlePublishRule) Run(ctx context.Context, bc *Context) error {
	if r.publisher == nil {
		return fmt.Errorf("rule %q: publisher is nil", r.name)
	}
	if r.subject == "" {
		return fmt.Errorf("rule %q: subject is empty", r.name)
	}

	payload, ok, err := r.resolvePayload(bc)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	env, err := envelope.NewEventEnvelope(r.subject, payload, "")
	if err != nil {
		return fmt.Errorf("rule %q: new envelope: %w", r.name, err)
	}
	body, err := envelope.Marshal(env)
	if err != nil {
		return fmt.Errorf("rule %q: marshal envelope: %w", r.name, err)
	}
	if err := r.publisher.Publish(ctx, r.subject, body); err != nil {
		return fmt.Errorf("rule %q: publish: %w", r.name, err)
	}
	return nil
}

func (r *handlePublishRule) Policy() *PolicyMode { return r.policy }

func (r *handlePublishRule) resolvePayload(bc *Context) (any, bool, error) {
	if r.buildPayload != nil {
		return r.buildPayload(bc)
	}

	payload := map[string]any{
		"subject": bc.Subject,
		"request": rawJSONOrString(bc.Request),
	}
	if len(bc.Response) > 0 {
		payload["response"] = rawJSONOrString(bc.Response)
	}
	return payload, true, nil
}

func rawJSONOrString(b []byte) any {
	if len(b) == 0 {
		return ""
	}
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return string(b)
	}
	return v
}
