package network

import "errors"

// The codes of this package's errors, for programs to check with errors.Is.
var (
	ErrStart       = errors.New("network.start-failed")
	ErrHostUnknown = errors.New("network.host-unknown")
	ErrConnect     = errors.New("network.connect-failed")
	ErrSend        = errors.New("network.send-failed")
	ErrReceive     = errors.New("network.receive-failed")
	ErrClosed      = errors.New("network.closed")
	ErrSecure      = errors.New("network.secure-failed")
	ErrCertificate = errors.New("network.certificate-refused")
)
