package subject_test

import (
	"testing"

	"github.com/include2md/eventsdk/sdk/internal/subject"
)

func TestCommandSubjectMapping(t *testing.T) {
	r := subject.NewResolver("TW.XX")

	got, err := r.CommandSubject("CreateMessage")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != "TW.XX.inbox.command.create" {
		t.Fatalf("unexpected subject: %s", got)
	}
}

func TestEventSubject(t *testing.T) {
	r := subject.NewResolver("TW.XX")
	got := r.EventSubject("UserRegistered")
	if got != "TW.XX.sdk.event.UserRegistered" {
		t.Fatalf("unexpected event subject: %s", got)
	}
}

func TestCommandSubjectUnknown(t *testing.T) {
	r := subject.NewResolver("TW.XX")
	_, err := r.CommandSubject("Unknown")
	if err == nil {
		t.Fatal("expected error for unknown command")
	}
}
