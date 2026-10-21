package game

import (
	"math"
	"strings"
	"testing"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/protocol"
)

// startsIn is a lobby's countdown, as the server sends it.
func startsIn(milliseconds float64) *float64 { return &milliseconds }

// wall is the test's mothership: one part, a wall 40 wide and high and 1 deep (2 in the world),
// 50 ahead of the origin.
var wall = []protocol.Layer{{Visible: true, Parts: []protocol.Part{{Type: "box", Position: flight.Vector{0, 0, 25}, Scale: flight.Vector{20, 20, 0.5}}}}}

// inRound is a game in an assault round: we are "me", an attacker at the origin facing the
// mothership; "them" is a gunner in a turret.
func inRound(now float64) *Game {
	return inRoundAs(protocol.TeamAttackers, nil, now)
}

// inRoundAs is inRound with us in team, in seat when a gunner; "them" plays the other team.
func inRoundAs(team string, seat *protocol.Seat, now float64) *Game {
	other, otherSeat := protocol.TeamGunners, &protocol.Seat{Position: flight.Vector{0, 5, 0}, Up: 1}
	if team == protocol.TeamGunners {
		other, otherSeat = protocol.TeamAttackers, nil
	}
	game := New()
	game.Handle(protocol.Incoming{Type: protocol.TypeLobby, Id: "me", MinPlayers: 2, MaxPlayers: 8}, nil, 0)
	game.Handle(protocol.Incoming{
		Type: protocol.TypeRoundStart, Id: "me", Settings: settings, EndsInMs: 180000,
		Rules:      protocol.Rules{MothershipLife: 200, RoundMs: 180000},
		Mothership: protocol.Mothership{Life: 200, MaxLife: 200, Scale: 2, Ship: protocol.Ship{Layers: wall}},
		Players: []protocol.Player{
			{Id: "me", Rotation: flight.NoRotation, Life: 10, Team: team, Seat: seat},
			{Id: "them", Position: flight.Vector{0, 0, 20}, Rotation: flight.NoRotation, Life: 10, Team: other, Seat: otherSeat},
		},
	}, nil, now)
	return game
}

func TestTheLobbyListsWhoWaitsAndWhenItStarts(t *testing.T) {
	game := New()
	game.Handle(protocol.Incoming{Type: protocol.TypeLobby, Id: "me", MinPlayers: 2, MaxPlayers: 8,
		Players: []protocol.Player{{Id: "me", Name: "Me"}, {Id: "them", Name: "Them"}}, StartsInMs: startsIn(5000)}, nil, 10)
	lobby := game.Lobby
	if lobby == nil || len(lobby.Players) != 2 || lobby.MinPlayers != 2 || lobby.MaxPlayers != 8 {
		t.Fatalf("lobby %+v", lobby)
	}
	if !lobby.CountingDown || lobby.StartsAt != 15 {
		t.Fatalf("counting down %v, starts at %v, want 15 s", lobby.CountingDown, lobby.StartsAt)
	}
	if game.Welcomed || game.Id != "me" {
		t.Fatalf("welcomed %v as %q in the lobby", game.Welcomed, game.Id)
	}
}

func TestARoundStartPlacesUsInOurTeam(t *testing.T) {
	game := inRound(20)
	round := game.Round
	if game.Lobby != nil || round == nil || !game.Welcomed || !game.Alive {
		t.Fatalf("lobby %v, round %v, welcomed %v, alive %v", game.Lobby, round, game.Welcomed, game.Alive)
	}
	if round.Team != protocol.TeamAttackers || round.Seat != nil || round.EndsAt != 200 || round.Mothership.Life != 200 {
		t.Fatalf("round %+v", round)
	}
	if game.Remotes["them"] == nil || game.Remotes["them"].Player.Seat == nil {
		t.Fatalf("the gunner is %+v", game.Remotes["them"])
	}
}

func TestHittingTheMothershipLowersItsLife(t *testing.T) {
	game := inRound(0)
	game.Handle(protocol.Incoming{Type: protocol.TypeDamage, Target: protocol.MothershipTarget, From: "me", Life: 190}, nil, 1)
	if game.Round.Mothership.Life != 190 {
		t.Fatalf("the mothership's life is %v, want 190", game.Round.Mothership.Life)
	}
}

