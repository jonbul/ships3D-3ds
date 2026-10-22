package touchscreen

import (
	"math"
	"strconv"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/game"
	"github.com/jonbul/ships3D-3ds/radar"
)

// The map's look, in pixels of the touch screen.
const (
	mapMargin         = 14
	blipSize          = 8
	ownMarkerSize     = 10
	sweepDots         = 40
	speedBarWidth     = 6
	speedBarMargin    = 40
	hitFlashSeconds   = 0.5
	hitFlashWidth     = 6
	sweepDotSize      = 2
	sweepTrailRadians = 0.06
	stalkWidth        = 2
	ringWidth         = 1.5
	heightShare       = 0.6 // how much of the map's radius a ship at the edge, straight above, rises
	labelSize         = 13
	labelMargin       = 8
	labelLineSpacing  = 16
)

var (
	mapRing         = draw2d.Hex("#2c4a5a")
	mapRingFaint    = draw2d.Hex("#1d323d")
	mapOwn          = draw2d.Hex("#4fc3f7")
	mapEnemy        = draw2d.Hex("#ff5252")
	mapEnemyFar     = draw2d.Hex("#8a3434")
	mapSweep        = draw2d.Hex("#4fc3f7")
	mapHit          = draw2d.Hex("#ff5252")
	sweepTrail      = []float32{0.9, 0.45, 0.2}
	mapLabel        = draw2d.Hex("#9fb3c0")
	mapTransparent  = draw2d.Color{}
	mapCircleRadius = float32(1 << 10) // rounding larger than the shape makes a circle
)

// Radar is what the touch screen's radar shows: the game, what the sweep has seen, how far it
// reaches and how it is drawn.
type Radar struct {
	State   *game.Game
	Scanner *radar.Scanner
	Range   int    // the index of its reach in radar.Ranges
	Flat    bool   // a flat map drawn here; otherwise the 3D radar is drawn under, by a scene of its own
	MenuKey string // the key the map says opens the menu; none is the 3DS's START
}

// PaintRadar draws the radar over area at now seconds: the flat map, when it is flat, and the
// readings around it either way.
func PaintRadar(drawing *draw2d.List, area draw2d.Rect, shown Radar, now float64) {
	if shown.Flat {
		PaintMap(drawing, area, shown, now)
	}
	paintReadings(drawing, area, shown, now) // the 3D radar is drawn under, by the radar scene
}

// PaintMap draws the flat map in area at now seconds: its rings, the sweep, the ships it saw with
// their stalks, and our ship in the middle.
func PaintMap(drawing *draw2d.List, area draw2d.Rect, shown Radar, now float64) {
	state := shown.State
	radius := min(area.Width, area.Height)/2 - mapMargin
	centreX, centreY := area.X+area.Width/2, area.Y+area.Height/2
	circle := func(x, y, size float32, fill, border draw2d.Color, borderWidth float32) {
		drawing.Rectangle(draw2d.Rect{X: x - size/2, Y: y - size/2, Width: size, Height: size}, mapCircleRadius, fill, border, borderWidth)
	}
	circle(centreX, centreY, radius*2, mapTransparent, mapRing, ringWidth)
	circle(centreX, centreY, radius, mapTransparent, mapRingFaint, ringWidth)
	drawing.Rectangle(draw2d.Rect{X: centreX - ringWidth/2, Y: centreY - radius, Width: ringWidth, Height: radius * 2}, 0, mapRingFaint, mapTransparent, 0)
	drawing.Rectangle(draw2d.Rect{X: centreX - radius, Y: centreY - ringWidth/2, Width: radius * 2, Height: ringWidth}, 0, mapRingFaint, mapTransparent, 0)

	toScreen := func(blip radar.Blip) (x, y, top float32) {
		x = centreX + float32(blip.Right)*radius
		y = centreY - float32(blip.Ahead)*radius
		return x, y, y - float32(blip.Above)*radius*heightShare
	}
	paintSweep(drawing, centreX, centreY, radius, now)
	for _, echo := range shown.Scanner.Echoes {
		x, y, top := toScreen(echo.Blip)
		colour := mapEnemy
		if !echo.Blip.InRange {
			colour = mapEnemyFar
		}
		colour[3] *= float32(echo.Brightness(now))
		// A stalk from the map's plane up (or down) to the ship, as in the browser: above and
		// below read at a glance.
		drawing.Rectangle(draw2d.Rect{X: x - stalkWidth/2, Y: min(y, top), Width: stalkWidth, Height: abs(top - y)}, 0, colour, mapTransparent, 0)
		circle(x, top, blipSize, colour, mapTransparent, 0)
	}
	if state.Alive {
		circle(centreX, centreY, ownMarkerSize, mapOwn, mapTransparent, 0)
	}
}

