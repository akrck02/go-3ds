package storage

import "errors"

// The codes of this package's errors, for programs to check with errors.Is.
var (
	ErrNotFound = errors.New("storage.not-found") // nothing was saved under that name yet
	ErrRead     = errors.New("storage.read-failed")
	ErrWrite    = errors.New("storage.write-failed")
)
