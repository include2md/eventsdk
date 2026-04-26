package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/include2md/eventsdk/sdk/internal/envelope"
)

type SDKClient struct {
	transport Transport
	timeout   time.Duration
}

func NewClient(transport Transport, timeout time.Duration) *SDKClient {
	return &SDKClient{transport: transport, timeout: timeout}
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
	return c.transport.HandleRequest(ctx, subject, func(ctx context.Context, request []byte) ([]byte, error) {
		response, err := handler(ctx, request)
		if err != nil {
			return nil, err
		}

		c.bridgeInboxFromRequest(ctx, request)
		return response, nil
	})
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

	return c.bridgeInboxFromPayload(ctx, payload, c.timeout)
}

func (c *SDKClient) bridgeInboxFromRequest(ctx context.Context, request []byte) {
	var payload any
	if err := json.Unmarshal(request, &payload); err != nil {
		return
	}

	_ = c.bridgeInboxFromPayload(ctx, payload, 3*time.Second)
}

func (c *SDKClient) bridgeInboxFromPayload(ctx context.Context, payload any, timeout time.Duration) error {
	mapped, ok := mapToInboxCreatePayload(payload)
	if !ok {
		return nil
	}

	reply, err := c.transport.Request(ctx, inboxCreateSubject, mustMarshal(mapped), timeout)
	if err != nil {
		return nil
	}
	_ = validateBridgeReply(reply)
	return nil
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}

func validateBridgeReply(replyData []byte) error {
	if len(replyData) == 0 {
		return nil
	}

	var reply struct {
		OK      *bool  `json:"ok"`
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(replyData, &reply); err != nil {
		return nil
	}
	if reply.OK != nil && !*reply.OK {
		return fmt.Errorf("bridge command rejected: code=%s message=%s", reply.Code, reply.Message)
	}
	return nil
}