// paintReadings draws what both radars show around the dish: the speed, a hit's flash, life,
// score, the range and how to open the options.
func paintReadings(drawing *draw2d.List, area draw2d.Rect, shown Radar, now float64) {
	state := shown.State
	PaintSpeed(drawing, area, state)
	PaintHit(drawing, area, state, now)
	PaintScore(drawing, area.X+labelMargin, area.Y+labelMargin, state, now)
	rangeText := ZoomText(shown.Range) + " UP/DOWN"
	drawing.Text(rangeText, area.X+labelMargin, area.Y+area.Height-labelMargin, labelPixels, draw2d.Mono, mapLabel)
	menuKey := shown.MenuKey
	if menuKey == "" {
		menuKey = "START"
	}
	menuText := menuKey + " MENU"
	drawing.Text(menuText, area.X+area.Width-labelMargin-float32(len(menuText))*labelPixels, area.Y+area.Height-labelMargin, labelPixels, draw2d.Mono, mapLabel)
}

// PaintHit flashes area's edge red, fading, just after we were hit.
func PaintHit(drawing *draw2d.List, area draw2d.Rect, state *game.Game, now float64) {
	if now-state.HitAt >= hitFlashSeconds || state.HitAt <= 0 {
		return
	}
	flash := mapHit
	flash[3] *= float32(1 - (now-state.HitAt)/hitFlashSeconds)
	drawing.Rectangle(area, 0, mapTransparent, flash, hitFlashWidth)
}

// PaintScore writes our kills and deaths, and how long until we fly again once destroyed, from
// x, y down; life is with the HUD (package hud).
func PaintScore(drawing *draw2d.List, x, y float32, state *game.Game, now float64) {
	lines := [2]string{strconv.Itoa(state.Player.Kills) + " kills, " + strconv.Itoa(state.Player.Deaths) + " deaths"}
	count := 1
	if !state.Alive {
		lines[1], count = respawnLine(state, now), 2
	}
	for index, line := range lines[:count] {
		drawing.Text(line, x, y+labelSize+float32(index)*labelLineSpacing, labelSize, draw2d.Regular, mapLabel)
	}
}

// ZoomText says how near the radar shows, by the index of its reach in radar.Ranges: the zoom
// rather than the distance to the edge, as the game's distances have no unit a player knows. The
// widest range is ×1.
func ZoomText(rangeIndex int) string {
	zoom := radar.Ranges[len(radar.Ranges)-1] / radar.Ranges[rangeIndex]
	return "ZOOM x" + strconv.Itoa(int(zoom))
}

// paintSweep draws the radar's turning line as a row of dots, since the draw list has no lines,
// with a fainter one just behind it.
func paintSweep(drawing *draw2d.List, centreX, centreY, radius float32, now float64) {
	for trail, brightness := range sweepTrail {
		angle := radar.Angle(now) - float64(trail)*sweepTrailRadians
		right, ahead := float32(math.Sin(angle)), float32(math.Cos(angle))
		colour := mapSweep
		colour[3] *= brightness
		for dot := 1; dot <= sweepDots; dot++ {
			share := float32(dot) / sweepDots
			drawing.Rectangle(draw2d.Rect{X: centreX + right*radius*share - sweepDotSize/2, Y: centreY - ahead*radius*share - sweepDotSize/2, Width: sweepDotSize, Height: sweepDotSize}, 0, colour, mapTransparent, 0)
		}
	}
}

// PaintSpeed draws our speed as a bar up area's right edge: empty at the slowest reverse, full
// at the fastest, with a mark where the ship stands still.
func PaintSpeed(drawing *draw2d.List, area draw2d.Rect, state *game.Game) {
	speed, slowest, fastest := state.Own.Speed, state.Settings.MinSpeed, state.Settings.MaxSpeed
	if fastest <= slowest {
		return
	}
	x := area.X + area.Width - labelMargin - speedBarWidth
	top, bottom := area.Y+speedBarMargin, area.Y+area.Height-speedBarMargin
	height := bottom - top
	share := float32((speed - slowest) / (fastest - slowest))
	stopped := bottom - float32(-slowest/(fastest-slowest))*height
	drawing.Rectangle(draw2d.Rect{X: x, Y: top, Width: speedBarWidth, Height: height}, speedBarWidth/2, mapRingFaint, mapTransparent, 0)
	drawing.Rectangle(draw2d.Rect{X: x, Y: bottom - share*height, Width: speedBarWidth, Height: share * height}, speedBarWidth/2, mapOwn, mapTransparent, 0)
	drawing.Rectangle(draw2d.Rect{X: x - 3, Y: stopped - ringWidth/2, Width: speedBarWidth + 6, Height: ringWidth}, 0, mapLabel, mapTransparent, 0)
	drawing.Text(strconv.Itoa(int(speed)), x-labelSize-14, bottom-share*height+labelSize/2, labelSize, draw2d.Regular, mapLabel)
}

// respawnLine says how long until the player may fly again, then how.
func respawnLine(state *game.Game, now float64) string {
	if state.CanRespawn(now) {
		return "Destroyed: A to fly again"
	}
	return "Destroyed: back in " + strconv.Itoa(int(math.Ceil(state.RespawnIn(now)))) + " s"
}
