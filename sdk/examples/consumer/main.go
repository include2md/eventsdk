package main

import (
	"context"
	"log"
	"os"

	natslib "github.com/nats-io/nats.go"

	"github.com/include2md/eventsdk/sdk"
	"github.com/include2md/eventsdk/sdk/internal/subject"
	transportnats "github.com/include2md/eventsdk/sdk/internal/transport/nats"
)

func main() {
	ctx := context.Background()
	natsURL := envOr("NATS_URL", "nats://127.0.0.1:4222")
	namespace := envOr("SDK_NAMESPACE", "TW.XX")

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

	service := sdk.NewService(transport, subject.NewResolver(namespace), 3)
	service.RegisterHandler("UserRegistered", func(ctx context.Context, payload []byte) error {
		log.Printf("received UserRegistered payload=%s", string(payload))
		return nil
	})

	log.Println("consumer started")
	if err := service.Run(ctx, sdk.RunConfig{
		Namespace:    namespace,
		ConsumerName: envOr("SDK_CONSUMER_NAME", "user-registered-consumer"),
	}); err != nil {
		log.Fatalf("run service: %v", err)
	}

	// Keep process alive to continuously receive events.
	select {}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
