package arena

import (
	"testing"

	"github.com/akrck02/go-renderer/models"

	"github.com/jonbul/ships3D-3ds/chase"
	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/game"
	"github.com/jonbul/ships3D-3ds/protocol"
)

func TestTheMatchSeesAsWideAsTheChaseCameraAfterTheTitle(test *testing.T) {
	world := NewWorld()
	var view models.Camera
	world.ShowTitle(0, &view)
	world.Show(game.New(), nil, &view, 0, 0)
	if view.FieldOfView != chase.FieldOfView {
		test.Errorf("the match sees %v radians across, want the chase camera's %v", view.FieldOfView, chase.FieldOfView)
	}
}

// assaultRound is a game in an assault round, we an attacker, "them" a gunner in a turret, and a
// one-part mothership at the middle of the world.
func assaultRound() *game.Game {
	state := game.New()
	state.Handle(protocol.Incoming{Type: protocol.TypeLobby, Id: "me"}, nil, 0)
	state.Handle(protocol.Incoming{
		Type: protocol.TypeRoundStart, Id: "me", EndsInMs: 180000,
		Settings: protocol.Settings{Settings: flight.Settings{WorldRadius: 1500}, ShipRadius: 4, TickRate: 20},
		Mothership: protocol.Mothership{Life: 100, MaxLife: 100, Scale: 3, Ship: protocol.Ship{Id: "mothership", Layers: []protocol.Layer{{Visible: true,
			Parts: []protocol.Part{{Type: "box", Scale: flight.Vector{10, 2, 20}, Color: "#888888"}}}}}},
		Players: []protocol.Player{
			{Id: "me", Position: flight.Vector{0, 0, -100}, Rotation: flight.NoRotation, Life: 10, Team: protocol.TeamAttackers},
			{Id: "them", Position: flight.Vector{0, 4, 0}, Rotation: flight.NoRotation, Life: 10, Team: protocol.TeamGunners, Seat: &protocol.Seat{Position: flight.Vector{0, 4, 0}, Up: 1}},
		},
	}, nil, 0)
	return state
}

func TestTheMothershipIsBuiltFromItsDesign(test *testing.T) {
	world := NewWorld()
	var view models.Camera
	world.Show(assaultRound(), nil, &view, 0, 0)
	if world.mothership.Hidden || world.mothership.Mesh == nil || len(world.mothership.Mesh.Positions) == 0 {
		test.Fatalf("the mothership is hidden %v with mesh %v", world.mothership.Hidden, world.mothership.Mesh)
	}
	world.Show(game.New(), nil, &view, 0, 1) // a free for all, or back in the lobby
	if !world.mothership.Hidden {
		test.Fatal("the mothership shows outside a round")
	}
}

func TestAGunnerIsATurretOnTheHull(test *testing.T) {
	world := NewWorld()
	var view models.Camera
	world.Show(assaultRound(), nil, &view, 0, 0)
	node := world.remotes["them"]
	if node == nil || node.Hidden || node.Mesh != world.turret {
		test.Fatalf("the gunner is drawn as %+v, want the turret", node)
	}
}
