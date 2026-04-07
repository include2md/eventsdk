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

func (c *SDKClient) Publish(ctx context.Context, subject string, payload any) error {
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

	return c.executeInboxBridge(ctx, env.CorrelationID, payload)
}

func (c *SDKClient) executeInboxBridge(ctx context.Context, correlationID string, payload any) error {
	mapped, ok := mapToInboxCreatePayload(payload)
	if !ok {
		return nil
	}

	replyData, err := c.Request(ctx, "TW.XX.inbox.command.create", mapped)
	if err != nil {
		return nil
	}

	_ = validateBridgeReply(replyData)
	_ = correlationID
	return nil
}

func mapToInboxCreatePayload(payload any) (map[string]any, bool) {
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, false
	}

	var p struct {
		UserID      string `json:"userId"`
		MessageID   string `json:"messageId"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Category    string `json:"category"`
		Box         string `json:"box"`
	}

	if err := json.Unmarshal(b, &p); err != nil {
		return nil, false
	}

	if p.UserID == "" || p.MessageID == "" || p.Title == "" || p.Description == "" || p.Category == "" || p.Box == "" {
		return nil, false
	}

	return map[string]any{
		"userId":      p.UserID,
		"messageId":   p.MessageID,
		"title":       p.Title,
		"description": p.Description,
		"category":    p.Category,
		"box":         p.Box,
	}, true
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
