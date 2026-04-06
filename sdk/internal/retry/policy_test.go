package retry_test

import (
	"testing"

	"github.com/include2md/eventsdk/sdk/internal/retry"
)

func TestCanRetry(t *testing.T) {
	p := retry.NewPolicy(3)
	if !p.CanRetry(1) {
		t.Fatal("attempt 1 should retry")
	}
	if !p.CanRetry(2) {
		t.Fatal("attempt 2 should retry")
	}
	if p.CanRetry(3) {
		t.Fatal("attempt 3 should not retry")
	}
}

func TestNextAttempt(t *testing.T) {
	p := retry.NewPolicy(3)
	if got := p.NextAttempt(2); got != 3 {
		t.Fatalf("expected 3, got %d", got)
	}
}
