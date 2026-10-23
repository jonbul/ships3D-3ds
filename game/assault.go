package game

import (
	"math"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/hull"
	"github.com/jonbul/ships3D-3ds/protocol"
)

// The assault goes round a loop: the lobby while players gather, a round with two teams (the
// attackers against the gunners in the mothership's turrets), the result, and back to the lobby.

// Lobby is the assault's waiting room: who waits, how many may play, and when the round starts.
type Lobby struct {
	Players                []protocol.Player // only their ids, names and whether they are guests
	MinPlayers, MaxPlayers int
	CountingDown           bool
	StartsAt               float64 // seconds, while it counts down
}

// Round is an assault round: our team (and our turret, as a gunner), the mothership, when it
// ends, and once it is over who won and when the lobby comes back.
type Round struct {
	Rules      protocol.Rules
	Team       string
	Seat       *protocol.Seat // ours, as a gunner
	Mothership protocol.Mothership
	StartedAt  float64 // seconds, for the message that tells our role
	EndsAt     float64 // seconds
	Winner     string  // a team, once the round is over
	ReturnAt   float64 // seconds, when the lobby comes back after the result
	// Aim is where our turret points, as a gunner.
	Aim   Aim
	shape hull.Hull // the mothership's, at the middle of the world
}

// Aim is where a turret points: turned about the up axis (positive is to the left) and raised
// above level, in radians.
type Aim struct {
	Yaw, Pitch float64
}

// highestAim is the furthest a turret raises or lowers its barrels, in radians from level: short
// of straight up, as in the browser.
const highestAim = 1.4

// handleAssault takes the assault's own messages; it tells whether the message was one.
func (game *Game) handleAssault(message protocol.Incoming, now float64) bool {
	switch message.Type {
	case protocol.TypeLobby:
		game.enterLobby(message, now)
	case protocol.TypeRoundStart:
		game.startRound(message, now)
	case protocol.TypeRoundOver:
		if game.Round != nil {
			game.Round.Winner = message.Winner
			game.Round.ReturnAt = now + message.ReturnInMs/millisecondsPerSecond
		}
	default:
		return false
	}
	return true
}

// enterLobby waits in the lobby; coming back from a round, everything in it goes.
func (game *Game) enterLobby(message protocol.Incoming, now float64) {
	if game.Round != nil {
		game.leaveRound()
	}
	game.Id = message.Id
	game.Lobby = &Lobby{Players: message.Players, MinPlayers: message.MinPlayers, MaxPlayers: message.MaxPlayers}
	if message.StartsInMs != nil {
		game.Lobby.CountingDown = true
		game.Lobby.StartsAt = now + *message.StartsInMs/millisecondsPerSecond
	}
}

// startRound leaves the lobby for a round: it welcomes us into it, as the free for all does.
func (game *Game) startRound(message protocol.Incoming, now float64) {
	game.Lobby = nil
	game.welcome(message, now)
	game.Round = &Round{
		Rules: message.Rules, Team: game.Player.Team, Seat: game.Player.Seat,
		Mothership: message.Mothership, StartedAt: now, EndsAt: now + message.EndsInMs/millisecondsPerSecond,
		shape: hull.New(message.Mothership.Ship.Layers, message.Mothership.Scale),
	}
	if seat := game.Round.Seat; seat != nil {
		game.Own.Position = seat.Position
		game.Round.Aim = aimOf(game.Own.Rotation)
	}
}

// leaveRound forgets a finished round: its ships, its bullets and its scores.
func (game *Game) leaveRound() {
	game.Round = nil
	game.Welcomed, game.Alive = false, false
	game.Player = protocol.Player{}
	clear(game.Remotes)
	game.Bullets, game.Others = game.Bullets[:0], game.Others[:0]
}

// damageMothership takes the mothership's new life, from a damage message aimed at it.
func (game *Game) damageMothership(life float64) {
	if game.Round != nil {
		game.Round.Mothership.Life = life
	}
}

// gunning tells whether we sit in a turret.
func (game *Game) gunning() bool { return game.Round != nil && game.Round.Seat != nil }

// attacking tells whether we fly against the mothership.
func (game *Game) attacking() bool {
	return game.Round != nil && game.Round.Team == protocol.TeamAttackers
}

// canHit tells whether our bullets can hit another player: anyone in a free for all, and only
// attackers for a gunner (attackers only hit the mothership).
func (game *Game) canHit(remote *Remote) bool {
	return game.Round == nil || game.gunning() && remote.Player.Team == protocol.TeamAttackers
}

// cooldownSeconds is how long a gun reloads: the turrets' own, or the settings' (the attackers'
// in a round).
func (game *Game) cooldownSeconds() float64 {
	if game.gunning() && game.Round.Rules.GunnerCooldownMs > 0 {
		return float64(game.Round.Rules.GunnerCooldownMs) / millisecondsPerSecond
	}
	return float64(game.Settings.FireCooldownMs) / millisecondsPerSecond
}

// aimTurret turns our turret with the stick, at the turrets' rate at full deflection; it stays in
// its seat.
func (game *Game) aimTurret(controls flight.Controls, seconds float64) {
	aim := &game.Round.Aim
	rate := game.Round.Rules.GunnerTurnRate
	aim.Yaw -= controls.Yaw * rate * seconds // a turn right is a negative angle about up
	aim.Pitch = max(-highestAim, min(highestAim, aim.Pitch+controls.Pitch*rate*seconds))
	game.Own.Position, game.Own.Speed = game.Round.Seat.Position, 0
	game.Own.Rotation = rotationOf(*aim)
}

// rotationOf is the way an aim faces: turned about up, then raised (raising the nose is a
// negative turn about the side axis, as in package flight).
func rotationOf(aim Aim) flight.Rotation {
	return flight.FromEulerXYZ(0, aim.Yaw, 0).Then(flight.FromEulerXYZ(-aim.Pitch, 0, 0))
}

// aimOf is the aim that faces the way a rotation does.
func aimOf(rotation flight.Rotation) Aim {
	facing := rotation.Apply(flight.Forward)
	return Aim{Yaw: math.Atan2(facing[0], facing[2]), Pitch: max(-highestAim, min(highestAim, math.Asin(max(-1, min(1, facing[1])))))}
}
