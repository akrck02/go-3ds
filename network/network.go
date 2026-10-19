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

// sendAll writes every byte, waiting at most timeoutMs in all while the system's buffer is full,
// so a peer that stopped reading can't freeze the program. 0 when done, -1 if not.
static int sendAll(int handle, const void *bytes, int size, int timeoutMs) {
	const char *next = bytes;
	int waitedMs = 0;
	while (size > 0) {
		int count = send(handle, next, size, 0);
		if (count < 0) {
			if ((errno == EAGAIN || errno == EWOULDBLOCK) && waitedMs < timeoutMs) {
				svcSleepThread(1000000); // a millisecond
				waitedMs++;
				continue;
			}
			return -1;
		}
		next += count;
		size -= count;
	}
	return 0;
}
*/
import "C"

import (
	"fmt"
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

// How long Dial waits for the other end, and Write for room to send, in milliseconds. A frame is
// about 17 ms: a send that waits longer than this gives up instead of freezing the program.
const (
	connectTimeoutMilliseconds = 5000
	sendTimeoutMilliseconds    = 200
)

// connectStages name where a connection failed, for the error's details.
var connectStages = map[C.int]string{1: "opening a socket", 2: "connecting", 3: "waiting for an answer", 4: "refused"}

// Connection is an open TCP connection.
type Connection struct {
	handle C.int
	closed bool
	secure unsafe.Pointer // set by DialSecure (secure.go): then reads and writes go through TLS
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

// Write sends every byte of bytes, waiting while the system's buffer is full, and fails with
// ErrSend if it stays full for a fifth of a second.
func (connection *Connection) Write(bytes []byte) (int, error) {
	if connection.closed {
		return 0, ErrClosed
	}
	if len(bytes) == 0 {
		return 0, nil
	}
	if connection.secure != nil {
		if !writeSecure(connection.secure, bytes) {
			return 0, ErrSend
		}
		return len(bytes), nil
	}
	if C.sendAll(connection.handle, unsafe.Pointer(&bytes[0]), C.int(len(bytes)), sendTimeoutMilliseconds) != 0 {
		return 0, ErrSend
	}
	return len(bytes), nil
}

// Close closes the connection.
func (connection *Connection) Close() error {
	if !connection.closed {
		connection.closed = true
		if connection.secure != nil {
			closeSecure(connection.secure)
			connection.secure = nil
		}
		C.close(connection.handle)
	}
	return nil
}
