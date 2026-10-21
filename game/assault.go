package game

import "github.com/jonbul/ships3D-3ds/protocol"

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
	EndsAt     float64 // seconds
	Winner     string  // a team, once the round is over
	ReturnAt   float64 // seconds, when the lobby comes back after the result
}

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
	game.Round = &Round{Rules: message.Rules, Team: game.Player.Team, Seat: game.Player.Seat, Mothership: message.Mothership, EndsAt: now + message.EndsInMs/millisecondsPerSecond}
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
