package touchscreen

import (
	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/pixelart"
)

// The title screen's prompt: it blinks, as arcade screens did, at blinksPerSecond.
const (
	blinksPerSecond  = 1.5
	promptShowsShare = 0.65 // of each blink
	promptY          = 124
)

// TitlePrompt draws the title screen's touch screen clean: the night, its stars and a blinking
// prompt; the name is on the top screen.
func TitlePrompt(drawing *draw2d.List, measurer draw2d.Measurer, now float64) {
	Starfield(drawing, now)
	if pixelart.Shown(now, blinksPerSecond, promptShowsShare) {
		centred(drawing, measurer, "TOUCH OR PRESS A", promptY, labelPixels, draw2d.Mono, Colours.Ink)
	}
}

// The ship screen's measures, in pixels.
const (
	shipArrowY      = (Height - 2*shipArrowWidth) / 2
	shipArrowWidth  = 8
	shipArrowMargin = 12
	textShadowDrop  = 2
)

// ShipChoice draws over the hangar only the arrows the cross turns, at its sides, and a problem,
// if any, at the bottom: the name and which ship of how many are on the top screen.
func ShipChoice(drawing *draw2d.List, measurer draw2d.Measurer, problem string) {
	arrow(drawing, shipArrowMargin, shipArrowY, -1)
	arrow(drawing, Width-shipArrowMargin-shipArrowWidth, shipArrowY, 1)
	if problem != "" {
		shadowedText(drawing, measurer, Fitted(problem, LabelCharacters), Height-10, pixelTextSize, draw2d.Regular, Colours.Danger)
	}
}

// ShipArrowTouched tells which of the hangar's arrows a touch at x is on: -1 the left, 1 the
// right, 0 neither (the touch screen's left and right thirds).
func ShipArrowTouched(x float32) int {
	switch {
	case x < Width/3:
		return -1
	case x >= 2*Width/3:
		return 1
	}
	return 0
}

// shadowedText draws a centred line with a hard shadow a pixel below and right, so that it reads
// over the hangar's lights.
func shadowedText(drawing *draw2d.List, measurer draw2d.Measurer, text string, y, size float32, weight draw2d.Weight, colour draw2d.Color) {
	x := (Width - measurer.MeasureText(text, size, weight, 1).Width) / 2
	drawing.Text(text, x+textShadowDrop, y+textShadowDrop, size, weight, Colours.Shadow)
	drawing.Text(text, x, y, size, weight, colour)
}

// arrow draws a pixel triangle pointing left (-1) or right (1), shipArrowWidth wide.
func arrow(drawing *draw2d.List, x, y float32, direction int) {
	const height = 2*shipArrowWidth - 1
	for row := range height {
		width := float32(shipArrowWidth - int(abs(float32(row-shipArrowWidth+1))))
		left := x
		if direction < 0 {
			left = x + shipArrowWidth - width
		}
		drawing.Rectangle(draw2d.Rectangle{X: left + textShadowDrop, Y: y + float32(row) + textShadowDrop, Width: width, Height: 1}, 0, Colours.Shadow, draw2d.Color{}, 0)
		drawing.Rectangle(draw2d.Rectangle{X: left, Y: y + float32(row), Width: width, Height: 1}, 0, Colours.GoldLight, draw2d.Color{}, 0)
	}
}

// centred draws a line in the middle of the screen's width, its baseline at y.
func centred(drawing *draw2d.List, measurer draw2d.Measurer, text string, y, size float32, weight draw2d.Weight, colour draw2d.Color) {
	drawing.Text(text, (Width-measurer.MeasureText(text, size, weight, 1).Width)/2, y, size, weight, colour)
}
