//go:build nintendo3ds

// A check of the network package: it connects to an echo server, sends a line and prints PASS when
// the same line comes back. The host and port are set when building:
//
//	tinygo build -target=nintendo3ds -ldflags "-X main.host=192.168.1.20 -X main.port=7000" -o echo.3dsx ./cmd/echo
package main

import (
	"strconv"
	"time"

	"github.com/akrck02/go-3ds/network"
)

var host, port = "127.0.0.1", "7000"

const line = "hello from the 3DS\n"

// patience is how long the answer may take.
const patience = 5 * time.Second

func main() {
	portNumber, _ := strconv.Atoi(port) // a bad port fails to connect below, with its own error
	connection, err := network.Dial(host, portNumber)
	if err != nil {
		println(err.Error())
		println("FAIL")
		return
	}
	defer connection.Close()
	if _, err := connection.Write([]byte(line)); err != nil {
		println(err.Error())
		println("FAIL")
		return
	}
	received := []byte{}
	buffer := make([]byte, 64)
	deadline := time.Now().Add(patience)
	for time.Now().Before(deadline) && len(received) < len(line) {
		count, err := connection.Read(buffer)
		if err != nil {
			println(err.Error())
			break
		}
		received = append(received, buffer[:count]...)
		time.Sleep(10 * time.Millisecond)
	}
	println("received:", string(received))
	if string(received) == line {
		println("PASS")
		return
	}
	println("FAIL")
}
