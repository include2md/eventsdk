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

func TestResolveCloudEventSource(t *testing.T) {
	if got := resolveCloudEventSource(Options{}); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}

	if got := resolveCloudEventSource(Options{ConnectorID: "consumer-a"}); got != "urn:connector:consumer-a" {
		t.Fatalf("unexpected source: %q", got)
	}

	if got := resolveCloudEventSource(Options{
		ConnectorID:      "consumer-a",
		CloudEventSource: "urn:connector:explicit",
	}); got != "urn:connector:explicit" {
		t.Fatalf("unexpected source: %q", got)
	}
}
