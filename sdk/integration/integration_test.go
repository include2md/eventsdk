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
	transportnats "github.com/include2md/eventsdk/sdk/internal/transport/nats"
)

func TestSubjectFlowRequestReply(t *testing.T) {
	ns := envOr("SDK_NAMESPACE", "TW.XX")
	nc, js := mustConnect(t)
	defer nc.Close()

	tr, err := transportnats.New(nc, js)
	if err != nil {
		t.Fatalf("new transport: %v", err)
	}
	client := sdk.NewClient(tr, 3*time.Second)
	service := sdk.NewClient(tr, 3*time.Second)

	subject := fmt.Sprintf("%s.user.command.create", ns)
	err = service.Handle(context.Background(), subject, func(ctx context.Context, request []byte) ([]byte, error) {
		return []byte(`{"ok":true,"message":"processed"}`), nil
	})
	if err != nil {
		t.Fatalf("handle request: %v", err)
	}

	reply, err := client.Request(context.Background(), subject, map[string]any{"title": "hello"})
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if string(reply) == "" {
		t.Fatal("expected reply")
	}
}

func TestSubjectFlowPublishSubscribe(t *testing.T) {
	ns := envOr("SDK_NAMESPACE", "TW.XX")
	nc, js := mustConnect(t)
	defer nc.Close()

	tr, err := transportnats.New(nc, js)
	if err != nil {
		t.Fatalf("new transport: %v", err)
	}

	subject := fmt.Sprintf("%s.user.event.created", ns)
	if err := ensureStream(js, envOr("SDK_STREAM", "SDK_EVENTS"), fmt.Sprintf("%s.user.event.*", ns)); err != nil {
		t.Fatalf("ensure stream: %v", err)
	}

	client := sdk.NewClient(tr, 3*time.Second)
	service := sdk.NewClient(tr, 3*time.Second)

	handled := make(chan struct{}, 1)
	err = service.Listen(context.Background(), fmt.Sprintf("%s.user.event.*", ns), fmt.Sprintf("sdk-consumer-%d", time.Now().UnixNano()), func(ctx context.Context, msg sdk.Message) error {
		handled <- struct{}{}
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	if err := client.Emit(context.Background(), subject, map[string]any{"user_id": "u-1"}); err != nil {
		t.Fatalf("publish: %v", err)
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
	info, err := js.StreamInfo(streamName)
	if err == nil {
		for _, s := range info.Config.Subjects {
			if s == subject {
				return nil
			}
		}
		cfg := info.Config
		cfg.Subjects = append(cfg.Subjects, subject)
		_, err = js.UpdateStream(&cfg)
		return err
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
