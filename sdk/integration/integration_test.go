//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	natslib "github.com/nats-io/nats.go"

	"github.com/include2md/eventsdk/sdk"
	"github.com/include2md/eventsdk/sdk/internal/subject"
	transportnats "github.com/include2md/eventsdk/sdk/internal/transport/nats"
)

func TestCommandFlowRequestReply(t *testing.T) {
	ns := envOr("SDK_NAMESPACE", "TW.XX")
	nc, js := mustConnect(t)
	defer nc.Close()

	tr, err := transportnats.New(nc, js)
	if err != nil {
		t.Fatalf("new transport: %v", err)
	}
	resolver := subject.NewResolver(ns)
	client := sdk.NewClient(tr, resolver, 3*time.Second)

	commandSubject, err := resolver.CommandSubject("CreateMessage")
	if err != nil {
		t.Fatalf("resolve command subject: %v", err)
	}

	// Local responder validates Core NATS request/reply path.
	sub, err := nc.Subscribe(commandSubject, func(msg *natslib.Msg) {
		_ = msg.Respond([]byte(`{"status":"ok"}`))
	})
	if err != nil {
		t.Fatalf("subscribe responder: %v", err)
	}
	defer sub.Unsubscribe()

	_, err = client.SendCommand(context.Background(), sdk.Command{
		Name: "CreateMessage",
		Payload: map[string]any{
			"title": "hello",
		},
	})
	if err != nil {
		t.Fatalf("send command: %v", err)
	}
}

func TestEventFlowPublishConsume(t *testing.T) {
	ns := envOr("SDK_NAMESPACE", "TW.XX")
	nc, js := mustConnect(t)
	defer nc.Close()

	tr, err := transportnats.New(nc, js)
	if err != nil {
		t.Fatalf("new transport: %v", err)
	}

	if err := ensureStream(js, envOr("SDK_STREAM", "SDK_EVENTS"), fmt.Sprintf("%s.sdk.event.*", ns)); err != nil {
		t.Fatalf("ensure stream: %v", err)
	}

	resolver := subject.NewResolver(ns)
	client := sdk.NewClient(tr, resolver, 3*time.Second)
	service := sdk.NewService(tr, resolver, 3)

	handled := make(chan struct{}, 1)
	service.RegisterHandler("UserRegistered", func(ctx context.Context, payload []byte) error {
		handled <- struct{}{}
		return nil
	})

	runErr := service.Run(context.Background(), sdk.RunConfig{
		Namespace:    ns,
		ConsumerName: fmt.Sprintf("sdk-consumer-%d", time.Now().UnixNano()),
	})
	if runErr != nil {
		t.Fatalf("service run: %v", runErr)
	}

	if err := client.PublishEvent(context.Background(), sdk.Event{
		Type: "UserRegistered",
		Payload: map[string]any{
			"user_id": "u-1",
		},
	}); err != nil {
		t.Fatalf("publish event: %v", err)
	}

	select {
	case <-handled:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for handler")
	}
}

func mustConnect(t *testing.T) (*natslib.Conn, natslib.JetStreamContext) {
	t.Helper()

	url := envOr("NATS_URL", "nats://127.0.0.1:4222")
	nc, err := natslib.Connect(url)
	if err != nil {
		t.Fatalf("connect nats (%s): %v", url, err)
	}

	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		t.Fatalf("jetstream: %v", err)
	}

	return nc, js
}

func ensureStream(js natslib.JetStreamContext, streamName, subject string) error {
	_, err := js.StreamInfo(streamName)
	if err == nil {
		return nil
	}

	_, err = js.AddStream(&natslib.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject},
	})
	return err
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
