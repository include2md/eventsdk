package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
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
	UserID string `json:"user_id"`
	Email  string `json:"email"`
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
	client.SetBridgeObserver(logBridgeObserver{})

	// Internal bridge: when UserRegistered is published, SDK additionally issues inbox CreateMessage command.
	client.RegisterBridgeRule(sdk.BridgeRule{
		EventType:   "UserRegistered",
		CommandName: "CreateMessage",
		MapPayload: func(event sdk.Event) (any, error) {
			evt, _ := event.Payload.(UserRegistered)
			msgID, err := newMessageID()
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"userId":      evt.UserID,
				"messageId":   msgID,
				"title":       "Welcome",
				"description": "Welcome to EventSDK",
				"category":    "system",
				"box":         "primary",
			}, nil
		},
	})

	publisher := &UserPublisher{client: client}
	if err := publisher.PublishUserRegistered(ctx, UserRegistered{UserID: "u-1", Email: "u1@example.com"}); err != nil {
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

func newMessageID() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "msg_" + hex.EncodeToString(b), nil
}

type logBridgeObserver struct{}

func (logBridgeObserver) OnBridgeSuccess(ctx context.Context, eventType string, correlationID string, commandName string) {
	log.Printf("bridge success eventType=%s correlationID=%s command=%s", eventType, correlationID, commandName)
}

func (logBridgeObserver) OnBridgeFailure(ctx context.Context, eventType string, correlationID string, commandName string, err error) {
	log.Printf("bridge failure eventType=%s correlationID=%s command=%s err=%v", eventType, correlationID, commandName, err)
}
