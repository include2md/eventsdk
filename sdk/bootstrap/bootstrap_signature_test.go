package twsp

import "testing"

func TestNewClientSupportsOptionalOptionsArgument(t *testing.T) {
	_ = func() {
		_, _ = NewClient()
		_, _ = NewClient(Options{NATSURL: "nats://127.0.0.1:4222"})
	}
}
