package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	natslib "github.com/nats-io/nats.go"

	"github.com/include2md/eventsdk/sdk"
	transportnats "github.com/include2md/eventsdk/sdk/internal/transport/nats"
)

func main() {
	ctx := context.Background()
	natsURL := envOr("NATS_URL", "nats://127.0.0.1:4222")
	namespace := envOr("SDK_NAMESPACE", "TW.XX")
	eventSubject := fmt.Sprintf("%s.user.event.*", namespace)
	commandSubject := fmt.Sprintf("%s.user.command.create", namespace)

	nc, err := natslib.Connect(natsURL)
	if err != nil {
		log.Fatalf("connect nats: %v", err)
	}
	defer nc.Close()

	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("create jetstream: %v", err)
	}

	transport, err := transportnats.New(nc, js)
	if err != nil {
		log.Fatalf("create transport: %v", err)
	}

	service := sdk.NewService(transport)
	log.Printf("consumer started event_subject=%s command_subject=%s", eventSubject, commandSubject)

	err = service.Subscribe(ctx, eventSubject, envOr("SDK_CONSUMER_NAME", fmt.Sprintf("subject-consumer-%d", time.Now().UnixNano())), func(ctx context.Context, msg sdk.Message) error {
		log.Printf("received subject=%s event_id=%s correlation_id=%s payload=%s", msg.Subject, msg.EventID, msg.CorrelationID, string(msg.Payload))
		return nil
	})
	if err != nil {
		log.Fatalf("subscribe: %v", err)
	}

	err = service.HandleRequest(ctx, commandSubject, func(ctx context.Context, request []byte) ([]byte, error) {
		log.Printf("received command subject=%s payload=%s", commandSubject, string(request))
		return []byte(`{"ok":true,"message":"adapter processed"}`), nil
	})
	if err != nil {
		log.Fatalf("handle request: %v", err)
	}

	select {}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
