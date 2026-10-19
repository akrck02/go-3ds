//go:build nintendo3ds

// Package keyboard asks the player to type with the system's own keyboard (the software keyboard
// applet), the one the console's settings use. The applet takes both screens until the player
// confirms or cancels, so a program calls it between frames.
package keyboard

/*
#include <3ds.h>
#include <stdlib.h>

// askText shows the keyboard and returns 1 when the player confirmed, 0 when they cancelled.
static int askText(int numbers, int hidden, const char *hint, const char *initial, char *into, int size, int maximum) {
	SwkbdState keyboard;
	swkbdInit(&keyboard, numbers ? SWKBD_TYPE_NUMPAD : SWKBD_TYPE_NORMAL, 2, maximum);
	if (numbers) swkbdSetNumpadKeys(&keyboard, '.', ':');
	if (hidden) swkbdSetPasswordMode(&keyboard, SWKBD_PASSWORD_HIDE_DELAY);
	swkbdSetHintText(&keyboard, hint);
	swkbdSetInitialText(&keyboard, initial);
	swkbdSetValidation(&keyboard, SWKBD_NOTEMPTY_NOTBLANK, 0, 0);
	return swkbdInputText(&keyboard, into, size) == SWKBD_BUTTON_CONFIRM;
}
*/
import "C"

import "unsafe"

// Kind is which keys the keyboard offers.
type Kind int

const (
	// Text is the full keyboard.
	Text Kind = iota
	// Numbers is the number pad, with "." and ":" on its two extra keys: for addresses and ports.
	Numbers
	// Password is the full keyboard, each character hidden a second after it is typed.
	Password
)

// Ask shows the keyboard with a hint and the text to start from, and returns what the player
// typed, at most maximum characters. confirmed is false when they cancelled; text is then initial.
func Ask(kind Kind, hint, initial string, maximum int) (text string, confirmed bool) {
	cHint, cInitial := C.CString(hint), C.CString(initial)
	defer C.free(unsafe.Pointer(cHint))
	defer C.free(unsafe.Pointer(cInitial))
	// Each character can take up to four bytes in UTF-8, and one more ends the text.
	buffer := make([]byte, maximum*4+1)
	numbers := C.int(0)
	if kind == Numbers {
		numbers = 1
	}
	hidden := C.int(0)
	if kind == Password {
		hidden = 1
	}
	if C.askText(numbers, hidden, cHint, cInitial, (*C.char)(unsafe.Pointer(&buffer[0])), C.int(len(buffer)), C.int(maximum)) == 0 {
		return initial, false
	}
	return C.GoString((*C.char)(unsafe.Pointer(&buffer[0]))), true
}
