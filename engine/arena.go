package main

import (
	"math"
	"strconv"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/hud"
	"github.com/jonbul/ships3D-3ds/pixelart"
	"github.com/jonbul/ships3D-3ds/radar"
	"github.com/jonbul/ships3D-3ds/touchscreen"
)

// The arena's screens: what the menus do with the keys, and what each screen shows.

// pressMenus acts on the menus' keys; it tells whether the player chose to leave the match.
func (pilot *Pilot) pressMenus(context *engine.Context) bool {
	keys := context.Input
	read := panelInput(keys)
	state := pilot.played.Game
	switch {
	case pilot.opensMenu(keys):
		pilot.menu.toggle()
	case pilot.menu.open:
		pilot.menu.press(pilot.pages(), read, pilot.now)
	case !state.Welcomed:
		return pilot.pressJoining(read)
	case read.Pressed&touchscreen.ButtonUp != 0:
		pilot.zoom = max(pilot.zoom-1, 0)
	case read.Pressed&touchscreen.ButtonDown != 0:
		pilot.zoom = min(pilot.zoom+1, len(radar.Ranges)-1)
	}
	return false
}

// opensMenu tells whether the player opened or closed the menu: START, and on the desktop also
// Escape while it is closed, as games there pause (open, Escape is B and closes it).
func (pilot *Pilot) opensMenu(keys *engine.Input) bool {
	escaped := oneScreen() && !pilot.menu.open && pilot.played.Game.Welcomed && cancelled(keys)
	return menuPressed(keys) || escaped
}

// pressJoining acts on the window shown until the server welcomes us; it tells whether the
// player went back to the ships.
func (pilot *Pilot) pressJoining(read touchscreen.Input) bool {
	read.Closes = true
	sound, closed := pilot.joining.Press([]touchscreen.Page{pilot.joiningPage()}, read, panel(), pilot.now)
	playMenuSound(sound)
	return closed
}

// pressWithoutMatch is the keys when the match couldn't start or has ended: B (Escape) goes back
// to the ships.
func (pilot *Pilot) pressWithoutMatch(context *engine.Context) {
	if pilot.pressJoining(panelInput(context.Input)) || pilot.leaving {
		pilot.leave(context)
	}
}

// pages are the game's menu: the pilot (with the way out of the game), the figures, the settings
// and the account.
func (pilot *Pilot) pages() []touchscreen.Page {
	state := pilot.played.Game
	motion := touchscreen.Motion{Roll: turning.Roll, Pitch: turning.Pitch, Yaw: turning.Yaw, TiltPitch: tilted.Pitch, TiltSteering: tilted.Steer}
	return []touchscreen.Page{
		touchscreen.PilotPage(state, touchscreen.Pilot{
			Resume:     pilot.menu.toggle,
			ChooseShip: func() { pilot.leaving = true },
			Quit:       engine.Quit,
		}),
		touchscreen.FiguresPage(pilot.figures.Shown, float64(state.Settings.TickRate), motion),
		settingsPage(),
		accountPage(func() { pilot.leaving = true }), // another server or player: join again
	}
}

// joiningPage is the window shown until the server welcomes us: how the connection goes, and the
// way back.
func (pilot *Pilot) joiningPage() touchscreen.Page {
	if pilot.played != nil && pilot.played.Game.Lobby != nil {
		return touchscreen.LobbyPage(*pilot.played.Game.Lobby, pilot.played.Game.Id, pilot.now, func() { pilot.leaving = true })
	}
	rows := touchscreen.StatusRows(pilot.connectionStatus(), pilot.problem())
	rows = append(rows, touchscreen.Row{}, touchscreen.Row{Label: "Back to the ships", Activate: func() { pilot.leaving = true }})
	return touchscreen.Page{Title: touchscreen.JoiningTitle, Small: true, Rows: rows}
}

// connectionStatus says how the connection to the server goes.
func (pilot *Pilot) connectionStatus() string {
	switch {
	case pilot.played == nil:
		return pilot.status
	case pilot.played.Game.Welcomed:
		return "Playing on " + player.Server
	}
	return "Joining…"
}

// problem is the server's complaint, if it made one.
func (pilot *Pilot) problem() string {
	if pilot.played == nil || pilot.played.Game.Problem == "" {
		return ""
	}
	return "Server: " + pilot.played.Game.Problem
}

// paintTouchScreen draws the 3DS's touch screen: the menu, the window until we are welcomed, or
// the radar's readings over the 3D radar (or the flat map).
func (pilot *Pilot) paintTouchScreen(drawing *draw2d.List) {
	if pilot.menuShown() {
		pilot.paintMenus(drawing, touchscreen.FullScreen)
		return
	}
	shown := touchscreen.Radar{State: pilot.played.Game, Scanner: pilot.radar.scanner, Range: pilot.zoom, Flat: player.Settings.FlatRadar}
	touchscreen.PaintRadar(drawing, touchscreen.FullScreen, shown, pilot.now)
}

// menuShown tells whether a window covers the game: the menu, or the one shown until we are
// welcomed.
func (pilot *Pilot) menuShown() bool {
	return pilot.played == nil || !pilot.played.Game.Welcomed || pilot.menu.open
}

