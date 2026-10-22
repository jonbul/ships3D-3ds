package main

import (
	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/touchscreen"
)

// The game is laid out for its target. A 3DS has two screens: the game and its headings on the
// top one, the menus, the modes and the radar on the touch screen. Everywhere else there is one
// screen, and the touch screen's pieces go over the game, laid out for it: the menus in a window
// in the middle, the radar floating in a corner, the keys in a line of hints.

// oneScreen tells whether the game is laid out on one screen: every target but the 3DS.
func oneScreen() bool { return engine.RunningOn() != engine.Nintendo3DSTarget }

// The layout's pixels: a screen is laid out in pixels of the pixel art, each as many of the
// window's as fit. The 3DS lays it out in its top screen's 400 × 240; one screen in at least
// 640 × 360, so a window has room for the game beside the menus.
const (
	screenWide    = 400
	screenHigh    = 240
	oneScreenWide = 640
	oneScreenHigh = 360
)

// The menus' window on one screen, in the layout's pixels: as large as this, and margin from the
// window's edges.
const (
	panelWidth  = 560
	panelHeight = 320
	panelMargin = 8
)

// windowWidth and windowHeight are the window's size in pixels, as the last frame painted it: the
// mouse is read against it.
var windowWidth, windowHeight float32

// wasClicking is whether the mouse's button was down last frame: a click counts once, as it lands.
var wasClicking bool

// screenScale is how many of a window's pixels, width × height, each of the layout's pixels takes:
// as many as fit both ways, so nothing spills over the sides of a narrow window, and a whole
// number, so the pixel art stays crisp.
func screenScale(width, height float32) float32 {
	wide, high := float32(screenWide), float32(screenHigh)
	if oneScreen() {
		wide, high = oneScreenWide, oneScreenHigh
	}
	return float32(max(1, int(min(width/wide, height/high))))
}

// panelIn is where the menus and the modes are laid out on a screen width × height of the
// layout's pixels: the whole touch screen on a 3DS, a window in the middle on one screen.
func panelIn(width, height float32) draw2d.Rect {
	if !oneScreen() {
		return touchscreen.FullScreen
	}
	panelWide := min(panelWidth, width-2*panelMargin)
	panelHigh := min(panelHeight, height-2*panelMargin)
	return draw2d.Rect{X: (width - panelWide) / 2, Y: (height - panelHigh) / 2, Width: panelWide, Height: panelHigh}
}

// panel is where the menus and the modes are laid out now, as the last frame's window was.
func panel() draw2d.Rect {
	scale := screenScale(windowWidth, windowHeight)
	return panelIn(windowWidth/scale, windowHeight/scale)
}

// ground paints what is behind a touch screen's page: the 3DS's night on its touch screen, and on
// one screen nothing, the game shows through.
func ground(drawing *draw2d.List, now float64) {
	if !oneScreen() {
		touchscreen.Starfield(drawing, now)
	}
}

// panelInput is what the menus read of this frame: the keys, the touch screen where there is one,
// and on one screen a click, in the layout's pixels.
func panelInput(keys *engine.Input) touchscreen.Input {
	read := touchInput(keys)
	if !oneScreen() {
		return read
	}
	x, y, clicking := keys.Pointer()
	landed := clicking && !wasClicking
	wasClicking = clicking
	if landed {
		scale := float64(screenScale(windowWidth, windowHeight))
		read.Pressed |= touchscreen.ButtonTouch
		read.TouchX, read.TouchY = float32(x/scale), float32(y/scale)
	}
	return read
}
