package pixelart

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"
)

// Theme is a go-ui theme in the game's look, over base: night-blue glass, starlight ink, a sky
// blue for what is chosen, the pixel fonts' sizes, square corners and hard shadows.
func Theme(base ui.Theme) ui.Theme {
	theme := base
	theme.Background = draw2d.MustHex("#0a1430")
	theme.Surface = draw2d.MustHex("#12224a")
	theme.Sunken = draw2d.MustHex("#1a2d5c")
	theme.Line = draw2d.MustHex("#3a5590")
	theme.ControlLine = draw2d.MustHex("#9cc8f2")
	theme.Ink = draw2d.MustHex("#eaf4ff")
	theme.InkSoft = draw2d.MustHex("#a9bedd")
	theme.InkQuiet = draw2d.MustHex("#7f93b8")
	theme.Primary = draw2d.MustHex("#5fb0ff")
	theme.OnPrimary = draw2d.MustHex("#02060f")
	theme.PrimaryTint = draw2d.MustHex("#2a4a80")
	theme.FocusRing = draw2d.MustHex("#5fb0ff80")
	theme.Shadow = draw2d.MustHex("#02060f")
	theme.TextSize = TextSize
	theme.SmallTextSize = TextSize
	theme.HeadingSize = TextSize
	theme.ControlRadius = 0
	theme.PanelRadius = 0
	theme.ShadowBlur = 0 // a hard shadow, a pixel's step below
	theme.ShadowDrop = 2
	return theme
}
