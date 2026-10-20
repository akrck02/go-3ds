//go:build nintendo3ds

// A check of the storage package: it saves a file on the SD card, reads it back and prints PASS
// when it is the same; run twice, the second run also finds the first one's file.
package main

import (
	"errors"

	"github.com/akrck02/go-3ds/storage"
)

func main() {
	folder, err := storage.Open("storage-check")
	if err != nil {
		println(err.Error())
		println("FAIL")
		return
	}
	earlier, err := folder.Load("check.txt")
	switch {
	case errors.Is(err, storage.ErrNotFound):
		println("first run: nothing saved yet")
	case err != nil:
		println(err.Error())
	default:
		println("found from before:", string(earlier))
	}
	const text = "saved by go-3ds"
	if err := folder.Save("check.txt", []byte(text)); err != nil {
		println(err.Error())
		println("FAIL")
		return
	}
	read, err := folder.Load("check.txt")
	if err != nil || string(read) != text {
		println("read back:", string(read))
		println("FAIL")
		return
	}
	println("PASS")
}
