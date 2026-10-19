//go:build nintendo3ds

// Package network opens TCP connections on the Nintendo 3DS, through the system's socket service
// (soc:U, by way of libctru). Reads never wait: a program's loop asks once a frame for whatever
// has arrived, so a slow network never makes a frame late.
package network

/*
#include <3ds.h>
#include <malloc.h>
#include <string.h>
#include <errno.h>
#include <fcntl.h>
#include <unistd.h>
#include <netdb.h>
#include <arpa/inet.h>
#include <netinet/in.h>
#include <netinet/tcp.h>
#include <sys/socket.h>

// The socket service needs a buffer of its own, aligned to a page; this is libctru's usual size.
#define SOCKET_BUFFER_BYTES 0x100000

static u32 *socketBuffer;

static int startSockets(void) {
	if (socketBuffer) return 0;
	socketBuffer = memalign(0x1000, SOCKET_BUFFER_BYTES);
	if (!socketBuffer) return -1;
	if (R_FAILED(socInit(socketBuffer, SOCKET_BUFFER_BYTES))) {
		free(socketBuffer);
		socketBuffer = NULL;
		return -1;
	}
	return 0;
}

static void stopSockets(void) {
	if (!socketBuffer) return;
	socExit();
	free(socketBuffer);
	socketBuffer = NULL;
}

// openConnection connects to host:port within timeoutMs and leaves the socket not blocking. It
// returns the socket, -1 when the host is unknown and -2 when the connection fails or takes too
// long; then *stage says where (1 socket, 2 connect, 3 waiting, 4 refused) and *problem the
// system's error number.
static int openConnection(const char *host, int port, int timeoutMs, int *stage, int *problem) {
	struct hostent *entry = gethostbyname(host);
	if (!entry || !entry->h_addr_list[0]) return -1;
	struct sockaddr_in address;
	memset(&address, 0, sizeof address);
	address.sin_family = AF_INET;
	address.sin_port = htons(port);
	memcpy(&address.sin_addr, entry->h_addr_list[0], sizeof address.sin_addr);
	int handle = socket(AF_INET, SOCK_STREAM, 0);
	if (handle < 0) { *stage = 1; *problem = errno; return -2; }
	fcntl(handle, F_SETFL, fcntl(handle, F_GETFL, 0) | O_NONBLOCK);
	if (connect(handle, (struct sockaddr *)&address, sizeof address) < 0) {
		if (errno == EINPROGRESS || errno == EALREADY || errno == EWOULDBLOCK) {
			// The 3DS says the socket is ready before the connection is made, and its SO_ERROR
			// is the system's own number: ask connect again until it says connected.
			int waitedMs = 0;
			for (;;) {
				if (connect(handle, (struct sockaddr *)&address, sizeof address) == 0 || errno == EISCONN) break;
				if (errno != EINPROGRESS && errno != EALREADY && errno != EWOULDBLOCK) {
					*stage = 4; *problem = errno; close(handle); return -2;
				}
				if (waitedMs >= timeoutMs) { *stage = 3; *problem = ETIMEDOUT; close(handle); return -2; }
				svcSleepThread(10000000); // ten milliseconds
				waitedMs += 10;
			}
		} else {
			// Some systems refuse a connect that doesn't block: try again the plain way, which
			// waits as long as the system decides.
			int first = errno;
			close(handle);
			handle = socket(AF_INET, SOCK_STREAM, 0);
			if (handle < 0) { *stage = 1; *problem = errno; return -2; }
			if (connect(handle, (struct sockaddr *)&address, sizeof address) < 0) {
				*stage = 2; *problem = errno ? errno : first; close(handle); return -2;
			}
			fcntl(handle, F_SETFL, fcntl(handle, F_GETFL, 0) | O_NONBLOCK);
		}
	}
	int on = 1;
	setsockopt(handle, IPPROTO_TCP, TCP_NODELAY, &on, sizeof on);
	return handle;
}

// receiveSome reads what has arrived: the bytes read, 0 when nothing has, -1 when the other end
// closed and -2 on an error.
static int receiveSome(int handle, void *into, int size) {
	int count = recv(handle, into, size, 0);
	if (count > 0) return count;
	if (count == 0) return -1;
	if (errno == EAGAIN || errno == EWOULDBLOCK) return 0;
	return -2;
}

// sendSome sends what the system takes now, without waiting: the bytes sent, 0 when its buffer
// is full, or the negative of the system's error number.
static int sendSome(int handle, const void *bytes, int size) {
	int count = send(handle, bytes, size, 0);
	if (count >= 0) return count;
	if (errno == EAGAIN || errno == EWOULDBLOCK) return 0;
	return errno > 0 ? -errno : -1;
}
*/
import "C"

import (
	"fmt"
	"time"
	"unsafe"
)

// Start turns the socket service on. Dial calls it, so a program only calls it to know early
// whether the network is there; Stop turns the service off.
func Start() error {
	if C.startSockets() != 0 {
		return ErrStart
	}
	return nil
}

// Stop turns the socket service off. Connections still open stop working.
func Stop() { C.stopSockets() }

// How long Dial waits for the other end, in milliseconds.
const connectTimeoutMilliseconds = 5000

