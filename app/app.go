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

// startCStick turns on the service that reads the New 3DS's C-stick (ir:rst). On an original 3DS
// it fails quietly and the stick reads as centred.
static int startCStick(void) {
	return R_SUCCEEDED(irrstInit());
}

// readCStick reads the C-stick, after the controls were read this frame.
static void readCStick(int *x, int *y) {
	irrstScanInput();
	circlePosition position;
	irrstCstickRead(&position);
	*x = position.dx;
	*y = position.dy;
}
*/
import "C"

import (
	"math"
	"time"

	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/input"
)

// Frame is what a program gets every frame.
type Frame struct {
	Input   *input.State // keys, pointer and touches, as on the desktop
	Reading ctr.Reading  // the hardware itself, for the circle pad
	Pressed ctr.Buttons  // the buttons pressed this frame, not held since before
	// CStickX and CStickY are the New 3DS's C-stick, from -1 to 1 (up positive), 0 inside its dead
	// zone and on an original 3DS.
	CStickX, CStickY float64
	Seconds float64      // since the previous frame
}

// Run calls step once a frame, as fast as the screen shows frames (citro3d waits for it), until
// step returns false, START is pressed or the system asks the program to close.
func Run(step func(frame Frame) bool) {
	run(step, true)
}

// RunWithStart is Run for a program that gives START a use of its own (a pause menu): only step
// returning false, or the system, ends it.
func RunWithStart(step func(frame Frame) bool) {
	run(step, false)
}

func run(step func(frame Frame) bool, startLeaves bool) {
	C.fastCPU()
	state := input.NewState()
	state.FramebufferWidth, state.FramebufferHeight = ctr.BottomWidth, ctr.BottomHeight
	state.WindowWidth, state.WindowHeight = ctr.BottomWidth, ctr.BottomHeight // the touch screen, where the pointer is
	var reader ctr.Controls
	cStick := C.startCStick() != 0
	last := time.Now()
	for ctr.Running() {
		reading := ctr.ReadControls()
		pressed := reader.Pressed(reading)
		if startLeaves && pressed&ctr.ButtonStart != 0 {
			return
		}
		reader.Apply(state, reading)
		now := time.Now()
		seconds := now.Sub(last).Seconds()
		last = now
		frame := Frame{Input: state, Reading: reading, Pressed: pressed, Seconds: seconds}
		if cStick {
			frame.CStickX, frame.CStickY = readCStick()
		}
		if !step(frame) {
			return
		}
		state.EndFrame()
	}
}

// The C-stick's reach and dead zone: it reads about ±146 at its edge, and a little off centre at
// rest.
const (
	cStickReach    = 146
	cStickDeadZone = 0.15
)

func readCStick() (x, y float64) {
	var rawX, rawY C.int
	C.readCStick(&rawX, &rawY)
	x, y = float64(rawX)/cStickReach, float64(rawY)/cStickReach
	if math.Hypot(x, y) < cStickDeadZone {
		return 0, 0
	}
	return math.Max(-1, math.Min(1, x)), math.Max(-1, math.Min(1, y))
}
