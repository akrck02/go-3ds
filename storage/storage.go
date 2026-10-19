//go:build nintendo3ds

// Package storage keeps a program's small files on the console's SD card, in its own folder
// under sdmc:/3ds/: settings, a remembered login. libctru mounts the card when a .3dsx starts,
// and newlib's C files reach it.
package storage

/*
#include <stdio.h>
#include <stdlib.h>
#include <sys/stat.h>
#include <errno.h>

// readFile reads a whole file into *into (malloc'd, which the caller frees). It returns its size,
// -1 when there is no such file and -2 when it can't be read.
static long readFile(const char *path, char **into) {
	FILE *file = fopen(path, "rb");
	if (!file) return errno == ENOENT ? -1 : -2;
	if (fseek(file, 0, SEEK_END) != 0) { fclose(file); return -2; }
	long size = ftell(file);
	if (size < 0 || fseek(file, 0, SEEK_SET) != 0) { fclose(file); return -2; }
	*into = malloc(size > 0 ? size : 1);
	if (!*into) { fclose(file); return -2; }
	long read = fread(*into, 1, size, file);
	fclose(file);
	if (read != size) { free(*into); *into = NULL; return -2; }
	return size;
}

// writeFile replaces a file with bytes, through a temporary one renamed over it, so a power cut
// leaves the old file or the new one but never half of one. 0 when done.
static int writeFile(const char *path, const char *temporary, const char *bytes, long size) {
	FILE *file = fopen(temporary, "wb");
	if (!file) return -1;
	long written = fwrite(bytes, 1, size, file);
	if (fclose(file) != 0 || written != size) return -1;
	remove(path);
	return rename(temporary, path) == 0 ? 0 : -1;
}

// appendFile adds bytes at the end of a file, making it when missing. 0 when done.
static int appendFile(const char *path, const char *bytes, long size) {
	FILE *file = fopen(path, "ab");
	if (!file) return -1;
	long written = fwrite(bytes, 1, size, file);
	return fclose(file) == 0 && written == size ? 0 : -1;
}

static int makeFolder(const char *path) {
	return mkdir(path, 0777) == 0 || errno == EEXIST ? 0 : -1;
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

// Folder is where a program keeps its files: sdmc:/3ds/<name>/.
type Folder struct {
	path string
}

// Open is the folder of a program, made on the card when missing.
func Open(name string) (Folder, error) {
	folder := Folder{path: "sdmc:/3ds/" + name}
	for _, path := range []string{"sdmc:/3ds", folder.path} {
		cPath := C.CString(path)
		failed := C.makeFolder(cPath) != 0
		C.free(unsafe.Pointer(cPath))
		if failed {
			return Folder{}, fmt.Errorf("%w: can't make %s", ErrWrite, path)
		}
	}
	return folder, nil
}

// Load reads a file of the folder; ErrNotFound when it was never saved.
func (folder Folder) Load(name string) ([]byte, error) {
	path := C.CString(folder.path + "/" + name)
	defer C.free(unsafe.Pointer(path))
	var bytes *C.char
	size := C.readFile(path, &bytes)
	switch {
	case size == -1:
		return nil, ErrNotFound
	case size < 0:
		return nil, fmt.Errorf("%w: %s", ErrRead, name)
	}
	defer C.free(unsafe.Pointer(bytes))
	return C.GoBytes(unsafe.Pointer(bytes), C.int(size)), nil
}

// Save replaces a file of the folder with data.
func (folder Folder) Save(name string, data []byte) error {
	path := C.CString(folder.path + "/" + name)
	defer C.free(unsafe.Pointer(path))
	temporary := C.CString(folder.path + "/" + name + ".new")
	defer C.free(unsafe.Pointer(temporary))
	var bytes *C.char
	if len(data) > 0 {
		bytes = (*C.char)(unsafe.Pointer(&data[0]))
	}
	if C.writeFile(path, temporary, bytes, C.long(len(data))) != 0 {
		return fmt.Errorf("%w: %s", ErrWrite, name)
	}
	return nil
}

// Append adds data at the end of a file of the folder, making it when missing: for a log.
func (folder Folder) Append(name string, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	path := C.CString(folder.path + "/" + name)
	defer C.free(unsafe.Pointer(path))
	if C.appendFile(path, (*C.char)(unsafe.Pointer(&data[0])), C.long(len(data))) != 0 {
		return fmt.Errorf("%w: %s", ErrWrite, name)
	}
	return nil
}

// Remove deletes a file of the folder, if it is there.
func (folder Folder) Remove(name string) {
	path := C.CString(folder.path + "/" + name)
	defer C.free(unsafe.Pointer(path))
	C.remove(path)
}
