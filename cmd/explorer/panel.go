//go:build nintendo3ds

package main

import (
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// build lays out the bottom screen: a heading with the frame rate, and three tabs.
func (explorer *explorer) build(context *ui.Context) {
	options := ui.PanelOptions{Dock: ui.DockLeft, Width: ui.Fixed(ctr.BottomWidth), Height: ui.Grow()}
	context.PanelWith("Sample island", options, func() {
		context.Tabs("tabs", &explorer.tab, "View", "Light", "Figures")
		switch explorer.tab {
		case 0:
			explorer.buildView(context)
		case 1:
			explorer.buildLight(context)
		default:
			explorer.buildFigures(context)
		}
	})
}

// buildView shows the island's map beside the places to fly to, and the distance.
func (explorer *explorer) buildView(context *ui.Context) {
	context.Row(func() {
		context.Image(explorer.islandMap, ui.Layout{Width: ui.Fixed(mapSize), Height: ui.Fixed(mapSize)}, context.Theme.ControlRadius)
		context.Column(func() {
			context.Caption("Fly to")
			for _, destination := range places {
				if context.IconButton(destination.icon, destination.name) {
					explorer.flight.start(explorer.orbit, destination)
				}
			}
		})
	})
	distance := float32(explorer.orbit.Distance)
	if context.Slider("distance", &distance, float32(explorer.orbit.MinimumDistance), float32(explorer.orbit.MaximumDistance)) {
		explorer.flight.active = false
		explorer.orbit.Distance = float64(distance)
	}
}

// buildLight offers the time of day and the fog.
func (explorer *explorer) buildLight(context *ui.Context) {
	context.Row(func() {
		icon := "sun"
		if explorer.hour < 6.5 || explorer.hour > 18.5 {
			icon = "moon"
		}
		context.Icon(icon)
		context.Text("Time of day", ui.TextStyle{})
		context.Spacer()
		context.Text(clock(explorer.hour), ui.TextStyle{Weight: draw2d.Bold})
	})
	context.SliderStep("hour", &explorer.hour, 4, 22, 0.25)
	fog := explorer.fog
	if context.Toggle("Fog", &fog) {
		explorer.setFog(fog)
	}
	context.Caption("The light changes when you lift your finger.")
}

// buildFigures shows how fast frames go.
func (explorer *explorer) buildFigures(context *ui.Context) {
	figures := explorer.figures.shown
	theme := context.Theme
	look := ui.ChipLook{Fill: theme.PrimaryTint, Ink: theme.Primary}
	if figures.framesPerSecond < 55 {
		look = ui.ChipLook{Fill: theme.Sunken, Ink: theme.Warning}
	}
	context.Row(func() {
		context.Chip("gauge", itoa(figures.framesPerSecond)+" fps", look)
		context.Chip("cpu", tenths(figures.gpu)+" ms GPU", ui.ChipLook{Fill: theme.Sunken, Ink: theme.InkSoft})
	})
	context.Row(func() {
		context.Chip("triangle", itoa(figures.triangles)+" triangles", ui.ChipLook{Fill: theme.Sunken, Ink: theme.InkSoft})
		context.Chip("glasses", itoa(int(ctr.Slider3D()*100))+"% 3D", ui.ChipLook{Fill: theme.Sunken, Ink: theme.InkSoft})
	})
	context.Caption("Drawn by go-renderer, laid out by go-ui.")
}

// clock writes an hour of the day as 9:45.
func clock(hour float32) string {
	minutes := int(hour*60 + 0.5)
	minute := minutes % 60
	text := itoa(minutes/60) + ":"
	if minute < 10 {
		text += "0"
	}
	return text + itoa(minute)
}
