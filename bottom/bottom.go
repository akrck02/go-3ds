//go:build nintendo3ds

// Package bottom shows a go-ui interface on the Nintendo 3DS's bottom screen: go-ui lays it out
// as a draw2d list, and go-renderer's ctr.FlatRenderer paints it. The touch screen is the pointer
// (go-renderer's ctr.Controls), so buttons, sliders and tabs answer the finger or the stylus.
package bottom

import (
	"github.com/akrck02/go-renderer/ctr"
	"github.com/akrck02/go-renderer/input"
	ui "github.com/akrck02/go-ui"
)

// Interface is a go-ui context on the bottom screen.
type Interface struct {
	Context  *ui.Context
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
	return &Interface{Context: context, renderer: renderer}, nil
}

// Paint builds the interface for this frame's input and paints it, between ctr.BeginFrame and
// ctr.EndFrame.
func (screen *Interface) Paint(state *input.State, build func(context *ui.Context)) {
	screen.Context.Begin(state)
	build(screen.Context)
	screen.Context.End()
	screen.renderer.Paint(screen.Context.Drawing(), screen.Context.Theme.Background)
}
