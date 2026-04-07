package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	natslib "github.com/nats-io/nats.go"

	"github.com/include2md/eventsdk/sdk"
	"github.com/include2md/eventsdk/sdk/internal/subject"
	transportnats "github.com/include2md/eventsdk/sdk/internal/transport/nats"
)

type UserRegistered struct {
	UserID      string `json:"userId"`
	MessageID   string `json:"messageId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Box         string `json:"box"`
	Email       string `json:"email"`
}

type UserPublisher struct {
	client *sdk.SDKClient
}

func (p *UserPublisher) PublishUserRegistered(ctx context.Context, event UserRegistered) error {
	return p.client.PublishEvent(ctx, sdk.Event{
		Type:    "UserRegistered",
		Payload: event,
	})
}

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

	if err := ensureStream(js, envOr("SDK_STREAM", "SDK_EVENTS"), fmt.Sprintf("%s.sdk.event.*", namespace)); err != nil {
		log.Fatalf("ensure stream: %v", err)
	}

	transport, err := transportnats.New(nc, js)
	if err != nil {
		log.Fatalf("create transport: %v", err)
	}

	client := sdk.NewClient(transport, subject.NewResolver(namespace), 3*time.Second)

	publisher := &UserPublisher{client: client}
	if err := publisher.PublishUserRegistered(ctx, UserRegistered{
		UserID:      "u-1",
		MessageID:   "m-1001",
		Title:       "Welcome",
		Description: "Welcome to EventSDK",
		Category:    "system",
		Box:         "primary",
		Email:       "u1@example.com",
	}); err != nil {
		log.Fatalf("publish user registered: %v", err)
	}

	log.Println("published UserRegistered")
}

func ensureStream(js natslib.JetStreamContext, streamName, subjectPattern string) error {
	if _, err := js.StreamInfo(streamName); err == nil {
		return nil
	}

	_, err := js.AddStream(&natslib.StreamConfig{
		Name:     streamName,
		Subjects: []string{subjectPattern},
	})
	return err
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
