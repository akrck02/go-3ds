//go:build nintendo3ds

// Package secondscreen shows a go-ui interface on the second screen, the Nintendo 3DS's bottom
// one: go-ui lays it out as a draw2d list, and go-renderer's ctr.FlatRenderer paints it. The touch
// screen is the pointer (go-renderer's ctr.Controls), so buttons, sliders and tabs answer the
// finger or the stylus.
package secondscreen

import (
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// Interface is a go-ui context on the second screen.
type Interface struct {
	Context  *ui.Context
	screen   *ctr.BottomScreen
	renderer *ctr.FlatRenderer
}

// New prepares the bottom screen and a go-ui context with the given theme, after ctr.Open.
func New(theme ui.Theme) (*Interface, error) {
	screen, err := ctr.OpenBottomScreen()
	if err != nil {
		return nil, err
	}
	renderer, err := ctr.NewFlatRenderer(screen)
	if err != nil {
		return nil, err
	}
	context, err := ui.New()
	if err != nil {
		return nil, err
	}
	context.Theme = theme
	return &Interface{Context: context, screen: screen, renderer: renderer}, nil
}

// Paint builds the interface for this frame's input and paints it, between ctr.BeginFrame and
// ctr.EndFrame.
func (screen *Interface) Paint(state *input.State, build func(context *ui.Context)) {
	screen.Context.Begin(state)
	build(screen.Context)
	screen.Context.End()
	screen.renderer.Paint(screen.Context.Drawing(), screen.Context.Theme.Background)
}

// Screen is the bottom screen, for drawing a scene on it (ctr.SceneRenderer's PaintOn) under the
// interface painted with PaintOver.
func (screen *Interface) Screen() *ctr.BottomScreen { return screen.screen }

// PaintOver builds the interface and paints it over what the bottom screen shows already, without
// clearing it: a scene drawn first with PaintOn shows through where the interface draws nothing.
func (screen *Interface) PaintOver(state *input.State, build func(context *ui.Context)) {
	screen.Context.Begin(state)
	build(screen.Context)
	screen.Context.End()
	screen.renderer.PaintOver(screen.screen, screen.Context.Drawing())
}
