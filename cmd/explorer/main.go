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

	var second secondFigures
	var statistics runtime.MemStats
	app.Run(func(frame app.Frame) bool {
		circleX, circleY := frame.Reading.Circle()
		orbit.Yaw -= circleX * circleYawSpeed * frame.Seconds
		orbit.Pitch += circleY * circlePitchSpeed * frame.Seconds
		orbit.Update(frame.Input, frame.Seconds, &view)

		separation := ctr.Slider3D() * float32(orbit.Distance) * separationAtFullSlider
		renderer.DrawStereo(world, view, separation, float32(orbit.Distance))
		second.add(frame.Seconds, float32(renderer.PlanTime().Microseconds())/1000, ctr.ProcessingTime(), ctr.DrawingTime())

		if second.elapsed >= 1 {
			runtime.ReadMemStats(&statistics)
			drawn := renderer.Statistics(0)
			print("\x1b[8;0H")
			println("frames a second:", second.frames, "   ")
			println("            mean   most")
			printMilliseconds("plan (ms):   ", second.plan/float32(second.frames), second.mostPlan)
			printMilliseconds("CPU (ms):    ", second.processing/float32(second.frames), second.mostProcessing)
			printMilliseconds("GPU (ms):    ", second.drawing/float32(second.frames), second.mostDrawing)
			println("triangles:      ", drawn.Triangles, "     ")
			println("baked meshes:   ", renderer.Bakes(), "   ")
			println("collections:    ", statistics.NumGC, "   ")
			println("3D:             ", int(ctr.Slider3D()*100), "%   ")
			second = secondFigures{}
		}
		return true
	})
}

// secondFigures gathers a second's frames: how many, and the mean and the worst of each time.
type secondFigures struct {
	frames                                int
	elapsed                               float64
	plan, processing, drawing             float32
	mostPlan, mostProcessing, mostDrawing float32
}

func (figures *secondFigures) add(seconds float64, plan, processing, drawing float32) {
	figures.frames++
	figures.elapsed += seconds
	figures.plan += plan
	figures.processing += processing
	figures.drawing += drawing
	figures.mostPlan = max(figures.mostPlan, plan)
	figures.mostProcessing = max(figures.mostProcessing, processing)
	figures.mostDrawing = max(figures.mostDrawing, drawing)
}

// printMilliseconds prints two times with one decimal: the console's println has no float format.
func printMilliseconds(label string, mean, most float32) {
	println(label, tenths(mean), " ", tenths(most), "   ")
}

func tenths(milliseconds float32) string {
	whole := int(milliseconds*10 + 0.5)
	return itoa(whole/10) + "." + itoa(whole%10)
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}
