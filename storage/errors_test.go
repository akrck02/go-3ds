package storage

import "testing"

func TestEveryErrorHasItsOwnCode(t *testing.T) {
	codes := map[error]string{ErrNotFound: "storage.not-found", ErrRead: "storage.read-failed", ErrWrite: "storage.write-failed"}
	for err, code := range codes {
		if err.Error() != code {
			t.Errorf("%v has the code %q, want %q", err, err.Error(), code)
		}
	}
}
