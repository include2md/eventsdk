package sdk_test

import (
	"context"
	"testing"

	"github.com/include2md/eventsdk/sdk"
)

func TestPublicContracts(t *testing.T) {
	var _ sdk.Client = (*fakeClient)(nil)
	var _ sdk.Service = (*fakeService)(nil)

	msg := sdk.Message{Subject: "TW.XX.user.created", Payload: []byte(`{}`)}
	if msg.Subject == "" {
		t.Fatal("message should expose subject")
	}
}

type fakeClient struct{}

func (f *fakeClient) Publish(ctx context.Context, subject string, payload any) error { return nil }
func (f *fakeClient) Request(ctx context.Context, subject string, payload any) ([]byte, error) {
	return nil, nil
}

type fakeService struct{}

func (f *fakeService) Subscribe(ctx context.Context, subject string, consumerName string, handler sdk.Handler) error {
	return nil
}
func (f *fakeService) HandleRequest(ctx context.Context, subject string, handler sdk.RequestHandler) error {
	return nil
}