// How long bytes may wait to be sent, and how many, before the connection counts as stuck: a
// console's wifi can stall for a moment, which is not a reason to drop a game.
const (
	stuckSeconds        = 5
	mostPendingBytes    = 512 << 10
	flushAttemptsAFrame = 8 // sends tried each Flush while the system takes bytes
)

// connectStages name where a connection failed, for the error's details.
var connectStages = map[C.int]string{1: "opening a socket", 2: "connecting", 3: "waiting for an answer", 4: "refused"}

// Connection is an open TCP connection.
type Connection struct {
	handle      C.int
	closed      bool
	pending     []byte         // written, not yet taken by the system
	waitingFrom time.Time      // since when the oldest pending byte has waited
	failure     error          // the send that failed, reported by the next Write or Flush
	secure      unsafe.Pointer // set by DialSecure (secure.go): then reads and writes go through TLS
}

// Dial connects to a host (a name or an address such as 192.168.1.20) on a port. It waits until
// the connection is made or refused, five seconds at most.
func Dial(host string, port int) (*Connection, error) {
	if err := Start(); err != nil {
		return nil, err
	}
	name := C.CString(host)
	defer C.free(unsafe.Pointer(name))
	var stage, problem C.int
	handle := C.openConnection(name, C.int(port), connectTimeoutMilliseconds, &stage, &problem)
	if handle == -1 {
		return nil, fmt.Errorf("%w: %s", ErrHostUnknown, host)
	}
	if handle < 0 {
		return nil, fmt.Errorf("%w: %s:%d (%s, error %d)", ErrConnect, host, port, connectStages[stage], int(problem))
	}
	return &Connection{handle: handle}, nil
}

// Read copies into buffer what has arrived, and returns at once: 0 bytes and no error when
// nothing has. It returns ErrClosed once the other end has closed the connection.
func (connection *Connection) Read(buffer []byte) (int, error) {
	if connection.closed {
		return 0, ErrClosed
	}
	if len(buffer) == 0 {
		return 0, nil
	}
	var count C.int
	if connection.secure != nil {
		count = C.int(readSecure(connection.secure, buffer))
	} else {
		count = C.receiveSome(connection.handle, unsafe.Pointer(&buffer[0]), C.int(len(buffer)))
	}
	switch {
	case count >= 0:
		return int(count), nil
	case count == -1:
		connection.closed = true
		return 0, ErrClosed
	default:
		return 0, ErrReceive
	}
}

// Write queues bytes and sends what the system takes now; the rest goes in later calls of Write
// or Flush. It never waits. It fails with ErrSend when the system refused bytes, or when bytes
// have waited stuckSeconds or more than mostPendingBytes wait (the other end stopped reading).
func (connection *Connection) Write(bytes []byte) (int, error) {
	if connection.closed {
		return 0, ErrClosed
	}
	if len(connection.pending) == 0 {
		connection.waitingFrom = time.Now()
	}
	connection.pending = append(connection.pending, bytes...)
	if err := connection.Flush(); err != nil {
		return 0, err
	}
	return len(bytes), nil
}

// Flush sends what the system takes now of what Write queued; call it once a frame.
func (connection *Connection) Flush() error {
	if connection.failure != nil {
		return connection.failure
	}
	for attempt := 0; attempt < flushAttemptsAFrame && len(connection.pending) > 0; attempt++ {
		count := connection.sendSome(connection.pending)
		if count < 0 {
			connection.failure = fmt.Errorf("%w: %s", ErrSend, connection.sendProblem(count))
			return connection.failure
		}
		if count == 0 {
			break // the system's buffer is full: the rest waits for the next frame
		}
		connection.pending = connection.pending[:copy(connection.pending, connection.pending[count:])]
		connection.waitingFrom = time.Now()
	}
	if len(connection.pending) == 0 {
		return nil
	}
	if waited := time.Since(connection.waitingFrom); waited > stuckSeconds*time.Second || len(connection.pending) > mostPendingBytes {
		connection.failure = fmt.Errorf("%w: %d bytes waiting for %.1f s", ErrSend, len(connection.pending), waited.Seconds())
	}
	return connection.failure
}

// Pending is how many written bytes still wait to be sent: a game skips updates that a later one
// replaces while the network is behind.
func (connection *Connection) Pending() int { return len(connection.pending) }

// sendSome sends the start of bytes, plain or through TLS: the bytes sent, 0 when nothing can be
// sent now, or a negative error.
func (connection *Connection) sendSome(bytes []byte) int {
	if connection.secure != nil {
		return writeSecure(connection.secure, bytes)
	}
	return int(C.sendSome(connection.handle, unsafe.Pointer(&bytes[0]), C.int(len(bytes))))
}

// sendProblem names a failed send's error: the system's number, or mbedtls's.
func (connection *Connection) sendProblem(code int) string {
	if connection.secure != nil {
		return fmt.Sprintf("TLS error %#x", -code)
	}
	return fmt.Sprintf("system error %d", -code)
}

// Close closes the connection, also after the other end has closed it: the socket and, for a
// secure connection, TLS's memory are let go either way.
func (connection *Connection) Close() error {
	connection.closed = true
	if connection.secure != nil {
		closeSecure(connection.secure)
		connection.secure = nil
	}
	if connection.handle >= 0 {
		C.close(connection.handle)
		connection.handle = -1
	}
	return nil
}
