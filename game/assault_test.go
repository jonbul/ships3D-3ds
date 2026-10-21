package game

import (
	"testing"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/protocol"
)

// startsIn is a lobby's countdown, as the server sends it.
func startsIn(milliseconds float64) *float64 { return &milliseconds }

// inRound is a game in an assault round: we are "me", an attacker at the origin; "them" is a
// gunner in a turret.
func inRound(now float64) *Game {
	game := New()
	game.Handle(protocol.Incoming{Type: protocol.TypeLobby, Id: "me", MinPlayers: 2, MaxPlayers: 8}, nil, 0)
	game.Handle(protocol.Incoming{
		Type: protocol.TypeRoundStart, Id: "me", Settings: settings, EndsInMs: 180000,
		Rules:      protocol.Rules{MothershipLife: 200, RoundMs: 180000},
		Mothership: protocol.Mothership{Life: 200, MaxLife: 200, Scale: 2},
		Players: []protocol.Player{
			{Id: "me", Rotation: flight.NoRotation, Life: 10, Team: protocol.TeamAttackers},
			{Id: "them", Rotation: flight.NoRotation, Life: 10, Team: protocol.TeamGunners, Seat: &protocol.Seat{Position: flight.Vector{0, 5, 0}, Up: 1}},
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
