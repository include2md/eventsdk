package twsp

import "testing"

func TestNATSConnectOptions_None(t *testing.T) {
	got := natsConnectOptions(Options{})
	if got != nil {
		t.Fatalf("expected nil options, got len=%d", len(got))
	}
}

func TestNATSConnectOptions_UserInfo(t *testing.T) {
	got := natsConnectOptions(Options{
		Username: "user",
		Password: "pass",
	})
	if len(got) != 1 {
		t.Fatalf("expected one option, got len=%d", len(got))
	}
}

func TestNATSConnectOptions_TokenTakesPrecedence(t *testing.T) {
	got := natsConnectOptions(Options{
		Username: "user",
		Password: "pass",
		Token:    "token",
	})
	if len(got) != 1 {
		t.Fatalf("expected one option, got len=%d", len(got))
	}
}
