// Package hud draws the game's readings over the 3D view, as the 3DS shows them on its top
// screen: life as a capsule of charges with a heart at the top left, an assault round's mothership
// and time at the top right, and the copilot's box at the bottom when it has something to say.
// Any client paints them in a draw2d list over its view.
package hud

import (
	"math"
	"strconv"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/game"
	"github.com/jonbul/ships3D-3ds/pixelart"
)

// The top screen's readings over the game: life as a bar of charges at the top left, and the
// copilot's word at the bottom when there is something to say. Sizes are in pixels.
const (
	lifeLeft          = 10
	lifeTop           = 10
	lifeMostSegments  = 10
	lifeSegmentWidth  = 10
	lifeSegmentHeight = 10
	lifeSegmentGap    = 2
	lifeHeartRoom     = 24 // the heart's place at the bar's left
	lifeBeatsASecond  = 1.6
	copilotPixel      = 2.4 // the copilot's face is drawn this many pixels a pixel
	copilotFaceLeft   = 8
	copilotWordsLeft  = 52
	copilotWordsShare = 0.75 // of each blink the words show
	heartBeatShare    = 0.3  // of each beat the heart is bigger
	hudShadowDrop     = 2
	lifeLowShare      = 0.3 // below this much life the charges turn red
	copilotHeight     = 44
	copilotMargin     = 8
	copilotWarnShare  = 0.35 // how near the edge the copilot speaks up
	copilotBlinks     = 2.0  // the warning's blinks a second
)

// copilotEdgeWarning is what the copilot says as the ship heads out of the arena.
const copilotEdgeWarning = "We're leaving the combat arena, sir!"

// copilotTurningBack is what the copilot says while it turns the ship back itself.
const copilotTurningBack = "Taking the helm, sir. Back to the arena!"

// copilotFace is the copilot, a little round robot: a shell ('#', '='), a dark screen for a
// face ('-'), big eyes with a shine ('o', '*'), blushing cheeks ('p'), a smile ('s') and an
// antenna with a light on top ('a').
var copilotFace = []string{
	"......aa......",
	"......aa......",
	".......#......",
	"...########...",
	"..##------##..",
	".##--------##.",
	".#--oo--oo--#.",
	".#-o*o--o*o-#.",
	".#--oo--oo--#.",
	".#pp------pp#.",
	".#----ss----#.",
	".##--------##.",
	"..##======##..",
	"...########...",
}

// copilotPalette are the copilot's colours, by letter.
var copilotPalette = map[byte]draw2d.Color{
	'#': draw2d.MustHex("#dfe9f8"), '=': draw2d.MustHex("#9cb4d6"), '-': draw2d.MustHex("#16233f"),
	'o': draw2d.MustHex("#5ee0ff"), '*': draw2d.MustHex("#ffffff"), 'p': draw2d.MustHex("#ff8fb8"),
	's': draw2d.MustHex("#5ee0ff"), 'a': draw2d.MustHex("#ff8fb8"),
}

// lifeHeart is the bar's heart, and its colours.
var (
	lifeHeart = []string{
		".##.##.",
		"#*#####",
		"#######",
		".#####.",
		"..###..",
		"...#...",
	}
	heartPalette = map[byte]draw2d.Color{'#': draw2d.MustHex("#ff6f91"), '*': draw2d.MustHex("#ffd6e0")}
)

// The life bar's cells run from a light blue to a deep one; low, they are red.
var (
	lifeCells    = [2]draw2d.Color{draw2d.MustHex("#7fe8ff"), draw2d.MustHex("#3d8fe0")}
	lifeCellsLow = [2]draw2d.Color{draw2d.MustHex("#ff9aa8"), draw2d.MustHex("#e0405a")}
)

// hudColours are the readings' colours: the spent charges, the boxes and their shine.
var hudColours = struct {
	spent, box, edge, ink, shadow, shine draw2d.Color
}{
	spent: draw2d.MustHex("#1c2a48"), box: draw2d.MustHex("#0e1c42e0"), edge: draw2d.MustHex("#9cc8f2"),
	ink: draw2d.MustHex("#eaf4ff"), shadow: draw2d.MustHex("#02060fa0"), shine: draw2d.MustHex("#ffffff70"),
}

