package main

import (
	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/jonbul/ships3D-3ds/pixelart"
	"github.com/jonbul/ships3D-3ds/touchscreen"
)

// The touch screen is the 3DS's bottom one, painted by package touchscreen as the hand-written
// program paints it: the title's prompt, the modes, the hangar's arrows, the radar and the menu.

// stickSide is where the circle pad was pushed last frame: -1 left, 0 the middle, 1 right. A push
// to a side counts once, as the cross's press does.
var stickSide int

// stickPush is how far the circle pad must go to a side to count as a push, of its reach.
const stickPush = 0.5

// wasTouching is whether the touch screen was touched last frame: a touch counts once, as it
// lands.
var wasTouching bool

// menuActions are the actions the touch screen's menus read, by the buttons they stand for.
var menuActions = []struct {
	action string
	button touchscreen.Buttons
}{
	{"up", touchscreen.ButtonUp}, {"down", touchscreen.ButtonDown},
	{"left", touchscreen.ButtonLeft}, {"right", touchscreen.ButtonRight},
	{"confirm", touchscreen.ButtonA}, {"back", touchscreen.ButtonB},
	{"previous page", touchscreen.ButtonL}, {"next page", touchscreen.ButtonR},
}

// touchInput is what the touch screen reads of this frame: the buttons pressed, and a touch that
// has just landed, in the touch screen's pixels. Call it once a frame.
func touchInput(keys *engine.Input) touchscreen.Input {
	hintKeys = keys
	var read touchscreen.Input
	for _, each := range menuActions {
		if keys.Action(each.action).Pressed {
			read.Pressed |= each.button
		}
	}
	read.Pressed |= stickButtons(keys)
	x, y, touching := keys.SecondScreenTouch()
	if touching && !wasTouching {
		read.Pressed |= touchscreen.ButtonTouch
		read.TouchX, read.TouchY = float32(x), float32(y)
	}
	wasTouching = touching
	return read
}

// stickButtons are the cross's left or right, when the circle pad has just been pushed that way.
func stickButtons(keys *engine.Input) touchscreen.Buttons {
	x, _ := keys.Stick()
	side := 0
	switch {
	case x > stickPush:
		side = 1
	case x < -stickPush:
		side = -1
	}
	pushed := side != stickSide
	stickSide = side
	switch {
	case pushed && side < 0:
		return touchscreen.ButtonLeft
	case pushed && side > 0:
		return touchscreen.ButtonRight
	}
	return 0
}

// showTouchScreen makes paint the touch screen, in its pixels, until another scene changes it.
func showTouchScreen(paint func(drawing *draw2d.List)) {
	layout := ui.Layout{Width: ui.Fixed(touchscreen.Width), Height: ui.Fixed(touchscreen.Height)}
	drawInterface(func(context *ui.Context) {
		context.Paint(layout, func(area draw2d.Rectangle, drawing *draw2d.List) {
			start := len(drawing.Commands)
			paint(drawing)
			pixelart.Move(drawing, start, area.X, area.Y)
		})
	})
}

// playMenuSound plays the sound a touch screen's menu asked for.
func playMenuSound(sound touchscreen.Sound) {
	playSound(func(played *soundSet) {
		switch sound {
		case touchscreen.Moved:
			played.Play(played.Move, 1)
		case touchscreen.Chose:
			played.Play(played.Choose, 1)
		case touchscreen.Went:
			played.Play(played.Back, 1)
		}
	})
}
