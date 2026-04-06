package sdk_test

import (
	"context"
	"testing"

	"github.com/include2md/eventsdk/sdk"
)

func TestPublicContracts(t *testing.T) {
	var _ sdk.Client = (*fakeClient)(nil)
	var _ sdk.Service = (*fakeService)(nil)

	_ = sdk.Command{Name: "CreateMessage", Payload: map[string]any{"x": 1}}
	_ = sdk.Event{Type: "UserRegistered", Payload: map[string]any{"id": "u1"}}

	cfg := sdk.RunConfig{ConsumerName: "consumer", Namespace: "TW.XX"}
	if cfg.ConsumerName == "" || cfg.Namespace == "" {
		t.Fatal("run config should expose required fields")
	}
}

type fakeClient struct{}

func (f *fakeClient) SendCommand(ctx context.Context, cmd sdk.Command) ([]byte, error) {
	return nil, nil
}
func (f *fakeClient) PublishEvent(ctx context.Context, event sdk.Event) error { return nil }

type fakeService struct{}

func (f *fakeService) RegisterHandler(eventType string, handler sdk.Handler) {}
func (f *fakeService) Run(ctx context.Context, cfg sdk.RunConfig) error      { return nil }
