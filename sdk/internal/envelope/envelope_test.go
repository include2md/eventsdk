package envelope_test

import (
	"encoding/json"
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
	if e.Source == "" {
		t.Fatal("expected source")
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

func TestMarshalProducesCloudEventShape(t *testing.T) {
	e, err := envelope.NewEventEnvelope("evt.app.demo.user.created", map[string]any{"id": "u1"}, "c-1")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	e.Source = "urn:connector:demo-consumer"
	e.AppID = "demo-app"
	e.Subject = "resource-1"
	e.CommandSubject = "cmd.app.demo.user.create"

	raw, err := envelope.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("json unmarshal: %v", err)
	}

	if got["specversion"] != "1.0" {
		t.Fatalf("unexpected specversion: %v", got["specversion"])
	}
	if got["type"] != "evt.app.demo.user.created" {
		t.Fatalf("unexpected type: %v", got["type"])
	}
	if got["source"] != "urn:connector:demo-consumer" {
		t.Fatalf("unexpected source: %v", got["source"])
	}
	if got["subject"] != "resource-1" {
		t.Fatalf("unexpected subject: %v", got["subject"])
	}
	if got["correlationid"] != "c-1" {
		t.Fatalf("unexpected correlationid: %v", got["correlationid"])
	}
	if got["appid"] != "demo-app" {
		t.Fatalf("unexpected appid: %v", got["appid"])
	}

	data, _ := got["data"].(map[string]any)
	request, _ := data["request"].(map[string]any)
	if request["id"] != "u1" {
		t.Fatalf("unexpected data.request: %#v", request)
	}
	if data["subject"] != "cmd.app.demo.user.create" {
		t.Fatalf("unexpected data.subject: %v", data["subject"])
	}
}

func TestUnmarshalCloudEventData(t *testing.T) {
	raw := []byte(`{
		"specversion":"1.0",
		"type":"evt.app.demo.user.created",
		"source":"urn:connector:demo-consumer",
		"subject":"resource-1",
		"id":"e-1",
		"time":"2026-04-27T00:00:00Z",
		"datacontenttype":"application/json",
		"data":{"request":{"id":"u1"},"subject":"cmd.app.demo.user.create"},
		"correlationid":"c-1",
		"appid":"demo-app"
	}`)

	env, err := envelope.Unmarshal(raw)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if env.EventType != "evt.app.demo.user.created" {
		t.Fatalf("unexpected type: %s", env.EventType)
	}
	if env.Source != "urn:connector:demo-consumer" {
		t.Fatalf("unexpected source: %s", env.Source)
	}
	if env.Subject != "resource-1" {
		t.Fatalf("unexpected subject: %s", env.Subject)
	}
	if env.EventID != "e-1" {
		t.Fatalf("unexpected id: %s", env.EventID)
	}
	if env.CorrelationID != "c-1" {
		t.Fatalf("unexpected correlation id: %s", env.CorrelationID)
	}
	if env.AppID != "demo-app" {
		t.Fatalf("unexpected appid: %s", env.AppID)
	}
	if env.CommandSubject != "cmd.app.demo.user.create" {
		t.Fatalf("unexpected command subject: %s", env.CommandSubject)
	}
	payload, _ := env.Payload.(map[string]any)
	if payload["id"] != "u1" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
}
