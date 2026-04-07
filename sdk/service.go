package sdk

import (
	"context"
	"encoding/json"
	"fmt"

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
	return s.transport.HandleRequest(ctx, subject, handler)
}
