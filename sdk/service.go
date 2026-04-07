package sdk

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/include2md/eventsdk/sdk/internal/envelope"
)

type SDKService struct {
	transport Transport
}

func NewService(transport Transport) *SDKService {
	return &SDKService{transport: transport}
}

func (s *SDKService) Subscribe(ctx context.Context, subject string, consumerName string, handler Handler) error {
	return s.transport.Subscribe(ctx, subject, consumerName, func(ctx context.Context, delivery Delivery) error {
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

func (s *SDKService) HandleRequest(ctx context.Context, subject string, handler RequestHandler) error {
	return s.transport.HandleRequest(ctx, subject, func(ctx context.Context, request []byte) ([]byte, error) {
		response, err := handler(ctx, request)
		if err != nil {
			return nil, err
		}

		s.tryInboxBridge(ctx, request)
		return response, nil
	})
}

func (s *SDKService) tryInboxBridge(ctx context.Context, request []byte) {
	var payload any
	if err := json.Unmarshal(request, &payload); err != nil {
		return
	}

	mapped, ok := mapToInboxCreatePayload(payload)
	if !ok {
		return
	}

	reply, err := s.transport.Request(ctx, inboxCreateSubject, mustMarshal(mapped), 3*time.Second)
	if err != nil {
		return
	}
	_ = validateBridgeReply(reply)
}

func mustMarshal(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}
