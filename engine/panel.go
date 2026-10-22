package main

import (
	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/pixelart"
	"github.com/jonbul/ships3D-3ds/touchscreen"
)

// The panel is the touch screen's picture where a platform has no touch screen: on the desktop,
// in the middle of the window, over the game, and touched with the mouse.

// windowWidth and windowHeight are the window's size in pixels, as the last frame painted it: the
// mouse is read against it.
var windowWidth, windowHeight float32

// wasClicking is whether the mouse's button was down last frame: a click counts once, as it lands.
var wasClicking bool

// panelArea is where the panel shows in a screen width × height of the 3DS top screen's pixels:
// in the middle.
func panelArea(width, height float32) draw2d.Rect {
	return draw2d.Rect{X: (width - touchscreen.Width) / 2, Y: (height - touchscreen.Height) / 2, Width: touchscreen.Width, Height: touchscreen.Height}
}

// paintPanel paints the touch screen's picture (paint, in its own pixels) in the middle of a
// screen width × height of the top screen's pixels.
func paintPanel(drawing *draw2d.List, width, height float32, paint func(drawing *draw2d.List)) {
	area := panelArea(width, height)
	start := len(drawing.Commands)
	paint(drawing)
	pixelart.Move(drawing, start, area.X, area.Y)
}

// panelInput is what the touch screen's pages read of this frame: the keys, the touch screen
// where there is one, and otherwise a click on the panel, in the panel's pixels.
func panelInput(keys *engine.Input) touchscreen.Input {
	read := touchInput(keys)
	if hasTouchScreen {
		return read
	}
	x, y, clicking := keys.Pointer()
	landed := clicking && !wasClicking
	wasClicking = clicking
	if !landed {
		return read
	}
	touchX, touchY, inside := onPanel(x, y)
	if inside {
		read.Pressed |= touchscreen.ButtonTouch
		read.TouchX, read.TouchY = touchX, touchY
	}
	return read
}

// onPanel is where a point of the window lands on the panel, in the panel's pixels, and whether
// it lands on it at all.
func onPanel(x, y float64) (touchX, touchY float32, inside bool) {
	scale := screenScale(windowWidth, windowHeight)
	area := panelArea(windowWidth/scale, windowHeight/scale)
	touchX, touchY = float32(x)/scale-area.X, float32(y)/scale-area.Y
	inside = touchX >= 0 && touchY >= 0 && touchX < area.Width && touchY < area.Height
	return touchX, touchY, inside
}
