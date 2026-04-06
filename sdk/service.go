package sdk

import (
	"context"
	"fmt"

	"github.com/include2md/eventsdk/sdk/internal/dispatcher"
	"github.com/include2md/eventsdk/sdk/internal/envelope"
	"github.com/include2md/eventsdk/sdk/internal/registry"
	"github.com/include2md/eventsdk/sdk/internal/retry"
	"github.com/include2md/eventsdk/sdk/internal/subject"
)

type SDKService struct {
	transport Transport
	resolver  subject.Resolver
	registry  *registry.SafeRegistry
	retry     retry.Policy
}

func NewService(transport Transport, resolver subject.Resolver, maxRetries int) *SDKService {
	return &SDKService{
		transport: transport,
		resolver:  resolver,
		registry:  registry.New(),
		retry:     retry.NewPolicy(maxRetries),
	}
}

func (s *SDKService) RegisterHandler(eventType string, handler Handler) {
	s.registry.Register(eventType, registry.Handler(handler))
}

func (s *SDKService) Run(ctx context.Context, cfg RunConfig) error {
	d := dispatcher.New(s.registry, s.retry, s)
	subjectName := s.resolver.EventConsumeSubject()

	return s.transport.Subscribe(ctx, subjectName, cfg.ConsumerName, func(ctx context.Context, delivery Delivery) error {
		_ = d.Handle(ctx, delivery.Data)
		if delivery.Ack != nil {
			if err := delivery.Ack(); err != nil {
				return fmt.Errorf("ack message: %w", err)
			}
		}
		return nil
	})
}

func (s *SDKService) Republish(ctx context.Context, env envelope.EventEnvelope) error {
	payload, err := envelope.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal retry envelope: %w", err)
	}

	subjectName := s.resolver.EventSubject(env.EventType)
	if err := s.transport.Publish(ctx, subjectName, payload); err != nil {
		return fmt.Errorf("publish retry event: %w", err)
	}
	return nil
}
