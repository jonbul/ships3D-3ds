package pixelart

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"
)

func TestScalingMakesWhatWasDrawnLargerFromTheTopLeft(test *testing.T) {
	var drawing draw2d.List
	drawing.Rectangle(draw2d.Rectangle{X: 1, Y: 2, Width: 3, Height: 4}, 0, draw2d.MustHex("#ffffff"), draw2d.Color{}, 0)
	drawing.Text("hi", 5, 6, 8, draw2d.Mono, draw2d.MustHex("#ffffff"))
	Scale(&drawing, 1, 2)
	if got := drawing.Commands[0].Area; got != (draw2d.Rectangle{X: 1, Y: 2, Width: 3, Height: 4}) {
		test.Fatalf("the command before start moved: %v", got)
	}
	text := drawing.Commands[1]
	if text.Area.X != 10 || text.Baseline != 12 || text.Size != 16 {
		test.Fatalf("the text is at %v, baseline %v, size %v", text.Area.X, text.Baseline, text.Size)
	}
}

func TestMovingADrawingMovesItsTextsBaselineToo(test *testing.T) {
	var drawing draw2d.List
	drawing.Text("hello", 10, 20, 16, draw2d.Regular, draw2d.Color{1, 1, 1, 1})
	Move(&drawing, 0, 5, 7)
	command := drawing.Commands[0]
	if command.Area.X != 15 || command.Baseline != 27 {
		test.Errorf("the text moved to x %v, baseline %v; want 15 and 27", command.Area.X, command.Baseline)
	}
}