// The assault's readings at the top right: the mothership's life as a bar under its name, and
// the round's time left under it. Sizes are in pixels.
const (
	roundRight      = 10
	roundTop        = 10
	roundBarWidth   = 110
	roundBarHeight  = 8
	roundLabelRoom  = 12 // the name's line, above the bar
	roundTimeRoom   = 18 // the time's line, below the bar
	secondsAMinute  = 60
	roundLabelInset = 4
)

// roundColours are the mothership's bar: its life, warm as a target's.
var roundColours = [2]draw2d.Color{draw2d.MustHex("#ffd36f"), draw2d.MustHex("#e0602d")}

// Paint lays the readings out in drawing for this frame, which it empties first, over a view of
// width × height pixels.
func Paint(drawing *draw2d.List, state *game.Game, now float64, width, height float32) {
	drawing.Reset()
	if state.Round == nil || state.Round.Seat == nil { // a gunner's turret is not hit, only the hull
		paintLife(drawing, state.Player.Life, state.Settings.MaxLife, now)
	}
	if state.Round != nil {
		paintRound(drawing, state.Round, now, width)
	}
	switch {
	case state.Returning:
		paintCopilot(drawing, copilotTurningBack, now, width, height)
	case state.EdgeNearness() >= copilotWarnShare:
		paintCopilot(drawing, copilotEdgeWarning, now, width, height)
	}
}

// paintLife draws life as charges in a capsule with a heart: one cell for each tenth of the
// most life, lit as far as life reaches, the last one partly, each with a shine along its top.
// Low, the cells turn red and the heart beats.
func paintLife(drawing *draw2d.List, life, most, now float64) {
	if most <= 0 {
		return
	}
	segments := min(int(math.Ceil(most)), lifeMostSegments)
	share := life / most
	cells := lifeCells
	low := share < lifeLowShare
	if low {
		cells = lifeCellsLow
	}
	width := float32(lifeHeartRoom + segments*(lifeSegmentWidth+lifeSegmentGap) + lifeSegmentGap + 2)
	frame := draw2d.Rectangle{X: lifeLeft, Y: lifeTop, Width: width, Height: lifeSegmentHeight + 8}
	shadow := frame
	shadow.X, shadow.Y = shadow.X+hudShadowDrop, shadow.Y+hudShadowDrop
	pixelart.Stepped(drawing, shadow, 2, hudColours.shadow)
	pixelart.Stepped(drawing, frame, 2, hudColours.edge)
	pixelart.Stepped(drawing, pixelart.Inset(frame, 2), 2, hudColours.box)
	heartSize := float32(2)
	if low && pixelart.Shown(now, lifeBeatsASecond, heartBeatShare) {
		heartSize = 2.4
	}
	pixelart.PaintedBitmap(drawing, lifeHeart, frame.X+7, frame.Y+(frame.Height-6*heartSize)/2, heartSize, heartPalette)
	left := frame.X + lifeHeartRoom
	top := frame.Y + 4
	for segment := range segments {
		x := left + float32(segment*(lifeSegmentWidth+lifeSegmentGap))
		drawing.Rectangle(draw2d.Rectangle{X: x, Y: top, Width: lifeSegmentWidth, Height: lifeSegmentHeight}, 0, hudColours.spent, draw2d.Color{}, 0)
		filled := min(max(share*float64(segments)-float64(segment), 0), 1)
		if filled == 0 {
			continue
		}
		colour := mixColours(cells[0], cells[1], float32(segment)/float32(max(segments-1, 1)))
		fill := float32(filled) * lifeSegmentWidth
		drawing.Rectangle(draw2d.Rectangle{X: x, Y: top, Width: fill, Height: lifeSegmentHeight}, 0, colour, draw2d.Color{}, 0)
		drawing.Rectangle(draw2d.Rectangle{X: x, Y: top, Width: fill, Height: 2}, 0, hudColours.shine, draw2d.Color{}, 0)
	}
}

