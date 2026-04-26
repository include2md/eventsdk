package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	bridgeint "github.com/include2md/eventsdk/sdk/internal/bridge"
	"github.com/include2md/eventsdk/sdk/internal/envelope"
)

type SDKClient struct {
	transport Transport
	timeout   time.Duration
	hooks     *bridgeint.Hooks
}

func NewClient(transport Transport, timeout time.Duration) *SDKClient {
	return newClientWithOptions(transport, timeout, bridgeint.Options{})
}

func newClientWithOptions(transport Transport, timeout time.Duration, opts bridgeint.Options) *SDKClient {
	return &SDKClient{
		transport: transport,
		timeout:   timeout,
		hooks:     bridgeint.NewHooks(opts),
	}
}

func (c *SDKClient) Request(ctx context.Context, subject string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal request payload: %w", err)
	}
	return c.transport.Request(ctx, subject, body, c.timeout)
}

func (c *SDKClient) Listen(ctx context.Context, subject string, consumerName string, handler Handler) error {
	return c.transport.Subscribe(ctx, subject, consumerName, func(ctx context.Context, delivery Delivery) error {
		env, err := envelope.Unmarshal(delivery.Data)
		if err != nil {
			return fmt.Errorf("unmarshal envelope: %w", err)
		}

		payload, err := json.Marshal(env.Payload)
		if err != nil {
			return fmt.Errorf("marshal payload: %w", err)
		}

		if err := handler(ctx, Message{
			Subject:       delivery.Subject,
			EventID:       env.EventID,
			CorrelationID: env.CorrelationID,
			Timestamp:     env.Timestamp,
			Attempt:       env.Attempt,
			Payload:       payload,
		}); err != nil {
			return err
		}

		if delivery.Ack != nil {
			if err := delivery.Ack(); err != nil {
				return fmt.Errorf("ack message: %w", err)
			}
		}
		return nil
	})
}

func (c *SDKClient) Handle(ctx context.Context, subject string, handler RequestHandler) error {
	hooks := c.hooksForHandleSubject(subject)
	return c.transport.HandleRequest(ctx, subject, func(ctx context.Context, request []byte) ([]byte, error) {
		beforeCtx := &bridgeint.Context{
			Stage:   bridgeint.StageBeforeHandle,
			Subject: subject,
			Request: request,
		}
		if err := hooks.Apply(ctx, beforeCtx); err != nil {
			return nil, err
		}

		response, err := handler(ctx, request)
		if err != nil {
			return nil, err
		}

		afterCtx := &bridgeint.Context{
			Stage:    bridgeint.StageAfterHandle,
			Subject:  subject,
			Request:  request,
			Response: response,
		}
		_ = hooks.Apply(ctx, afterCtx)

		return response, nil
	})
}

func (c *SDKClient) hooksForHandleSubject(handleSubject string) *bridgeint.Hooks {
	if c.hooks == nil {
		return nil
	}

	lifecycleRules, err := bridgeint.BuildHandleLifecycleRules(c.transport, handleSubject)
	if err == nil {
		return c.hooks.WithExtraRules(lifecycleRules)
	}
	return c.hooks
}

func (c *SDKClient) Emit(ctx context.Context, subject string, payload any) error {
	env, err := envelope.NewEventEnvelope(subject, payload, "")
	if err != nil {
		return err
	}

	body, err := envelope.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}

	if err := c.transport.Publish(ctx, subject, body); err != nil {
		return err
	}

	return nil
}
