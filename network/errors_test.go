package network

import (
	"errors"
	"testing"
)

func TestEveryErrorHasItsOwnCode(t *testing.T) {
	codes := map[error]string{
		ErrStart: "network.start-failed", ErrHostUnknown: "network.host-unknown",
		ErrConnect: "network.connect-failed", ErrSend: "network.send-failed",
		ErrReceive: "network.receive-failed", ErrClosed: "network.closed",
	}
	for err, code := range codes {
		if err.Error() != code {
			t.Errorf("%v has the code %q, want %q", err, err.Error(), code)
		}
		if !errors.Is(err, err) {
			t.Errorf("%v is not itself", err)
		}
	}
}
