package sdk_test

import (
	"testing"

	"github.com/include2md/eventsdk/sdk"
)

func TestPublicContracts(t *testing.T) {
	msg := sdk.Message{Subject: "TW.XX.user.created", Payload: []byte(`{}`)}
	if msg.Subject == "" {
		t.Fatal("message should expose subject")
	}
}
