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

// startGyroscope turns the gyroscope on and returns how many of its raw steps make a degree a
// second (about 14 on a New 3DS), or 0 when it can't be turned on.
static float startGyroscope(void) {
	float stepsPerDegree = 0;
	if (R_FAILED(HIDUSER_EnableGyroscope())) return 0;
	if (R_FAILED(HIDUSER_GetGyroscopeRawToDpsCoefficient(&stepsPerDegree))) return 0;
	return stepsPerDegree;
}

static void stopGyroscope(void) {
	HIDUSER_DisableGyroscope();
}

// readGyroscope reads the turn rates, raw: roll, pitch and yaw as libctru names them.
static void readGyroscope(int *roll, int *pitch, int *yaw) {
	angularRate rate;
	hidGyroRead(&rate);
	*roll = rate.x;
	*pitch = rate.y;
	*yaw = rate.z;
}

static void stopCStick(void) {
	irrstExit();
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
	// Gyroscope is how fast the console turns, in degrees a second, on the axes libctru names roll,
	// pitch and yaw; all 0 when the gyroscope couldn't be turned on.
	Gyroscope Gyroscope
	Seconds   float64 // since the previous frame
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
	if cStick {
		defer C.stopCStick()
	}
	stepsPerDegree := float64(C.startGyroscope())
	if stepsPerDegree != 0 {
		defer C.stopGyroscope()
	}
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
		if stepsPerDegree != 0 {
			frame.Gyroscope = readGyroscope(stepsPerDegree)
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

// Gyroscope is the console's turn rates, in degrees a second.
type Gyroscope struct {
	Roll, Pitch, Yaw float64
}

// readGyroscope reads the rates in degrees a second: libctru's coefficient divides the raw steps
// (multiplying by it read a slight turn as thousands of degrees a second).
func readGyroscope(stepsPerDegree float64) Gyroscope {
	var roll, pitch, yaw C.int
	C.readGyroscope(&roll, &pitch, &yaw)
	return Gyroscope{Roll: float64(roll) / stepsPerDegree, Pitch: float64(pitch) / stepsPerDegree, Yaw: float64(yaw) / stepsPerDegree}
}