// paintMenus draws the menu, or the window shown until we are welcomed, in screen.
func (pilot *Pilot) paintMenus(drawing *draw2d.List, screen draw2d.Rect) {
	if pilot.played != nil && pilot.menu.open {
		pilot.menu.paint(drawing, screen, pilot.pages(), pilot.now)
		return
	}
	ground(drawing, pilot.now)
	pilot.joining.Paint(drawing, screen, []touchscreen.Page{pilot.joiningPage()}, joiningHints.shown(), pilot.now)
}

// paintReadings paints over the 3D, laid out in the layout's pixels (paintScaled): the HUD, and on
// one screen what the 3DS's touch screen shows, laid out around the game (oneScreenReadings) or,
// while a window covers the game, that window in the middle.
func (pilot *Pilot) paintReadings(drawing *draw2d.List, width, height float32) {
	paintScaled(drawing, width, height, func(drawing *draw2d.List, width, height float32) {
		welcomed := pilot.played != nil && pilot.played.Game.Welcomed
		hudWidth := width
		if oneScreen() && welcomed {
			hudWidth = width - cornerRadarSide - 2*cornerRadarMargin // the copilot keeps clear of the radar
		}
		if welcomed {
			pilot.paintFlight(drawing, hudWidth, height)
		}
		if !oneScreen() {
			return
		}
		if pilot.menuShown() {
			engine.PlaceSecondScreen(nil) // the window covers the corner
			pilot.paintMenus(drawing, panelIn(width, height))
			return
		}
		pilot.paintOneScreenReadings(drawing, width, height)
	})
}

// paintFlight is the top screen's readings: the life bar and the copilot, or who destroyed us;
// over an assault round, also our role at its start and the result at its end.
func (pilot *Pilot) paintFlight(drawing *draw2d.List, width, height float32) {
	state := pilot.played.Game
	over := state.Round != nil && state.Round.Winner != ""
	switch {
	case over:
	case !state.Alive:
		topHeading(drawing, width, "Destroyed", pilot.deathCaption())
	default:
		hud.Paint(drawing, state, pilot.now, width, height)
	}
	if state.Round == nil {
		return
	}
	if heading, caption, shown := roundHeading(state.Round, pilot.now); shown && (over || state.Alive) {
		topHeading(drawing, width, heading, caption)
	}
}

// One screen's radar, as the browser's: a square in the bottom right corner, margin pixels from
// the window's edges, its range above it and the speed beside it, in the layout's pixels.
const (
	cornerRadarSide   = 150
	cornerRadarMargin = 8
	speedBarRoom      = 40 // the speed's bar and its number, left of the radar
	radarLabelAbove   = 4
	scoreTop          = 34 // under the life bar
	scoreLeft         = 10
)

// paintOneScreenReadings lays the touch screen's readings out around the game: the radar in the
// corner (the 3D sphere placed there, or the flat map), its range and the speed beside it, the
// score under the life bar, a hit's flash round the window, and the keys along the bottom.
func (pilot *Pilot) paintOneScreenReadings(drawing *draw2d.List, width, height float32) {
	state := pilot.played.Game
	corner := draw2d.Rect{X: width - cornerRadarSide - cornerRadarMargin, Y: height - cornerRadarSide - cornerRadarMargin, Width: cornerRadarSide, Height: cornerRadarSide}
	flat := player.Settings.FlatRadar
	pilot.placeCornerSphere(corner, height, flat)
	if flat {
		drawing.Rectangle(corner, cornerRadarSide/2, touchscreen.Colours.Screen, draw2d.Color{}, 0)
		shown := touchscreen.Radar{State: state, Scanner: pilot.radar.scanner, Range: pilot.zoom, Flat: true}
		touchscreen.PaintMap(drawing, corner, shown, pilot.now)
	}
	label := "RADAR  " + touchscreen.ZoomText(pilot.zoom)
	drawing.Text(label, corner.X, corner.Y-radarLabelAbove, pixelart.LabelSize, draw2d.Mono, colours.soft)
	speed := draw2d.Rect{X: corner.X - speedBarRoom, Y: corner.Y, Width: speedBarRoom, Height: corner.Height}
	touchscreen.PaintSpeed(drawing, speed, state)
	touchscreen.PaintScore(drawing, scoreLeft, scoreTop, state, pilot.now)
	touchscreen.PaintHit(drawing, draw2d.Rect{Width: width, Height: height}, state, pilot.now)
	hints(drawing, matchHints.shown(), width, height)
}

// placeCornerSphere puts the 3D radar's sphere in the corner (in the layout's pixels, a screen
// height of them high), in the window's pixels; or nowhere with the flat map.
func (pilot *Pilot) placeCornerSphere(corner draw2d.Rect, height float32, flat bool) {
	if flat {
		engine.PlaceSecondScreen(nil)
		return
	}
	scale := windowHeight / height
	pilot.cornerArea = draw2d.Rect{X: corner.X * scale, Y: corner.Y * scale, Width: corner.Width * scale, Height: corner.Height * scale}
	engine.PlaceSecondScreenOver(&pilot.cornerArea) // floating over the game, as the browser's
}

// deathCaption says who destroyed us, and when we may fly again.
func (pilot *Pilot) deathCaption() string {
	state := pilot.played.Game
	caption := ""
	if killer := state.KilledBy(); killer != "" {
		caption = "by " + killer + "  ·  "
	}
	if state.CanRespawn(pilot.now) {
		return caption + "A to fly again"
	}
	return caption + "back in " + strconv.Itoa(int(math.Ceil(state.RespawnIn(pilot.now)))) + " s"
}
