//go:build nintendo3ds

// The sample island to explore on the Nintendo 3DS. The top screen shows the island, in 3D with
// the slider; the bottom screen is a go-ui panel, touched with a finger or the stylus: places to
// fly to, the time of day and the fog, and how fast frames go. The circle pad or the cross turn
// the camera, L and R move it away and closer.
//
//	tinygo build -target=nintendo3ds -o explorer.3dsx ./cmd/explorer
package main

import (
	"math"

	"github.com/akrck02/go-3ds/app"
	"github.com/akrck02/go-3ds/bottom"
	"github.com/akrck02/go-renderer/camera"
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/input"
	"github.com/akrck02/go-renderer/models"
	"github.com/akrck02/go-renderer/scene"
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

// The island's fog, when it is on.
const fogNear, fogFar = 250, 900

func main() {
	screen, err := ctr.Open()
	if err != nil {
		println(err.Error())
		return
	}
	defer screen.Close()
	panel, err := bottom.New(bottom.Theme())
	if err != nil {
		println(err.Error())
		return
	}
	renderer, err := ctr.NewSceneRenderer(screen)
	if err != nil {
		println(err.Error())
		return
	}

	explorer := newExplorer(sample.Island())
	view := models.Camera{Far: 20000}
	app.Run(func(frame app.Frame) bool {
		explorer.steer(frame, &view)
		explorer.figures.add(frame.Seconds, ctr.DrawingTime(), renderer.Statistics(0).Triangles)

		ctr.BeginFrame()
		separation := ctr.Slider3D() * float32(explorer.orbit.Distance) * separationAtFullSlider
		renderer.PaintStereo(explorer.world, view, separation, float32(explorer.orbit.Distance))
		panel.Paint(frame.Input, explorer.build)
		ctr.EndFrame()
		explorer.applyLight(frame.Input)
		return true
	})
}

// explorer is the program's state: the island, the camera and what the panel shows.
type explorer struct {
	world     *scene.Scene
	islandMap *draw2d.Image
	orbit     *camera.Orbit
	flight    flight
	tab       int
	hour      float32 // on the slider
	litHour   float32 // the hour the island is lit for
	fog       bool
	figures   figures
}

func newExplorer(world *scene.Scene) *explorer {
	orbit := camera.NewOrbit(graphics.Vec4{0, 0, 0, 1}, 450)
	orbit.Pitch = 0.5
	orbit.MaximumDistance = 850 // the fog is total at 900
	orbit.Ground = world.GroundHeight
	explorer := &explorer{world: world, islandMap: islandMap(world), orbit: orbit, hour: 10, fog: true}
	explorer.light(explorer.hour)
	return explorer
}

// steer turns and zooms the camera with the circle pad, the cross and L, R, or flies it to a
// place chosen on the panel, and places the view.
func (explorer *explorer) steer(frame app.Frame, view *models.Camera) {
	if explorer.flight.active {
		explorer.flight.step(explorer.orbit, frame.Seconds)
		explorer.orbit.Update(nil, frame.Seconds, view)
		return
	}
	circleX, circleY := frame.Reading.Circle()
	explorer.orbit.Yaw -= circleX * circleYawSpeed * frame.Seconds
	explorer.orbit.Pitch += circleY * circlePitchSpeed * frame.Seconds
	// The keys only: the touch screen is the panel's, so it does not turn the camera too.
	keys := *frame.Input
	keys.SetButton(input.MouseLeft, false)
	keys.ScrollDelta = 0
	explorer.orbit.Update(&keys, frame.Seconds, view)
}

// light sets the sun for an hour of the day: it rises in the east at six, is highest at noon and
// sets in the west at six in the evening.
func (explorer *explorer) light(hour float32) {
	dayAngle := float64(hour-6) / 12 * math.Pi
	elevation := math.Sin(dayAngle) * 1.1
	sun := graphics.Vec4{float32(-math.Cos(dayAngle) * math.Cos(elevation)), float32(math.Sin(elevation)), float32(-0.35 * math.Cos(elevation)), 0}
	explorer.world.Environment.SunDirection = sun.Normalize()
	explorer.litHour = hour
}

// applyLight lights the island for the slider's hour once the finger leaves the screen: a new
// light works the meshes' colours out again, which would make every step of a drag stutter.
func (explorer *explorer) applyLight(state *input.State) {
	if explorer.hour != explorer.litHour && !state.Button(input.MouseLeft) {
		explorer.light(explorer.hour)
	}
}

// setFog turns the island's fog on or off.
func (explorer *explorer) setFog(on bool) {
	explorer.fog = on
	explorer.world.Environment.FogNear, explorer.world.Environment.FogFar = fogNear, fogFar
	if !on {
		explorer.world.Environment.FogFar = 0
	}
}