// mixColours is between first and second, share of the way.
func mixColours(first, second draw2d.Color, share float32) draw2d.Color {
	var mixed draw2d.Color
	for index := range mixed {
		mixed[index] = first[index] + (second[index]-first[index])*share
	}
	return mixed
}

// paintCopilot draws the copilot's face and words in a box along the bottom; the words blink.
func paintCopilot(drawing *draw2d.List, words string, now float64, width, height float32) {
	box := draw2d.Rectangle{X: copilotMargin, Y: height - copilotMargin - copilotHeight, Width: width - 2*copilotMargin, Height: copilotHeight}
	shadow := box
	shadow.X, shadow.Y = shadow.X+hudShadowDrop, shadow.Y+hudShadowDrop
	pixelart.Stepped(drawing, shadow, 2, hudColours.shadow)
	pixelart.Stepped(drawing, box, 2, hudColours.edge)
	pixelart.Stepped(drawing, pixelart.Inset(box, 2), 2, hudColours.box)
	faceTop := box.Y + (copilotHeight-float32(len(copilotFace))*copilotPixel)/2
	pixelart.PaintedBitmap(drawing, copilotFace, box.X+copilotFaceLeft, faceTop, copilotPixel, copilotPalette)
	if pixelart.Shown(now, copilotBlinks, copilotWordsShare) {
		drawing.Text(words, box.X+copilotWordsLeft, box.Y+copilotHeight/2+5, pixelart.TextSize, draw2d.Regular, hudColours.ink)
	}
}

// paintRound draws the mothership's life and the round's time left, at the top right.
func paintRound(drawing *draw2d.List, round *game.Round, now float64, width float32) {
	frame := draw2d.Rectangle{X: width - roundRight - roundBarWidth - 8, Y: roundTop, Width: roundBarWidth + 8, Height: roundLabelRoom + roundBarHeight + roundTimeRoom + 6}
	shadow := frame
	shadow.X, shadow.Y = shadow.X+hudShadowDrop, shadow.Y+hudShadowDrop
	pixelart.Stepped(drawing, shadow, 2, hudColours.shadow)
	pixelart.Stepped(drawing, frame, 2, hudColours.edge)
	pixelart.Stepped(drawing, pixelart.Inset(frame, 2), 2, hudColours.box)
	left := frame.X + roundLabelInset
	drawing.Text("MOTHERSHIP", left, frame.Y+roundLabelRoom, pixelart.LabelSize, draw2d.Mono, hudColours.ink)
	bar := draw2d.Rectangle{X: left, Y: frame.Y + roundLabelRoom + 3, Width: roundBarWidth, Height: roundBarHeight}
	drawing.Rectangle(bar, 0, hudColours.spent, draw2d.Color{}, 0)
	share := float32(0)
	if round.Mothership.MaxLife > 0 {
		share = float32(min(max(round.Mothership.Life/round.Mothership.MaxLife, 0), 1))
	}
	lit := bar
	lit.Width *= share
	drawing.Rectangle(lit, 0, mixColours(roundColours[1], roundColours[0], share), draw2d.Color{}, 0)
	drawing.Rectangle(draw2d.Rectangle{X: lit.X, Y: lit.Y, Width: lit.Width, Height: 2}, 0, hudColours.shine, draw2d.Color{}, 0)
	drawing.Text(clock(round.EndsAt-now)+" LEFT", left, bar.Y+roundBarHeight+roundTimeRoom-4, pixelart.TextSize, draw2d.Regular, hudColours.ink)
}

// clock writes seconds as minutes and seconds, 85 as 1:25; never below 0:00.
func clock(seconds float64) string {
	whole := max(int(math.Ceil(seconds)), 0)
	return strconv.Itoa(whole/secondsAMinute) + ":" + strconv.Itoa(whole%secondsAMinute/10) + strconv.Itoa(whole%10)
}
