package arena

import (
	_ "embed"
	"encoding/json"
	"math"

	"github.com/akrck02/go-renderer/camera"
	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/scene"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/protocol"
	"github.com/jonbul/ships3D-3ds/ship"
)

// titleShipDesign is the Kestrel, the server's built-in fighter (game/defaultShips.go), kept here
// so the title screen needs no server.
//
//go:embed title-ship.json
var titleShipDesign []byte

// The title's flight: the ship crosses the screen from left to right ahead of a still camera,
// rolling as it goes, once every titleLoopSeconds.
const (
	titleLoopSeconds  = 7
	titleCrossing     = 90 // world units from one side to the other
	titleDistance     = 38 // ahead of the camera
	titleHeight       = -4
	titleRollsPerLoop = 1
	titleShipScale    = 1.2
)

// titleCamera looks straight ahead, +Z, from the origin.
var titleCamera = camera.Lens{
	Position:           graphics.Vector{0, 0, 0, 1},
	Target:             graphics.Vector{0, 0, 1, 1},
	Up:                 graphics.Vector{0, 1, 0, 0},
	Near:               CameraNear,
	Far:                CameraFar,
	FieldOfViewRadians: math.Pi / 4,
}

// titleShip builds the title's ship node, hidden until the title shows.
func titleShip() *scene.Node {
	var layers []protocol.Layer
	if err := json.Unmarshal(titleShipDesign, &layers); err != nil {
		println(err.Error()) // the debug log; the title goes on without its ship
	}
	return &scene.Node{Name: "title ship", Mesh: ship.Build(layers, ship.Fine), Hidden: true, Moving: true}
}

// showTitle flies the title's ship at now seconds and places the still camera.
func (world *World) ShowTitle(now float64, view *camera.Lens) {
	*view = titleCamera
	world.view.Eye, world.view.Target = flight.Vector{}, flight.Forward
	share := math.Mod(now, titleLoopSeconds) / titleLoopSeconds
	// Screen right is -X, looking along +Z: the ship goes from +X to -X, its nose toward -X.
	position := flight.Vector{titleCrossing * (0.5 - share), titleHeight + 3*math.Sin(share*2*math.Pi), titleDistance}
	heading := flight.FromEulerXYZ(0, -math.Pi/2, 0)
	roll := flight.FromEulerXYZ(0, 0, share*titleRollsPerLoop*2*math.Pi)
	world.title.Mesh, world.title.Hidden = world.titleMesh, false
	world.title.Transform = TransformOf(position, heading.Then(roll), titleShipScale)
	world.own.Hidden, world.reticle.Hidden = true, true
	dustCellsAround(world.dust, flight.Vector{})
	world.sky.Transform = graphics.Identity()
}
