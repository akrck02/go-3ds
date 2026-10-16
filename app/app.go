//go:build nintendo3ds

// Package app runs a program on the Nintendo 3DS: it reads the controls into go-renderer's
// input.State every frame, and stops when START is pressed or the system asks the program to
// close (the HOME menu, the power button).
package app

/*
#include <3ds.h>

// fastCPU asks for the New 3DS's 804 MHz clock and its L2 cache; the original 3DS ignores it.
static void fastCPU(void) {
	osSetSpeedupEnable(true);
}
*/
import "C"

import (
	"time"

	"github.com/akrck02/go-3ds/controls"
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/input"
)

// Frame is what a program gets every frame.
type Frame struct {
	Input   *input.State     // keys, pointer and touches, as on the desktop
	Reading controls.Reading // the hardware itself, for the circle pad
	Seconds float64          // since the previous frame
}

// Run calls step once a frame, as fast as the screen shows frames (citro3d waits for it), until
// step returns false, START is pressed or the system asks the program to close.
func Run(step func(frame Frame) bool) {
	C.fastCPU()
	state := input.NewState()
	state.FramebufferWidth, state.FramebufferHeight = ctr.BottomWidth, ctr.BottomHeight
	state.WindowWidth, state.WindowHeight = ctr.BottomWidth, ctr.BottomHeight // the touch screen, where the pointer is
	var reader controls.Controls
	last := time.Now()
	for ctr.Running() {
		reading := controls.Read()
		if reader.Pressed(reading)&controls.ButtonStart != 0 {
			return
		}
		reader.Apply(state, reading)
		now := time.Now()
		seconds := now.Sub(last).Seconds()
		last = now
		if !step(Frame{Input: state, Reading: reading, Seconds: seconds}) {
			return
		}
		state.EndFrame()
	}
}
