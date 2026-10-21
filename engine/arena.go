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
	escaped := !hasTouchScreen && !pilot.menu.open && pilot.played.Game.Welcomed && cancelled(keys)
	return menuPressed(keys) || escaped
}

// pressJoining acts on the window shown until the server welcomes us; it tells whether the
// player went back to the ships.
func (pilot *Pilot) pressJoining(read touchscreen.Input) bool {
	read.Closes = true
	sound, closed := pilot.joining.Press([]touchscreen.Page{pilot.joiningPage()}, read, pilot.now)
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

// paintTouchScreen draws the touch screen: the menu, the window until we are welcomed, or the
// radar's readings over the 3D radar (or the flat map).
func (pilot *Pilot) paintTouchScreen(drawing *draw2d.List) {
	switch {
	case pilot.menu.open && pilot.played != nil:
		pilot.menu.paint(drawing, pilot.pages(), pilot.now)
	case pilot.played == nil || !pilot.played.Game.Welcomed:
		pilot.joining.Paint(drawing, []touchscreen.Page{pilot.joiningPage()}, "B SHIPS", pilot.now)
	default:
		shown := touchscreen.Radar{State: pilot.played.Game, Scanner: pilot.radar.scanner, Range: pilot.zoom, Flat: player.Settings.FlatRadar}
		touchscreen.PaintRadar(drawing, touchscreen.FullScreen, shown, pilot.now)
	}
}

// paintReadings paints over the 3D: the 3DS's life bar and the copilot, or the heading of a
// destroyed ship, laid out in the 3DS's pixels and larger on a larger screen. The desktop, with no
// touch screen, shows its pages in the middle, and the flat radar in a corner.
func (pilot *Pilot) paintReadings(drawing *draw2d.List, width, height float32) {
	paintScaled(drawing, width, height, func(drawing *draw2d.List, width, height float32) {
		welcomed := pilot.played != nil && pilot.played.Game.Welcomed
		if welcomed {
			pilot.paintFlight(drawing, width, height)
		}
		if hasTouchScreen {
			return
		}
		if welcomed && !pilot.menu.open {
			pilot.paintCornerRadar(drawing, width, height)
			return
		}
		paintPanel(drawing, width, height, pilot.paintTouchScreen)
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

// The desktop's radar: the touch screen's picture at this share of its size, in the bottom right
// corner, margin pixels of the top screen from its edges.
const (
	cornerRadarShare  = 0.5
	cornerRadarMargin = 6
)

// paintCornerRadar is the desktop's radar, as the browser has it in a corner: the 3DS's touch
// screen smaller, the 3D dish (the second screen's scene, placed there) or the flat map under its
// readings.
func (pilot *Pilot) paintCornerRadar(drawing *draw2d.List, width, height float32) {
	flat := player.Settings.FlatRadar
	x := width - touchscreen.Width*cornerRadarShare - cornerRadarMargin
	y := height - touchscreen.Height*cornerRadarShare - cornerRadarMargin
	pilot.placeCornerDish(x, y, height, flat)
	start := len(drawing.Commands)
	if flat {
		drawing.Rectangle(touchscreen.FullScreen, 0, touchscreen.Colours.Screen, draw2d.Color{}, 0)
	}
	shown := touchscreen.Radar{State: pilot.played.Game, Scanner: pilot.radar.scanner, Range: pilot.zoom, Flat: flat, MenuKey: "TAB"}
	touchscreen.PaintRadar(drawing, touchscreen.FullScreen, shown, pilot.now)
	pixelart.Scale(drawing, start, cornerRadarShare)
	pixelart.Move(drawing, start, x, y)
}

// placeCornerDish puts the 3D dish in the corner at x, y (in the top screen's pixels, as painted),
// in the window's pixels, which height (also as painted) tells the scale of; or nowhere with the
// flat map.
func (pilot *Pilot) placeCornerDish(x, y, height float32, flat bool) {
	if flat {
		engine.PlaceSecondScreen(nil)
		return
	}
	scale := windowHeight / height
	pilot.cornerArea = draw2d.Rect{X: x * scale, Y: y * scale, Width: touchscreen.Width * cornerRadarShare * scale, Height: touchscreen.Height * cornerRadarShare * scale}
	engine.PlaceSecondScreen(&pilot.cornerArea)
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
