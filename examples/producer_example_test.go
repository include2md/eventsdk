package examples_test

import (
	"context"

	"github.com/include2md/eventsdk/sdk"
)

type MessageAdapter struct {
	client sdk.Client
}

type CreateMessageRequest struct {
	Title string
}

func (a *MessageAdapter) CreateMessage(ctx context.Context, req CreateMessageRequest) error {
	_, err := a.client.SendCommand(ctx, sdk.Command{
		Name:    "CreateMessage",
		Payload: req,
	})
	return err
}

type UserPublisher struct {
	client sdk.Client
}

type UserRegistered struct {
	UserID string
}

func (p *UserPublisher) PublishUserRegistered(ctx context.Context, event UserRegistered) error {
	return p.client.PublishEvent(ctx, sdk.Event{
		Type:    "UserRegistered",
		Payload: event,
	})
}
