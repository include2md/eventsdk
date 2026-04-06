package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/include2md/eventsdk/sdk/internal/envelope"
	"github.com/include2md/eventsdk/sdk/internal/subject"
)

type SDKClient struct {
	transport Transport
	resolver  subject.Resolver
	timeout   time.Duration

	bridgeMode     BridgeMode
	bridgeObserver BridgeObserver
	bridgeMu       sync.RWMutex
	bridgeRules    map[string]BridgeRule
}

func NewClient(transport Transport, resolver subject.Resolver, timeout time.Duration) *SDKClient {
	return &SDKClient{
		transport:      transport,
		resolver:       resolver,
		timeout:        timeout,
		bridgeMode:     BridgeModeDefault,
		bridgeObserver: noopBridgeObserver{},
		bridgeRules:    map[string]BridgeRule{},
	}
}

func (c *SDKClient) RegisterBridgeRule(rule BridgeRule) {
	c.bridgeMu.Lock()
	defer c.bridgeMu.Unlock()
	c.bridgeRules[rule.EventType] = rule
}

func (c *SDKClient) SetBridgeMode(mode BridgeMode) {
	if mode == "" {
		mode = BridgeModeDefault
	}
	c.bridgeMode = mode
}

func (c *SDKClient) SetBridgeObserver(observer BridgeObserver) {
	if observer == nil {
		observer = noopBridgeObserver{}
	}
	c.bridgeObserver = observer
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
	rule, ok := c.lookupBridgeRule(event.Type)
	if !ok {
		return nil
	}

	mappedPayload := event.Payload
	if rule.MapPayload != nil {
		var err error
		mappedPayload, err = rule.MapPayload(event)
		if err != nil {
			c.bridgeObserver.OnBridgeFailure(ctx, event.Type, correlationID, rule.CommandName, err)
			if c.bridgeMode == BridgeModeStrict {
				return fmt.Errorf("bridge map payload: %w", err)
			}
			return nil
		}
	}

	if err := validateBridgePayload(rule.CommandName, mappedPayload); err != nil {
		c.bridgeObserver.OnBridgeFailure(ctx, event.Type, correlationID, rule.CommandName, err)
		if c.bridgeMode == BridgeModeStrict {
			return fmt.Errorf("bridge payload validation: %w", err)
		}
		return nil
	}

	replyData, err := c.SendCommand(ctx, Command{
		Name:          rule.CommandName,
		Payload:       mappedPayload,
		CorrelationID: correlationID,
	})
	if err != nil {
		c.bridgeObserver.OnBridgeFailure(ctx, event.Type, correlationID, rule.CommandName, err)
		if c.bridgeMode == BridgeModeStrict {
			return fmt.Errorf("bridge command send: %w", err)
		}
		return nil
	}

	if err := validateBridgeReply(replyData); err != nil {
		c.bridgeObserver.OnBridgeFailure(ctx, event.Type, correlationID, rule.CommandName, err)
		if c.bridgeMode == BridgeModeStrict {
			return fmt.Errorf("bridge reply validation: %w", err)
		}
		return nil
	}

	c.bridgeObserver.OnBridgeSuccess(ctx, event.Type, correlationID, rule.CommandName)
	return nil
}

func (c *SDKClient) lookupBridgeRule(eventType string) (BridgeRule, bool) {
	c.bridgeMu.RLock()
	defer c.bridgeMu.RUnlock()
	rule, ok := c.bridgeRules[eventType]
	return rule, ok
}

func validateBridgePayload(commandName string, payload any) error {
	if commandName != "CreateMessage" {
		return nil
	}

	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload for validation: %w", err)
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
		return fmt.Errorf("unmarshal payload for validation: %w", err)
	}

	switch {
	case p.UserID == "":
		return fmt.Errorf("missing required field: userId")
	case p.MessageID == "":
		return fmt.Errorf("missing required field: messageId")
	case p.Title == "":
		return fmt.Errorf("missing required field: title")
	case p.Description == "":
		return fmt.Errorf("missing required field: description")
	case p.Category == "":
		return fmt.Errorf("missing required field: category")
	case p.Box == "":
		return fmt.Errorf("missing required field: box")
	}
	return nil
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
