package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/include2md/eventsdk/sdk/internal/envelope"
	"github.com/include2md/eventsdk/sdk/internal/subject"
)

type SDKClient struct {
	transport Transport
	resolver  subject.Resolver
	timeout   time.Duration
}

func NewClient(transport Transport, resolver subject.Resolver, timeout time.Duration) *SDKClient {
	return &SDKClient{
		transport: transport,
		resolver:  resolver,
		timeout:   timeout,
	}
}

func (c *SDKClient) SendCommand(ctx context.Context, cmd Command) ([]byte, error) {
	subjectName, err := c.resolver.CommandSubject(cmd.Name)
	if err != nil {
		return nil, err
	}

	payload, err := json.Marshal(cmd.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshal command payload: %w", err)
	}

	return c.transport.Request(ctx, subjectName, payload, c.timeout)
}

func (c *SDKClient) PublishEvent(ctx context.Context, event Event) error {
	env, err := envelope.NewEventEnvelope(event.Type, event.Payload, event.CorrelationID)
	if err != nil {
		return err
	}

	payload, err := envelope.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal event envelope: %w", err)
	}

	subjectName := c.resolver.EventSubject(event.Type)
	if err := c.transport.Publish(ctx, subjectName, payload); err != nil {
		return err
	}

	return c.executeBridge(ctx, event, env.CorrelationID)
}

func (c *SDKClient) executeBridge(ctx context.Context, event Event, correlationID string) error {
	payload, ok := mapToInboxCreatePayload(event.Payload)
	if !ok {
		return nil
	}

	replyData, err := c.SendCommand(ctx, Command{
		Name:          "CreateMessage",
		Payload:       payload,
		CorrelationID: correlationID,
	})
	if err != nil {
		return nil
	}

	_ = validateBridgeReply(replyData)
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
