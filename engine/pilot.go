package main

import (
	"log/slog"
	"math"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/camera"
	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/arena"
	"github.com/jonbul/ships3D-3ds/chase"
	"github.com/jonbul/ships3D-3ds/figures"
	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/match"
	"github.com/jonbul/ships3D-3ds/protocol"
	"github.com/jonbul/ships3D-3ds/radar"
	"github.com/jonbul/ships3D-3ds/sounds"
	"github.com/jonbul/ships3D-3ds/touchscreen"
	"github.com/jonbul/ships3D-3ds/transport"
)

// Pilot plays a match on the server: it joins with a ship, flies it with the player's keys, and
// shows the arena (package arena) inside the engine's scene, its camera chasing the ship. On the
// 3DS the touch screen shows the radar and, with START, the menu, as the hand-written program
// does; on the desktop the radar is in a corner and the menu's pages in the middle (panel.go).
type Pilot struct {
	world      *arena.World
	radar      *radarScreen
	sound      *soundSet
	played     *match.Match
	view       camera.Lens
	status     string
	now        float64 // seconds, this frame
	menu       touchMenu
	joining    touchscreen.Menu // the window while the server hasn't welcomed us yet
	zoom       int              // the radar's reach, an index in radar.Ranges
	figures    figures.Counter
	logged     float64          // seconds, when the figures were last written to the log
	leaving    bool             // a choice of the menu asked to go back to the ships
	cornerArea draw2d.Rectangle // where the desktop shows the 3D radar, in the window's pixels
}

// The real server, and what a guest flies when the scene names nothing.
const (
	realServer  = "https://jonbul.ddns.net:3000"
	defaultShip = "default:arrow"
	guestName   = "Pilot"
)

// Start builds the arena in the engine's scene and starts the match.
func (pilot *Pilot) Start(context *engine.Context) {
	pilot.world = stage(context)
	drawInterface(askedAccountForm)
	pilot.radar = openRadar()
	pilot.zoom = radar.DefaultRange
	drawOver(pilot.paintReadings)
	if !oneScreen() {
		showTouchScreen(pilot.paintTouchScreen)
	}
	pilot.view = camera.Lens{Far: arena.CameraFar, Near: arena.CameraNear, FieldOfViewRadians: chase.FieldOfView}
	pilot.sound = gameSound
	pilot.status = "Connecting to " + player.Server + "…"
	played, err := match.Start(transport.Dial, player.Server, chosen.mode, matchCookie(), orDefault(chosen.shipID, defaultShip), playerName(), pilot.hooks())
	if err != nil {
		pilot.status = err.Error()
		slog.Warn("the match could not start", "error", err)
		return
	}
	pilot.played = played
}

// hooks are what the match tells the game as it goes: the shots' sounds, and the figures.
func (pilot *Pilot) hooks() match.Hooks {
	return match.Hooks{
		Fired: func() { pilot.sound.Play(pilot.sound.Laser, 1) },
		Received: func(message protocol.Incoming, bytes int) {
			pilot.figures.Received(bytes)
			if message.Type == protocol.TypeFire {
				pilot.sound.Play(pilot.sound.Laser, sounds.OtherShotsVolume)
			}
		},
		Sent: func(message string) { pilot.figures.Sent(len(message)) },
	}
}

// matchCookie is the player's session for the match, or none for a guest.
func matchCookie() string {
	if player.LoggedIn() {
		return session().Cookie()
	}
	return ""
}

// Update moves the match on a frame and shows it, through the scene's camera.
func (pilot *Pilot) Update(context *engine.Context) {
	pilot.now = context.Time
	darkenSky(context.Scene)
	if pilot.played == nil {
		pilot.pressWithoutMatch(context)
		return
	}
	if pilot.pressMenus(context) || pilot.leaving {
		pilot.leave(context)
		return
	}
	if err := pilot.played.Exchange(pilot.controls(context)); err != nil {
		pilot.status = "Disconnected: " + err.Error()
		slog.Warn("the match ended", "error", err)
		pilot.played.Close()
		pilot.played = nil
		return
	}
	pilot.show(context)
}

// controls are this frame's controls and whether the ship fires, with its seconds and time; the
// ship flies on straight while a menu is open.
func (pilot *Pilot) controls(context *engine.Context) (flight.Controls, bool, float64, float64) {
	turnTilt(context.Input, context.Delta)
	if pilot.menu.open {
		return flight.Controls{}, false, context.Delta, context.Time
	}
	controls, firing := controlsOf(context.Input, player.Settings)
	return controls, firing, context.Delta, context.Time
}

// show plays the frame's sounds and shows the arena, the radar and the figures.
func (pilot *Pilot) show(context *engine.Context) {
	state := pilot.played.Game
	events := state.TakeEvents()
	pilot.sound.Hear(events, state)
	flying := state.Welcomed && state.Alive
	pilot.sound.RunEngine(flying, math.Abs(state.Own.Speed)/math.Max(state.Settings.MaxSpeed, 1))
	pilot.world.Show(state, events, &pilot.view, context.Delta, context.Time)
	pilot.radar.show(state, context.Time, pilot.zoom, player.Settings.FlatRadar)
	look(context, pilot.view)
	delay := 0.0
	if state.SnapshotAt > 0 {
		delay = context.Time - state.SnapshotAt
	}
	pilot.figures.AddFrame(figures.Frame{Seconds: context.Delta, Players: len(state.Remotes) + 1, SnapshotDelay: delay})
	pilot.logFigures(context.Time)
}

// figuresLogSeconds is how often a match writes its figures to the log, as the 3DS program does:
// a console's log is how they are read without looking at its screen.
const figuresLogSeconds = 10

// logFigures writes the last second's figures to the log, every figuresLogSeconds.
func (pilot *Pilot) logFigures(now float64) {
	if now-pilot.logged < figuresLogSeconds {
		return
	}
	pilot.logged = now
	shown := pilot.figures.Shown
	slog.Info("figures", "fps", shown.FramesPerSecond, "slowest frame ms", int(shown.SlowestFrameMilliseconds),
		"processor ms", int(shown.ProcessorMilliseconds), "players", shown.Players, "in B/s", shown.BytesIn, "out B/s", shown.BytesOut)
}

// leave ends the match and goes back to the ships.
func (pilot *Pilot) leave(context *engine.Context) {
	if pilot.played != nil {
		pilot.played.Close()
		pilot.played = nil
	}
	pilot.radar.close()
	saveSettings()
	playSound(func(sound *soundSet) {
		sound.Play(sound.Back, 1)
		sound.RunEngine(false, 0)
	})
	context.ChangeScene(shipScene)
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// playerName is the logged-in player's name, or a guest's.
func playerName() string {
	if player.LoggedIn() {
		return player.Username
	}
	return guestName
}
