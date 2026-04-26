package envelope

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/event"
)

type EventEnvelope struct {
	EventID        string    `json:"event_id"`
	EventType      string    `json:"event_type"`
	Source         string    `json:"source"`
	Subject        string    `json:"subject,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
	CorrelationID  string    `json:"correlation_id"`
	AppID          string    `json:"app_id,omitempty"`
	Attempt        int       `json:"attempt"`
	CommandSubject string    `json:"command_subject,omitempty"`
	Payload        any       `json:"payload"`
}

const defaultSource = "urn:connector:unknown"

func NewEventEnvelope(eventType string, payload any, correlationID string) (EventEnvelope, error) {
	eventID, err := randomID()
	if err != nil {
		return EventEnvelope{}, err
	}

	corrID := correlationID
	if corrID == "" {
		corrID, err = randomID()
		if err != nil {
			return EventEnvelope{}, err
		}
	}

	return EventEnvelope{
		EventID:       eventID,
		EventType:     eventType,
		Source:        defaultSource,
		Timestamp:     time.Now().UTC(),
		CorrelationID: corrID,
		Attempt:       1,
		Payload:       payload,
	}, nil
}

func Marshal(e EventEnvelope) ([]byte, error) {
	ce := cloudevents.NewEvent()
	ce.SetType(e.EventType)
	ce.SetSource(valueOrDefault(e.Source, defaultSource))
	if e.Subject != "" {
		ce.SetSubject(e.Subject)
	}
	if e.EventID != "" {
		ce.SetID(e.EventID)
	}
	if !e.Timestamp.IsZero() {
		ce.SetTime(e.Timestamp.UTC())
	}

	data := map[string]any{
		"request": e.Payload,
	}
	if e.CommandSubject != "" {
		data["subject"] = e.CommandSubject
	}
	if err := ce.SetData(cloudevents.ApplicationJSON, data); err != nil {
		return nil, fmt.Errorf("set cloud event data: %w", err)
	}

	if e.CorrelationID != "" {
		if err := ce.Context.SetExtension("correlationid", e.CorrelationID); err != nil {
			return nil, fmt.Errorf("set extension correlationid: %w", err)
		}
	}
	if e.AppID != "" {
		if err := ce.Context.SetExtension("appid", e.AppID); err != nil {
			return nil, fmt.Errorf("set extension appid: %w", err)
		}
	}
	if e.Attempt > 0 {
		if err := ce.Context.SetExtension("attempt", e.Attempt); err != nil {
			return nil, fmt.Errorf("set extension attempt: %w", err)
		}
	}
	if e.EventType != "" {
		if err := ce.Context.SetExtension("natssubject", e.EventType); err != nil {
			return nil, fmt.Errorf("set extension natssubject: %w", err)
		}
	}

	raw, err := json.Marshal(ce)
	if err != nil {
		return nil, fmt.Errorf("marshal cloud event: %w", err)
	}
	return raw, nil
}

func Unmarshal(data []byte) (EventEnvelope, error) {
	var ce event.Event
	if err := json.Unmarshal(data, &ce); err != nil {
		return EventEnvelope{}, err
	}

	env := EventEnvelope{
		EventID:       ce.ID(),
		EventType:     ce.Type(),
		Source:        ce.Source(),
		Subject:       ce.Subject(),
		Timestamp:     ce.Time(),
		CorrelationID: extensionString(ce, "correlationid"),
		AppID:         extensionString(ce, "appid"),
		Attempt:       extensionInt(ce, "attempt", 1),
	}

	if len(ce.Data()) == 0 {
		return env, nil
	}

	var wrapped struct {
		Request json.RawMessage `json:"request"`
		Subject string          `json:"subject"`
	}
	if err := json.Unmarshal(ce.Data(), &wrapped); err == nil && len(wrapped.Request) > 0 {
		env.CommandSubject = wrapped.Subject
		var payload any
		if err := json.Unmarshal(wrapped.Request, &payload); err != nil {
			return EventEnvelope{}, fmt.Errorf("unmarshal data.request: %w", err)
		}
		env.Payload = payload
		return env, nil
	}

	var payload any
	if err := json.Unmarshal(ce.Data(), &payload); err != nil {
		return EventEnvelope{}, fmt.Errorf("unmarshal data: %w", err)
	}
	env.Payload = payload
	return env, nil
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random id: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func valueOrDefault(v string, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func extensionString(e event.Event, key string) string {
	raw := e.Extensions()[key]
	switch v := raw.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func extensionInt(e event.Event, key string, fallback int) int {
	raw := e.Extensions()[key]
	switch v := raw.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	default:
		return fallback
	}
}
