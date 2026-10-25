package main

import (
	"strings"
	"testing"

	"github.com/jonbul/ships3D-3ds/game"
	"github.com/jonbul/ships3D-3ds/protocol"
)

func TestARoundFirstTellsOurRoleThenNothingThenTheResult(test *testing.T) {
	round := &game.Round{Team: protocol.TeamAttackers, StartedAt: 10, EndsAt: 100}
	if heading, _, shown := roundHeading(round, 11); !shown || heading != "Attacker" {
		test.Errorf("at the start the heading is %q (shown %v), want our role", heading, shown)
	}
	if heading, _, shown := roundHeading(round, 30); shown {
		test.Errorf("in the middle of the round the heading is %q", heading)
	}
	round.Winner, round.ReturnAt = protocol.TeamGunners, 40
	heading, caption, shown := roundHeading(round, 35)
	if !shown || heading != "You lost" || !strings.Contains(caption, "held out") || !strings.Contains(caption, "5 s") {
		test.Errorf("after the round: %q, %q (shown %v)", heading, caption, shown)
	}
}
