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
	return &SDKClient{transport: transport, resolver: resolver, timeout: timeout}
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
	return c.transport.Publish(ctx, subjectName, payload)
}
