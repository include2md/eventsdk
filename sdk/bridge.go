package sdk

import "context"

type BridgeMode string

const (
	BridgeModeDefault BridgeMode = "default"
	BridgeModeStrict  BridgeMode = "strict"
)

type BridgeRule struct {
	EventType   string
	CommandName string
	MapPayload  func(event Event) (any, error)
}

type BridgeObserver interface {
	OnBridgeSuccess(ctx context.Context, eventType string, correlationID string, commandName string)
	OnBridgeFailure(ctx context.Context, eventType string, correlationID string, commandName string, err error)
}

type noopBridgeObserver struct{}

func (noopBridgeObserver) OnBridgeSuccess(context.Context, string, string, string)        {}
func (noopBridgeObserver) OnBridgeFailure(context.Context, string, string, string, error) {}
