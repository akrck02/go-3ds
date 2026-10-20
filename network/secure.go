//go:build nintendo3ds

package network

/*
// devkitPro's portlibs build of mbedtls (2.28): the path is where devkitPro installs it, as the
// nintendo3ds target expects devkitPro there too.
#cgo CFLAGS: -I/opt/devkitpro/portlibs/3ds/include
#cgo LDFLAGS: -L/opt/devkitpro/portlibs/3ds/lib -lmbedtls -lmbedx509 -lmbedcrypto

#include <3ds.h>
#include <errno.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/socket.h>
#include <mbedtls/ssl.h>
#include <mbedtls/ctr_drbg.h>
#include <mbedtls/x509_crt.h>
#include <mbedtls/error.h>
#include <mbedtls/net_sockets.h>

// secureConnection is everything mbedtls keeps for one connection, in C's memory: mbedtls points
// into it, and Go's collector must not move or free it.
typedef struct {
	mbedtls_ssl_context ssl;
	mbedtls_ssl_config config;
	mbedtls_x509_crt roots;
	mbedtls_ctr_drbg_context random;
	int handle;
} secureConnection;

// randomBytes is mbedtls's source of entropy: the console's own generator (the ps service).
static int randomBytes(void *unused, unsigned char *into, size_t size) {
	return R_FAILED(PS_GenerateRandomBytes(into, size)) ? MBEDTLS_ERR_CTR_DRBG_ENTROPY_SOURCE_FAILED : 0;
}

// sendSome and receiveSome put mbedtls on the socket, which doesn't block: "would block" becomes
// mbedtls's "want write" or "want read", so the caller tries again later.
static int sendSome(void *context, const unsigned char *bytes, size_t size) {
	int count = send(((secureConnection *)context)->handle, bytes, size, 0);
	if (count >= 0) return count;
	if (errno == EAGAIN || errno == EWOULDBLOCK) return MBEDTLS_ERR_SSL_WANT_WRITE;
	return MBEDTLS_ERR_NET_SEND_FAILED;
}

static int receiveSome(void *context, unsigned char *into, size_t size) {
	int count = recv(((secureConnection *)context)->handle, into, size, 0);
	if (count >= 0) return count; // 0: the other end closed
	if (errno == EAGAIN || errno == EWOULDBLOCK) return MBEDTLS_ERR_SSL_WANT_READ;
	return MBEDTLS_ERR_NET_RECV_FAILED;
}

static void freeSecure(secureConnection *secure) {
	mbedtls_ssl_free(&secure->ssl);
	mbedtls_ssl_config_free(&secure->config);
	mbedtls_x509_crt_free(&secure->roots);
	mbedtls_ctr_drbg_free(&secure->random);
	free(secure);
}

// openSecure runs TLS's handshake on a connected socket, checking the server's certificate
// against roots (PEM, ending in a NUL) and its name against host, within timeoutMs. It returns the
// connection, or NULL with *problem set to mbedtls's error, or to the certificate's verification
// flags (*flags) when the certificate was refused.
static secureConnection *openSecure(int handle, const char *host, const char *roots, size_t rootsSize,
		int timeoutMs, int *problem, unsigned int *flags) {
	*problem = 0;
	*flags = 0;
	if (R_FAILED(psInit())) { *problem = -1; return NULL; }
	secureConnection *secure = calloc(1, sizeof *secure);
	if (!secure) { *problem = -1; return NULL; }
	secure->handle = handle;
	mbedtls_ssl_init(&secure->ssl);
	mbedtls_ssl_config_init(&secure->config);
	mbedtls_x509_crt_init(&secure->roots);
	mbedtls_ctr_drbg_init(&secure->random);
	int result = mbedtls_ctr_drbg_seed(&secure->random, randomBytes, NULL, (const unsigned char *)"go-3ds", 6);
	if (result == 0) result = mbedtls_x509_crt_parse(&secure->roots, (const unsigned char *)roots, rootsSize);
	if (result == 0) result = mbedtls_ssl_config_defaults(&secure->config, MBEDTLS_SSL_IS_CLIENT,
		MBEDTLS_SSL_TRANSPORT_STREAM, MBEDTLS_SSL_PRESET_DEFAULT);
	if (result == 0) {
		mbedtls_ssl_conf_authmode(&secure->config, MBEDTLS_SSL_VERIFY_REQUIRED);
		mbedtls_ssl_conf_ca_chain(&secure->config, &secure->roots, NULL);
		mbedtls_ssl_conf_rng(&secure->config, mbedtls_ctr_drbg_random, &secure->random);
		result = mbedtls_ssl_setup(&secure->ssl, &secure->config);
	}
	if (result == 0) result = mbedtls_ssl_set_hostname(&secure->ssl, host);
	if (result != 0) { *problem = result; freeSecure(secure); return NULL; }
	mbedtls_ssl_set_bio(&secure->ssl, secure, sendSome, receiveSome, NULL);
	int waitedMs = 0;
	while ((result = mbedtls_ssl_handshake(&secure->ssl)) != 0) {
		if (result != MBEDTLS_ERR_SSL_WANT_READ && result != MBEDTLS_ERR_SSL_WANT_WRITE) break;
		if (waitedMs >= timeoutMs) { result = MBEDTLS_ERR_SSL_TIMEOUT; break; }
		svcSleepThread(1000000); // a millisecond
		waitedMs++;
	}
	if (result != 0) {
		*problem = result;
		*flags = mbedtls_ssl_get_verify_result(&secure->ssl);
		freeSecure(secure);
		return NULL;
	}
	return secure;
}

// readSecure reads what has arrived: the bytes read, 0 when nothing has, -1 when the other end
// closed and -2 on an error.
static int readSecure(secureConnection *secure, unsigned char *into, int size) {
	int count = mbedtls_ssl_read(&secure->ssl, into, size);
	if (count > 0) return count;
	if (count == MBEDTLS_ERR_SSL_WANT_READ || count == MBEDTLS_ERR_SSL_WANT_WRITE) return 0;
	if (count == 0 || count == MBEDTLS_ERR_SSL_PEER_CLOSE_NOTIFY) return -1;
	return -2;
}

// writeSecure sends what TLS takes now: the bytes taken, 0 when it can't now (call it again
// later with the same bytes, as mbedtls asks), or mbedtls's error, negative.
static int writeSecure(secureConnection *secure, const unsigned char *bytes, int size) {
	int count = mbedtls_ssl_write(&secure->ssl, bytes, size);
	if (count >= 0) return count;
	if (count == MBEDTLS_ERR_SSL_WANT_READ || count == MBEDTLS_ERR_SSL_WANT_WRITE) return 0;
	return count;
}

// closeSecure says goodbye to the server, as TLS asks, and lets the connection go.
static void closeSecure(secureConnection *secure) {
	mbedtls_ssl_close_notify(&secure->ssl);
	freeSecure(secure);
}
*/
import "C"

