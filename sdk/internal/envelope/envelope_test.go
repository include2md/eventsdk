package envelope_test

import (
	"testing"
	"time"

	"github.com/include2md/eventsdk/sdk/internal/envelope"
)

func TestNewEventEnvelopeSetsMetadata(t *testing.T) {
	e, err := envelope.NewEventEnvelope("UserRegistered", map[string]any{"id": "u1"}, "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if e.EventID == "" {
		t.Fatal("expected event id")
	}
	if e.CorrelationID == "" {
		t.Fatal("expected correlation id")
	}
	if time.Since(e.Timestamp) > time.Second {
		t.Fatal("expected recent timestamp")
	}
}

func TestNewEventEnvelopeKeepsCorrelationID(t *testing.T) {
	e, err := envelope.NewEventEnvelope("UserRegistered", nil, "c-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if e.CorrelationID != "c-1" {
		t.Fatalf("expected c-1, got %s", e.CorrelationID)
	}
}
