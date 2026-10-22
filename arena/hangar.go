package arena

import (
	"math"

	"github.com/akrck02/go-renderer/camera"
	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/scene"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/protocol"
	"github.com/jonbul/ships3D-3ds/ship"
)

// The hangar the ship screen's ship turns in: a round platform floating in open space, with a
// glowing rim, a planet far behind and spotlights round its edge whose beams sweep up into the
// stars. It is built of the same parts as the players' ships, lit as they are; the beams are
// translucent. Distances are in the preview's units; the ship is previewRadius big at
// HangarShipPlace.
const (
	platformTop       = -2.6
	platformRadius    = 10
	platformThickness = 1.2
	keelDepth         = 7
	deckInset         = 1.5 // the deck's plate inside the platform's rim
	spotlightsAround  = 4
	spotlightRadius   = platformRadius - 0.8
	lampSize          = 1
	beamHeight        = 40
	beamSlices        = 6
	beamWidthAtBase   = 0.6
	beamWidthAtTop    = 3
	beamOpacityAtBase = 0.1
	beamLean          = 0.3 // radians the beams lean outward
	beamSway          = 0.2 // radians they sweep either way
	beamSwaySpeed     = 0.5 // radians a second
	beamPhaseStep     = 1.7 // radians between one beam's sweep and the next's, so they never move as one
	previewSkyRadius  = 150 // the stars' sphere, inside the preview camera's reach
	planetDiameter    = 45
)

// The hangar's colours: the parts' as the editor writes them, the beams' and the dark's linear.
var (
	hangarHull   = "#3a4658"
	hangarDeck   = "#5a6a80"
	hangarKeel   = "#252d3a"
	hangarRim    = "#5fb0ff"
	hangarLamp   = "#e8f2ff"
	hangarPlanet = "#6a4fb0"
	hangarBeam   = [3]float32{0.45, 0.7, 1.0}
	// HangarBack is the far dark around the hangar, what the screen is cleared to.
	HangarBack  = graphics.Vector{0.002, 0.004, 0.012, 1} // space, what the screen is cleared to
	planetPlace = flight.Vector{-70, -20, 200}
)

// HangarShipPlace is where the ship turns on the platform, and HangarCamera looks at it a little
// from above, where a design shows best: its top, its cockpit and its wings, rather than its belly.
var (
	HangarShipPlace = flight.Vector{0, 0, 22}
	HangarCamera    = camera.Lens{
		Position: graphics.Vector{0, 7, -3, 1},
		Target:   graphics.Vector{0, 1, 22, 1},
		Up:       graphics.Vector{0, 1, 0, 0},
		Near:     1, Far: 400,
	}
)

// Hangar is the hangar's nodes: the still part and the spotlights' beams, which sweep.
type Hangar struct {
	Nodes []*scene.Node
	beams []spotlight
}

// spotlight is one beam, where it stands and which way it leans out.
type spotlight struct {
	node     *scene.Node
	position flight.Vector
	outward  float64 // -1 when it stands left of the ship, 1 right
	phase    float64 // where in its sweep it starts, radians
}

// NewHangar builds the hangar: add its Nodes to a scene, and Show it every frame.
func NewHangar() *Hangar {
	built := &Hangar{}
	shrink := float32(previewSkyRadius) / SkyRadius
	stars := &scene.Node{Name: "stars", Mesh: Stars(), Transform: graphics.Scale(graphics.Vector{shrink, shrink, shrink, 0})}
	still := &scene.Node{Name: "hangar", Transform: graphics.Identity(), Mesh: ship.Build([]protocol.Layer{{Visible: true, Parts: hangarParts()}}, ship.Fine)}
	built.Nodes = append(built.Nodes, stars, still)
	beam := beamMesh()
	for index := range spotlightsAround {
		position := spotlightPlace(index)
		node := &scene.Node{Name: "beam", Mesh: beam}
		built.Nodes = append(built.Nodes, node)
		built.beams = append(built.beams, spotlight{node: node, position: position, outward: math.Copysign(1, position[0]), phase: float64(index) * beamPhaseStep})
	}
	return built
}

// spotlightPlace is where a spotlight stands on the platform's edge, the four of them between
// the ship's nose, tail and sides.
func spotlightPlace(index int) flight.Vector {
	angle := (float64(index) + 0.5) * 2 * math.Pi / spotlightsAround
	return flight.Vector{spotlightRadius * math.Sin(angle), platformTop, HangarShipPlace[2] + spotlightRadius*math.Cos(angle)}
}

// hangarParts are the platform, its rim, its keel, the lamps and the far planet.
func hangarParts() []protocol.Part {
	centre := HangarShipPlace[2]
	part := func(kind string, position, rotation, scale flight.Vector, colour string) protocol.Part {
		return protocol.Part{Type: kind, Position: position, Rotation: rotation, Scale: scale, Color: colour}
	}
	upsideDown := flight.Vector{math.Pi, 0, 0}
	parts := []protocol.Part{
		part("cylinder", flight.Vector{0, platformTop - platformThickness/2, centre}, flight.Vector{}, flight.Vector{2 * platformRadius, platformThickness, 2 * platformRadius}, hangarHull),
		part("cylinder", flight.Vector{0, platformTop + 0.05, centre}, flight.Vector{}, flight.Vector{2 * (platformRadius - deckInset), 0.1, 2 * (platformRadius - deckInset)}, hangarDeck),
		part("torus", flight.Vector{0, platformTop, centre}, flight.Vector{math.Pi / 2, 0, 0}, flight.Vector{2*platformRadius + 0.6, 2*platformRadius + 0.6, 0.6}, hangarRim),
		part("cone", flight.Vector{0, platformTop - platformThickness - keelDepth/2, centre}, upsideDown, flight.Vector{1.4 * platformRadius, keelDepth, 1.4 * platformRadius}, hangarKeel),
		part("sphere", planetPlace, flight.Vector{}, flight.Vector{planetDiameter, planetDiameter, planetDiameter}, hangarPlanet),
	}
	for index := range spotlightsAround {
		lamp := spotlightPlace(index)
		lamp[1] += lampSize / 4
		parts = append(parts, part("sphere", lamp, flight.Vector{}, flight.Vector{lampSize, lampSize / 2, lampSize}, hangarLamp))
	}
	return parts
}

// Show sweeps the beams to where they are at now seconds.
func (built *Hangar) Show(now float64) {
	for _, light := range built.beams {
		lean := -light.outward * (beamLean + beamSway*math.Sin(now*beamSwaySpeed+light.phase))
		light.node.Transform = TransformOf(light.position, flight.FromEulerXYZ(0, 0, lean), 1)
	}
}

// beamMesh is a spotlight's beam, standing up from its foot: slices that widen and fade as they
// rise, so the light seems to spread into the dark.
func beamMesh() *scene.Mesh {
	mesh := &scene.Mesh{Name: "beam", Material: &scene.Material{Kind: scene.UnlitKind, BaseColor: graphics.Vector{1, 1, 1, 1}, Transparent: true, DoubleSided: true}}
	sliceHeight := float32(beamHeight) / beamSlices
	for slice := range beamSlices {
		share := (float32(slice) + 0.5) / beamSlices
		width := beamWidthAtBase + (beamWidthAtTop-beamWidthAtBase)*share
		start := len(mesh.Colors)
		AppendBox(mesh, [3]float32{0, (float32(slice) + 0.5) * sliceHeight, 0}, width, sliceHeight, width, hangarBeam)
		for colour := start + 3; colour < len(mesh.Colors); colour += 4 {
			mesh.Colors[colour] = beamOpacityAtBase * (1 - share)
		}
	}
	return mesh
}
