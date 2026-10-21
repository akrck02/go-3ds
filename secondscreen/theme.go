package secondscreen

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Theme is go-ui's dark theme made for the bottom screen: 320 × 240 pixels a finger touches, so
// controls are taller and text a little smaller than on a desktop. Its colours are the sea's: a
// deep blue, with a turquoise for what is chosen.
func Theme() ui.Theme {
	theme := ui.DarkTheme()
	theme.Background = draw2d.Hex("#0b1820")
	theme.Surface = draw2d.Hex("#10232d")
	theme.Sunken = draw2d.Hex("#183441")
	theme.Line = draw2d.Hex("#21424f")
	theme.ControlLine = draw2d.Hex("#2a5261")
	theme.Ink = draw2d.Hex("#e6f1f2")
	theme.InkSoft = draw2d.Hex("#a9c3c8")
	theme.InkQuiet = draw2d.Hex("#7895a0")
	theme.Primary = draw2d.Hex("#37c1b0")
	theme.OnPrimary = draw2d.Hex("#062421")
	theme.PrimaryTint = draw2d.Hex("#16403f")
	theme.FocusRing = draw2d.Hex("#37c1b080")
	theme.TextSize = 13
	theme.SmallTextSize = 11
	theme.HeadingSize = 16
	theme.IconSize = 16
	theme.ControlHeight = 30
	theme.Gap = 7
	theme.Padding = 10
	theme.ControlRadius = 8
	theme.PanelRadius = 12
	theme.PanelMinWidth = 0
	return theme
}