func TestTheRoundOverSaysWhoWonUntilTheLobbyComesBack(t *testing.T) {
	game := inRound(0)
	game.Handle(protocol.Incoming{Type: protocol.TypeRoundOver, Winner: protocol.TeamGunners, ReturnInMs: 8000}, nil, 100)
	if game.Round.Winner != protocol.TeamGunners || game.Round.ReturnAt != 108 {
		t.Fatalf("round %+v", game.Round)
	}
	game.Handle(protocol.Incoming{Type: protocol.TypeLobby, Id: "me", MinPlayers: 2, MaxPlayers: 8}, nil, 108)
	if game.Round != nil || game.Lobby == nil || game.Welcomed || len(game.Remotes) != 0 || len(game.Bullets) != 0 {
		t.Fatalf("back in the lobby: round %v, lobby %v, welcomed %v, %d remotes", game.Round, game.Lobby, game.Welcomed, len(game.Remotes))
	}
}

func TestAFreeForAllHasNoLobbyNorRound(t *testing.T) {
	game := welcomed(50)
	if game.Lobby != nil || game.Round != nil {
		t.Fatalf("lobby %v, round %v", game.Lobby, game.Round)
	}
}

func TestAnAttackersBulletHitsTheMothership(t *testing.T) {
	game := inRound(0)
	hits := 0
	for frame := range 60 {
		for _, message := range game.Step(flight.Controls{}, frame == 0, false, 1.0/60, float64(frame)/60) {
			if strings.Contains(message, `"type":"hit"`) {
				if !strings.Contains(message, `"target":"mothership"`) {
					t.Fatalf("the bullet hit %s", message)
				}
				hits++
			}
		}
	}
	if hits != 1 {
		t.Fatalf("%d hits on the mothership, want 1", hits)
	}
}

func TestAnAttackerFlyingIntoTheMothershipCrashesIntoIt(t *testing.T) {
	game := inRound(0)
	game.Own.Position = flight.Vector{0, 0, 47} // 2 short of the wall, a crash sphere of 2.8
	messages := game.Step(flight.Controls{}, false, false, 1.0/60, 0)
	if sentOfType(messages, protocol.TypeCrash) != 1 || !strings.Contains(strings.Join(messages, ""), `"with":"mothership"`) {
		t.Fatalf("sent %v, want a crash into the mothership", messages)
	}
}

func TestAnAttackerCannotRunIntoAGunnersTurret(t *testing.T) {
	game := inRound(0)
	game.Own.Position = game.Remotes["them"].Shown.Position
	if sentOfType(game.Step(flight.Controls{}, false, false, 1.0/60, 0), protocol.TypeCrash) != 0 {
		t.Fatal("ran into a turret")
	}
}

func TestAGunnerStaysInTheTurretAndTurnsItsAim(t *testing.T) {
	seat := &protocol.Seat{Position: flight.Vector{0, 10, 0}, Up: 1}
	game := inRoundAs(protocol.TeamGunners, seat, 0)
	game.Round.Rules.GunnerTurnRate = 1
	for frame := range 60 { // a second, turning right at full stick
		game.Step(flight.Controls{Yaw: 1, Throttle: 1}, false, false, 1.0/60, float64(frame)/60)
	}
	if game.Own.Position != seat.Position || game.Own.Speed != 0 {
		t.Fatalf("the gunner moved to %v at %v", game.Own.Position, game.Own.Speed)
	}
	facing := game.Own.Rotation.Apply(flight.Forward)
	if math.Abs(math.Atan2(facing[0], facing[2])+1) > 0.01 {
		t.Fatalf("after a second at 1 radian a second the turret faces %v, want 1 radian right", facing)
	}
}

func TestAGunnersBulletHitsAnAttackerAndNotTheMothership(t *testing.T) {
	seat := &protocol.Seat{Position: flight.Vector{0, 0, 0}, Up: 1}
	game := inRoundAs(protocol.TeamGunners, seat, 0)
	targets := []string{}
	for frame := range 30 {
		for _, message := range game.Step(flight.Controls{}, frame == 0, false, 1.0/60, float64(frame)/60) {
			if strings.Contains(message, `"type":"hit"`) {
				targets = append(targets, message)
			}
		}
	}
	if len(targets) != 1 || !strings.Contains(targets[0], `"target":"them"`) {
		t.Fatalf("hits %v, want one on the attacker", targets)
	}
}
