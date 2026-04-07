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

type UserRegistered struct {
	UserID      string `json:"userId"`
	MessageID   string `json:"messageId"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Box         string `json:"box"`
	Email       string `json:"email"`
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

	eventSubject := fmt.Sprintf("%s.user.event.registered", namespace)
	if err := ensureStream(js, envOr("SDK_STREAM", "SDK_EVENTS"), fmt.Sprintf("%s.user.event.*", namespace)); err != nil {
		log.Fatalf("ensure stream: %v", err)
	}

	transport, err := transportnats.New(nc, js)
	if err != nil {
		log.Fatalf("create transport: %v", err)
	}

	client := sdk.NewClient(transport, 3*time.Second)
	event := UserRegistered{
		UserID:      "u-1",
		MessageID:   "m-1001",
		Title:       "Welcome",
		Description: "Welcome to EventSDK",
		Category:    "system",
		Box:         "primary",
		Email:       "u1@example.com",
	}

	if err := client.Publish(ctx, eventSubject, event); err != nil {
		log.Fatalf("publish subject: %v", err)
	}

	log.Printf("published subject=%s", eventSubject)
}

func ensureStream(js natslib.JetStreamContext, streamName, subjectPattern string) error {
	info, err := js.StreamInfo(streamName)
	if err == nil {
		for _, s := range info.Config.Subjects {
			if s == subjectPattern {
				return nil
			}
		}
		cfg := info.Config
		cfg.Subjects = append(cfg.Subjects, subjectPattern)
		_, err = js.UpdateStream(&cfg)
		return err
	}

	_, err = js.AddStream(&natslib.StreamConfig{
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
