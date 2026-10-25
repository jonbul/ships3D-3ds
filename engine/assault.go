package main

import (
	"math"
	"strconv"

	"github.com/jonbul/ships3D-3ds/game"
	"github.com/jonbul/ships3D-3ds/protocol"
)

// roleSeconds is how long a round's start tells the player their role.
const roleSeconds = 4

// roundHeading is what the top screen says over an assault round, as the browser does: our role
// at its start, and who won once it is over; shown is false in between.
func roundHeading(round *game.Round, now float64) (heading, caption string, shown bool) {
	if round.Winner != "" {
		return roundResult(round, now)
	}
	if now-round.StartedAt > roleSeconds {
		return "", "", false
	}
	if round.Team == protocol.TeamGunners {
		return "Gunner", "You are a gunner: defend the mothership", true
	}
	return "Attacker", "You are an attacker: destroy the mothership", true
}

// roundResult says whether we won, how, and when the lobby comes back.
func roundResult(round *game.Round, now float64) (heading, caption string, shown bool) {
	heading = "You lost"
	if round.Winner == round.Team {
		heading = "You won"
	}
	caption = "The mothership held out!"
	if round.Winner == protocol.TeamAttackers {
		caption = "The attackers destroyed the mothership!"
	}
	back := int(math.Max(math.Ceil(round.ReturnAt-now), 0))
	return heading, caption + "  ·  Lobby in " + strconv.Itoa(back) + " s", true
}
