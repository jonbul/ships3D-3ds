package hud

import (
	"strings"
	"testing"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/game"
	"github.com/jonbul/ships3D-3ds/protocol"
)

func TestTheLifeBarIsDrawnAndTheCopilotOnlyNearTheEdge(test *testing.T) {
	state := game.New()
	state.Settings.MaxLife, state.Settings.WorldRadius = 10, 1500
	state.Player.Life = 5
	var middle, edge draw2d.List
	Paint(&middle, state, 0, 400, 240)
	if middle.IsEmpty() {
		test.Fatal("nothing drawn: the life bar is missing")
	}
	state.Own.Position = flight.Vector{0, 0, 1490}
	Paint(&edge, state, 0, 400, 240)
	if len(edge.Commands) <= len(middle.Commands) {
		test.Fatalf("%d commands at the edge, %d in the middle: the copilot is missing", len(edge.Commands), len(middle.Commands))
	}
}

// texts are the lines a list writes.
func texts(drawing draw2d.List) []string {
	lines := []string{}
	for _, command := range drawing.Commands {
		if command.Kind == draw2d.TextKind {
			lines = append(lines, command.Text)
		}
	}
	return lines
}

func TestAnAssaultShowsTheMothershipsLifeAndTheTimeLeft(test *testing.T) {
	state := game.New()
	state.Settings.MaxLife, state.Settings.WorldRadius = 10, 1500
	state.Player.Life = 10
	state.Round = &game.Round{Team: protocol.TeamAttackers, EndsAt: 95}
	state.Round.Mothership.Life, state.Round.Mothership.MaxLife = 50, 100
	var attacker draw2d.List
	Paint(&attacker, state, 10, 400, 240)
	if lines := strings.Join(texts(attacker), "|"); !strings.Contains(lines, "MOTHERSHIP") || !strings.Contains(lines, "1:25") {
		test.Errorf("the attacker's readings say %q, want the mothership and 1:25 left", lines)
	}
	state.Round.Team, state.Round.Seat = protocol.TeamGunners, &protocol.Seat{}
	var gunner draw2d.List
	Paint(&gunner, state, 10, 400, 240)
	if len(gunner.Commands) >= len(attacker.Commands) {
		test.Errorf("the gunner's readings have %d commands, the attacker's %d: the gunner has no life bar", len(gunner.Commands), len(attacker.Commands))
	}
}
