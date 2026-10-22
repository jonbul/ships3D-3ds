package main

import (
	"log/slog"
	"math"

	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/opengl/overlay"

	"github.com/jonbul/ships3D-3ds/pixelart"
)

// The screens' look, the same as the browser's and the 3DS program's: windows of night-blue
// glass with a light edge and a hard shadow, a sky-blue accent, a star for what is chosen. Sizes
// are in the 3DS top screen's pixels; a larger screen shows them larger (paintScaled).
var colours = struct {
	glass, edge, shadow, ink, soft, accent, chosen, dim draw2d.Color
}{
	glass: draw2d.Hex("#12224ae0"), edge: draw2d.Hex("#9cc8f2"), shadow: draw2d.Hex("#02060fb0"),
	ink: draw2d.Hex("#eaf4ff"), soft: draw2d.Hex("#a9bedd"), accent: draw2d.Hex("#5fb0ff"),
	chosen: draw2d.Hex("#2a4a80"), dim: draw2d.Hex("#5f7a9a"),
}

// The windows' measures, in pixels.
const (
	windowSteps    = 2
	edgeWidth      = 2
	shadowDrop     = 3
	headingSize    = 16
	titleSize      = 40
	ribbonHeight   = 20
	promptBlinks   = 0.8 // a second
	promptShows    = 0.65
	cursorStarSize = 1.4
)

// cursorStar is the cursor beside what is chosen, as on the 3DS: '#' the accent, '+' lighter.
var cursorStar = []string{"...#...", "...#...", "..#+#..", "##+++##", "..#+#..", "...#...", "...#..."}

// measurer measures text, to centre it; made once, after the pixel fonts are chosen.
var measurer draw2d.Measurer

// guessedMeasurer guesses text's width when no measurer could be made, so the text still shows.
type guessedMeasurer struct{}

// guessedShare is how wide a letter is guessed to be, of the text's size.
const guessedShare = 0.6

func (guessedMeasurer) MeasureText(text string, size float32, _ draw2d.Weight, _ float32) draw2d.TextMetrics {
	return draw2d.TextMetrics{Width: float32(len(text)) * size * guessedShare}
}

func (guessedMeasurer) HasIcon(string) bool { return false }

// textMeasurer is the game's measurer of text.
func textMeasurer() draw2d.Measurer {
	if measurer != nil {
		return measurer
	}
	made, err := overlay.NewMeasurer()
	if err != nil {
		slog.Warn("text could not be measured: it is guessed", "error", err)
		measurer = guessedMeasurer{}
		return measurer
	}
	measurer = made
	return measurer
}

func textWidth(text string, size float32, weight draw2d.Weight) float32 {
	return textMeasurer().MeasureText(text, size, weight, 1).Width
}

// The top screen's heading, as the 3DS program lays it out with go-ui: the heading in the title's
// face from headingTop down, its caption just under it, both centred.
const (
	captionDrop       = 26 // from the heading's baseline to the caption's
	headingTop        = 28
	headingBaseline   = headingTop + titleSize
	captionBaseline   = headingBaseline + captionDrop
	headingInk        = "#e9edff"
	headingCaptionInk = "#aab4ea"
)

// topHeading draws a heading and its caption over the top of the top screen.
func topHeading(drawing *draw2d.List, width float32, heading, caption string) {
	centred(drawing, heading, width, headingBaseline, titleSize, draw2d.Bold, draw2d.Hex(headingInk))
	centred(drawing, caption, width, captionBaseline, pixelart.TextSize, draw2d.Regular, draw2d.Hex(headingCaptionInk))
}

// paintScaled paints a screen laid out in the layout's pixels (layout.go), each as many of the
// window's as fit: paint gets the width and height in those pixels.
func paintScaled(drawing *draw2d.List, width, height float32, paint func(drawing *draw2d.List, width, height float32)) {
	scale := screenScale(width, height)
	start := len(drawing.Commands)
	paint(drawing, width/scale, height/scale)
	pixelart.Scale(drawing, start, scale)
}

// window draws a window of glass over area, with a title on a ribbon of the accent when title
// isn't empty.
func window(drawing *draw2d.List, area draw2d.Rect, title string) {
	shadow := area
	shadow.X, shadow.Y = shadow.X+shadowDrop, shadow.Y+shadowDrop
	pixelart.Stepped(drawing, shadow, windowSteps, colours.shadow)
	pixelart.Stepped(drawing, area, windowSteps, colours.edge)
	pixelart.Stepped(drawing, pixelart.Inset(area, edgeWidth), windowSteps, colours.glass)
	if title == "" {
		return
	}
	ribbon := draw2d.Rect{X: area.X + edgeWidth, Y: area.Y + edgeWidth, Width: area.Width - 2*edgeWidth, Height: ribbonHeight}
	drawing.Rectangle(ribbon, 0, colours.accent, draw2d.Color{}, 0)
	drawing.Text(title, ribbon.X+8, ribbon.Y+ribbon.Height-5, pixelart.LabelSize, draw2d.Mono, colours.shadow)
}

// centred draws a line of text in the middle of width, its baseline at y.
func centred(drawing *draw2d.List, text string, width, y, size float32, weight draw2d.Weight, colour draw2d.Color) {
	drawing.Text(text, (width-textWidth(text, size, weight))/2, y, size, weight, colour)
}

// prompt is a line that blinks, as arcade screens' "press start" did.
func prompt(drawing *draw2d.List, text string, width, y float32, now float64) {
	if pixelart.Shown(now, promptBlinks, promptShows) {
		centred(drawing, text, width, y, pixelart.LabelSize, draw2d.Mono, colours.ink)
	}
}

// cursor draws the star beside what is chosen, bobbing to and fro.
func cursor(drawing *draw2d.List, x, y float32, now float64) {
	bob := float32(math.Round(2 * math.Sin(now*2*math.Pi*1.6)))
	pixelart.PaintedBitmap(drawing, cursorStar, x+bob, y, cursorStarSize,
		map[byte]draw2d.Color{'#': colours.accent, '+': colours.ink})
}

// hints draws the buttons' hints along the bottom of the screen.
func hints(drawing *draw2d.List, text string, width, height float32) {
	centred(drawing, text, width, height-8, pixelart.LabelSize, draw2d.Mono, colours.soft)
}