import (
	_ "embed"
	"fmt"
	"unsafe"
)

// trustedRoots are the certificate authorities a secure connection trusts: Let's Encrypt's roots,
// ISRG Root X1 and X2 (https://letsencrypt.org/certificates/). A server whose certificate comes
// from anyone else is refused; add its root here.
//
//go:embed roots/isrg-root-x1.pem
var isrgRootX1 string

//go:embed roots/isrg-root-x2.pem
var isrgRootX2 string

// handshakeTimeoutMilliseconds is how long TLS's handshake may take: the 3DS's processor works
// out the keys in about a second.
const handshakeTimeoutMilliseconds = 10000

// DialSecure connects to a host on a port and talks TLS 1.2 over it, checking that the server's
// certificate comes from a trusted root and names host. It waits for the handshake, ten seconds
// at most.
func DialSecure(host string, port int) (*Connection, error) {
	connection, err := Dial(host, port)
	if err != nil {
		return nil, err
	}
	roots := C.CString(isrgRootX1 + "\n" + isrgRootX2)
	defer C.free(unsafe.Pointer(roots))
	name := C.CString(host)
	defer C.free(unsafe.Pointer(name))
	var problem C.int
	var flags C.uint
	rootsSize := C.size_t(len(isrgRootX1) + 1 + len(isrgRootX2) + 1) // with the NUL, as mbedtls wants for PEM
	secure := C.openSecure(connection.handle, name, roots, rootsSize, handshakeTimeoutMilliseconds, &problem, &flags)
	if secure == nil {
		connection.Close() //nolint:errcheck // the handshake's error is the one to report
		if flags != 0 {
			return nil, fmt.Errorf("%w: %s's certificate (flags %#x)", ErrCertificate, host, uint(flags))
		}
		return nil, fmt.Errorf("%w: %s (mbedtls %#x)", ErrSecure, host, -int(problem))
	}
	connection.secure = unsafe.Pointer(secure)
	return connection, nil
}

// readSecure, writeSecure and closeSecure are Connection's reads, writes and closing through
// TLS. They are here because C functions are seen only from the file that declares them.
func readSecure(secure unsafe.Pointer, buffer []byte) int {
	return int(C.readSecure((*C.secureConnection)(secure), (*C.uchar)(unsafe.Pointer(&buffer[0])), C.int(len(buffer))))
}

func writeSecure(secure unsafe.Pointer, bytes []byte) int {
	return int(C.writeSecure((*C.secureConnection)(secure), (*C.uchar)(unsafe.Pointer(&bytes[0])), C.int(len(bytes))))
}

func closeSecure(secure unsafe.Pointer) {
	C.closeSecure((*C.secureConnection)(secure))
}
