package examples_test

import (
	"context"

	"github.com/include2md/eventsdk/sdk"
)

func registerConsumer(service sdk.Service) {
	service.RegisterHandler("UserRegistered", func(ctx context.Context, payload []byte) error {
		return nil
	})

	_ = service.Run(context.Background(), sdk.RunConfig{
		Namespace:    "TW.XX",
		ConsumerName: "user-consumer",
	})
}
