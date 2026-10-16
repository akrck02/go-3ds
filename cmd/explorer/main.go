//go:build nintendo3ds

// The sample island to explore on the Nintendo 3DS: the circle pad or the cross turn the camera
// round it, L and R move it away and closer, dragging on the bottom screen turns it as a mouse
// would, and the 3D slider shows it in 3D.
//
//	tinygo build -target=nintendo3ds -o explorer.3dsx ./cmd/explorer
package main

import (
	"runtime"

	"github.com/akrck02/go-3ds/app"
	"github.com/akrck02/go-renderer/camera"
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/models"
	"github.com/akrck02/go-renderer/scene/sample"
)

// How fast the circle pad turns the camera at its edge, in radians a second.
const (
	circleYawSpeed   = 1.8
	circlePitchSpeed = 1.2
)

// The eyes stand this share of the distance to what the camera looks at apart, with the slider
// at its top.
const separationAtFullSlider = 0.03

func main() {
	screen, err := ctr.Open()
	if err != nil {
		println(err.Error())
		return
	}
	defer screen.Close()
	ctr.OpenBottomConsole()
	println("explorer: the sample island")
	println("circle pad or cross: turn")
	println("L / R: away / closer")
	println("drag on this screen: turn")
	println("3D slider: depth")
	println("START: exit")

	world := sample.Island()
	renderer, err := ctr.NewSceneRenderer(screen)
	if err != nil {
		println(err.Error())
		return
	}
	orbit := camera.NewOrbit(graphics.Vec4{0, 0, 0, 1}, 450)
	orbit.Pitch = 0.5
	orbit.MaximumDistance = 850 // the island's fog is total at 900
	orbit.Ground = world.GroundHeight
	view := models.Camera{Far: 20000}

	frames, elapsed := 0, 0.0
	var statistics runtime.MemStats
	app.Run(func(frame app.Frame) bool {
		circleX, circleY := frame.Reading.Circle()
		orbit.Yaw -= circleX * circleYawSpeed * frame.Seconds
		orbit.Pitch += circleY * circlePitchSpeed * frame.Seconds
		orbit.Update(frame.Input, frame.Seconds, &view)

		separation := ctr.Slider3D() * float32(orbit.Distance) * separationAtFullSlider
		renderer.DrawStereo(world, view, separation, float32(orbit.Distance))

		frames++
		elapsed += frame.Seconds
		if elapsed >= 1 {
			runtime.ReadMemStats(&statistics)
			print("\x1b[9;0H")
			println("frames a second:", frames, "   ")
			println("GPU (ms):       ", int(ctr.DrawingTime()*10)/10, ".", int(ctr.DrawingTime()*10)%10, "   ")
			println("3D:             ", int(ctr.Slider3D()*100), "%   ")
			println("collections:    ", statistics.NumGC, "   ")
			frames, elapsed = 0, 0
		}
		return true
	})
}
