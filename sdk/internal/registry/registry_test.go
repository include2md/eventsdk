package registry_test

import (
	"context"
	"testing"

	"github.com/include2md/eventsdk/sdk/internal/registry"
)

func TestRegistryRegisterAndGet(t *testing.T) {
	r := registry.New()
	h := func(context.Context, []byte) error { return nil }

	r.Register("UserRegistered", h)

	got, ok := r.Get("UserRegistered")
	if !ok {
		t.Fatal("expected handler")
	}
	if got == nil {
		t.Fatal("handler should not be nil")
	}
}

func TestRegistryGetUnknown(t *testing.T) {
	r := registry.New()

	got, ok := r.Get("Unknown")
	if ok {
		t.Fatal("expected no handler")
	}
	if got != nil {
		t.Fatal("expected nil handler")
	}
}

var _ registry.Registry = (*registry.SafeRegistry)(nil)
