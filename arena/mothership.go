package arena

import (
	"math"

	"github.com/akrck02/go-renderer/scene"

	"github.com/jonbul/ships3D-3ds/chase"
	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/game"
	"github.com/jonbul/ships3D-3ds/protocol"
	"github.com/jonbul/ships3D-3ds/ship"
)

// The assault's mothership is drawn from its design at the middle of the world, never turned;
// it doesn't move, so its light is baked once. Its gunners are ball turrets on its hull, as the
// browser draws them.

// turretShare is how large a turret is, in fighters' radii: the mothership is huge next to them.
const turretShare = 2

// turretDesign is a ball turret of radius one with twin barrels along +Z, amber and lit from
// within so it stands out against the grey hull (the browser's turret.ts).
var turretDesign = []protocol.Layer{{Name: "turret", Visible: true, Parts: []protocol.Part{
	{Type: "sphere", Scale: flight.Vector{2, 2, 2}, Color: "#ffb74d"},
	{Type: "cylinder", Position: flight.Vector{0.35, 0.1, 1.6}, Rotation: flight.Vector{quarterTurn, 0, 0}, Scale: flight.Vector{0.4, 2.6, 0.4}, Color: "#3d4148", Mirror: true},
}}}

// quarterTurn is a quarter of a circle, in radians: a cylinder stands up the Y axis, and the
// barrels point along Z.
const quarterTurn = math.Pi / 2

// newTurret is the turrets' mesh, shared by all of them.
func newTurret() *scene.Mesh { return ship.Build(turretDesign, ship.Fine) }

// showMothership draws the round's mothership, or hides it outside a round.
func (world *World) showMothership(state *game.Game) {
	round := state.Round
	world.mothership.Hidden = round == nil
	if round == nil {
		return
	}
	design := round.Mothership
	Wear(world.mothership, world.ShipModel(design.Ship.Id, design.Ship.Layers, false))
	world.mothership.Transform = TransformOf(flight.Vector{}, flight.NoRotation, design.Scale)
}

// turretEyeShare is how far above (or, hanging under the hull, below) its seat a gunner looks
// out from, in fighters' radii: from the top of the ball.
const turretEyeShare = 0.8

// gunnerView is what a gunner sees: out of their turret the way it points, with the world's up,
// for the mothership is level.
func gunnerView(state *game.Game) chase.View {
	seat := state.Round.Seat
	eye := seat.Position.Add(flight.Vector{0, seat.Up * turretEyeShare * state.Settings.ShipRadius, 0})
	return chase.View{Eye: eye, Target: eye.Add(state.Own.Rotation.Apply(flight.Forward).Times(reticleDistance)), Up: flight.Vector{0, 1, 0}}
}
