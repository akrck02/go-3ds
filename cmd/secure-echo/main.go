//go:build nintendo3ds

// A check of TLS: it fetches a page over HTTPS and prints PASS when the server answers with a
// status line. The host, port and path are set when building:
//
//	tinygo build -target=nintendo3ds -ldflags "-X main.host=example.org" -o secure.3dsx ./cmd/secure-echo
package main

import (
	"strconv"
	"strings"
	"time"

	"github.com/akrck02/go-3ds/network"
)

var host, port, path = "letsencrypt.org", "443", "/"

// patience is how long the answer may take.
const patience = 10 * time.Second

func main() {
	portNumber, _ := strconv.Atoi(port) // a bad port fails to connect below, with its own error
	started := time.Now()
	connection, err := network.DialSecure(host, portNumber)
	if err != nil {
		println(err.Error())
		println("FAIL")
		return
	}
	defer connection.Close()
	println("handshake:", time.Since(started).Milliseconds(), "ms")
	request := "GET " + path + " HTTP/1.1\r\nHost: " + host + "\r\nConnection: close\r\n\r\n"
	if _, err := connection.Write([]byte(request)); err != nil {
		println(err.Error())
		println("FAIL")
		return
	}
	received := []byte{}
	buffer := make([]byte, 512)
	deadline := time.Now().Add(patience)
	for time.Now().Before(deadline) && !strings.Contains(string(received), "\r\n") {
		count, err := connection.Read(buffer)
		if err != nil {
			break
		}
		received = append(received, buffer[:count]...)
		time.Sleep(10 * time.Millisecond)
	}
	firstLine, _, _ := strings.Cut(string(received), "\r\n")
	println("answer:", firstLine)
	if strings.HasPrefix(firstLine, "HTTP/1.1 ") {
		println("PASS")
		return
	}
	println("FAIL")
}
