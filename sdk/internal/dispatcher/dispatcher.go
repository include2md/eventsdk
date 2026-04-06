package dispatcher

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/include2md/eventsdk/sdk/internal/envelope"
	"github.com/include2md/eventsdk/sdk/internal/registry"
	"github.com/include2md/eventsdk/sdk/internal/retry"
)

type RetryPublisher interface {
	Republish(ctx context.Context, env envelope.EventEnvelope) error
}

type Dispatcher struct {
	registry       registry.Registry
	retryPolicy    retry.Policy
	retryPublisher RetryPublisher
}

func New(reg registry.Registry, retryPolicy retry.Policy, retryPublisher RetryPublisher) *Dispatcher {
	return &Dispatcher{registry: reg, retryPolicy: retryPolicy, retryPublisher: retryPublisher}
}

func (d *Dispatcher) Handle(ctx context.Context, raw []byte) error {
	env, err := envelope.Unmarshal(raw)
	if err != nil {
		return fmt.Errorf("unmarshal envelope: %w", err)
	}

	h, ok := d.registry.Get(env.EventType)
	if !ok {
		return fmt.Errorf("no handler for event type: %s", env.EventType)
	}

	payload, err := json.Marshal(env.Payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	if err := h(ctx, payload); err != nil {
		if d.retryPolicy.CanRetry(env.Attempt) {
			env.Attempt = d.retryPolicy.NextAttempt(env.Attempt)
			if err := d.retryPublisher.Republish(ctx, env); err != nil {
				return fmt.Errorf("republish retry event: %w", err)
			}
			return nil
		}
		return fmt.Errorf("handler failed after max retry: %w", err)
	}

	return nil
}
