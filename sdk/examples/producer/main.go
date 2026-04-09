package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/include2md/eventsdk/sdk/bootstrap"
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

	eventSubject := fmt.Sprintf("%s.user.event.registered", namespace)
	client, err := bootstrap.NewClient(bootstrap.Options{NATSURL: natsURL})
	if err != nil {
		log.Fatalf("new client: %v", err)
	}
	defer client.Close()

	if err := client.EnsureStream(envOr("SDK_STREAM", "SDK_EVENTS"), fmt.Sprintf("%s.user.event.*", namespace)); err != nil {
		log.Fatalf("ensure stream: %v", err)
	}

	event := UserRegistered{
		UserID:      "u-1",
		MessageID:   "m-1001",
		Title:       "Welcome",
		Description: "Welcome to EventSDK",
		Category:    "system",
		Box:         "primary",
		Email:       "u1@example.com",
	}

	if err := client.Emit(ctx, eventSubject, event); err != nil {
		log.Fatalf("publish subject: %v", err)
	}

	log.Printf("published subject=%s", eventSubject)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
