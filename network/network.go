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

// openConnection connects to host:port and leaves the socket not blocking. It returns the socket,
// -1 when the host is unknown and -2 when the connection fails.
static int openConnection(const char *host, int port) {
	struct hostent *entry = gethostbyname(host);
	if (!entry || !entry->h_addr_list[0]) return -1;
	struct sockaddr_in address;
	memset(&address, 0, sizeof address);
	address.sin_family = AF_INET;
	address.sin_port = htons(port);
	memcpy(&address.sin_addr, entry->h_addr_list[0], sizeof address.sin_addr);
	int handle = socket(AF_INET, SOCK_STREAM, 0);
	if (handle < 0) return -2;
	if (connect(handle, (struct sockaddr *)&address, sizeof address) < 0) {
		close(handle);
		return -2;
	}
	int on = 1;
	setsockopt(handle, IPPROTO_TCP, TCP_NODELAY, &on, sizeof on);
	fcntl(handle, F_SETFL, fcntl(handle, F_GETFL, 0) | O_NONBLOCK);
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

// sendAll writes every byte, waiting while the system's buffer is full. 0 when done, -1 if not.
static int sendAll(int handle, const void *bytes, int size) {
	const char *next = bytes;
	while (size > 0) {
		int count = send(handle, next, size, 0);
		if (count < 0) {
			if (errno == EAGAIN || errno == EWOULDBLOCK) { svcSleepThread(1000000); continue; }
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

// Connection is an open TCP connection.
type Connection struct {
	handle C.int
	closed bool
}

// Dial connects to a host (a name or an address such as 192.168.1.20) on a port. It waits until
// the connection is made or refused.
func Dial(host string, port int) (*Connection, error) {
	if err := Start(); err != nil {
		return nil, err
	}
	name := C.CString(host)
	defer C.free(unsafe.Pointer(name))
	handle := C.openConnection(name, C.int(port))
	if handle == -1 {
		return nil, fmt.Errorf("%w: %s", ErrHostUnknown, host)
	}
	if handle < 0 {
		return nil, fmt.Errorf("%w: %s:%d", ErrConnect, host, port)
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
	count := C.receiveSome(connection.handle, unsafe.Pointer(&buffer[0]), C.int(len(buffer)))
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

// Write sends every byte of bytes, waiting while the system's buffer is full.
func (connection *Connection) Write(bytes []byte) (int, error) {
	if connection.closed {
		return 0, ErrClosed
	}
	if len(bytes) == 0 {
		return 0, nil
	}
	if C.sendAll(connection.handle, unsafe.Pointer(&bytes[0]), C.int(len(bytes))) != 0 {
		return 0, ErrSend
	}
	return len(bytes), nil
}

// Close closes the connection.
func (connection *Connection) Close() error {
	if !connection.closed {
		connection.closed = true
		C.close(connection.handle)
	}
	return nil
}
