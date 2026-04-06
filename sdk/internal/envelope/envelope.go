package envelope

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type EventEnvelope struct {
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	Timestamp     time.Time `json:"timestamp"`
	CorrelationID string    `json:"correlation_id"`
	Attempt       int       `json:"attempt"`
	Payload       any       `json:"payload"`
}

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
		Timestamp:     time.Now().UTC(),
		CorrelationID: corrID,
		Attempt:       1,
		Payload:       payload,
	}, nil
}

func Marshal(e EventEnvelope) ([]byte, error) {
	return json.Marshal(e)
}

func Unmarshal(data []byte) (EventEnvelope, error) {
	var env EventEnvelope
	if err := json.Unmarshal(data, &env); err != nil {
		return EventEnvelope{}, err
	}
	return env, nil
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate random id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
