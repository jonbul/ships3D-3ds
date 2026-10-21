package arena

import (
	"math"

	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/models"
	"github.com/akrck02/go-renderer/scene"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/protocol"
	"github.com/jonbul/ships3D-3ds/radar"
	"github.com/jonbul/ships3D-3ds/ship"
)

// The 3D radar, as the browser's (ships-vue-3d, src/game/radar.ts): a sphere of radius 1 seen from
// above and behind our ship, its equator (the ship's own plane) and two meridians, a tick to the
// nose, and a blip for each ship with a stalk down to the equator, so above and below read at a
// glance.
const (
	radarBlips        = 64   // most blips drawn at once
	radarBlipSize     = 0.06 // a blip's radius
	sphereScale       = 2    // a ship part's sphere is half a unit across a unit of scale
	radarStalkWidth   = 0.012
	radarRingWidth    = 0.012
	radarRingSegments = 48
	radarMarkerSize   = 0.14
)

// The radar's colours, linear as go-renderer takes them: the browser's #4fc3f7 lines at its
// opacities (0.55 for the equator, 0.4 the nose's tick, 0.18 the meridians), here mixed into the
// dark so the 3DS needs no blending; its red blips, dimmer out of range, and its pink stalks.
var (
	radarLineColour     = [3]float32{0.078, 0.546, 0.93} // #4fc3f7
	radarEquatorColour  = dimmed(radarLineColour, 0.55)
	radarNoseColour     = dimmed(radarLineColour, 0.4)
	radarMeridianColour = dimmed(radarLineColour, 0.18)
	radarShellColour    = graphics.Vec4{0.078, 0.546, 0.93, 0.07}
	radarEnemyColour    = graphics.Vec4{1, 0.085, 0.085, 1}   // #ff5252
	radarFarColour      = graphics.Vec4{0.4, 0.034, 0.034, 1} // #ff5252 at 0.4
	radarStalkColour    = [3]float32{0.6, 0.152, 0.13}        // #ff8a80 at 0.6
	// RadarBackground is the radar's dark, what its screen is cleared to.
	RadarBackground = graphics.Vec4{0.0044, 0.010, 0.016, 1} // #0e1a22, the touch screen's dark
)

// radarFieldOfView is the browser's radar camera's, 32 degrees, in radians.
const radarFieldOfView = 32 * math.Pi / 180

// RadarView looks at the sphere from above and behind, as the browser's radar camera does.
var RadarView = models.Camera{
	Position:    graphics.Vec4{0, 1.5, -3.4, 1},
	Target:      graphics.Vec4{0, -0.1, 0, 1},
	Up:          graphics.Vec4{0, 1, 0, 0},
	Near:        0.1,
	Far:         20,
	FieldOfView: radarFieldOfView,
}

// Radar is the 3D radar's scene and what moves in it.
type Radar struct {
	Scene  *scene.Scene
	blips  []*scene.Node // each with a mesh of its own, so in range and out of it can differ
	stalks []*scene.Node
}

// NewRadar builds the radar in a scene of its own.
func NewRadar() *Radar { return NewRadarIn(scene.New()) }

// NewRadarIn builds the radar inside a scene that already exists, as a game engine's.
func NewRadarIn(sphere *scene.Scene) *Radar {
	// By day, so nothing is dimmed for the night; the sun straight up, so there is no twilight
	// glow on the background.
	sphere.Environment.SunDirection = graphics.Vec4{0, 1, 0, 0}
	sphere.Environment.SkyZenith, sphere.Environment.SkyHorizon = RadarBackground, RadarBackground
	sphere.Environment.FogColor = RadarBackground
	still := func(name string, mesh *scene.Mesh, transform graphics.Mat4) {
		sphere.Nodes = append(sphere.Nodes, &scene.Node{Name: name, Transform: transform, Mesh: mesh})
	}
	still("shell", shell(), graphics.Identity())
	still("equator", ring(1, radarEquatorColour), graphics.Identity())
	still("meridian", ring(1, radarMeridianColour), graphics.RotateX(quarterTurn))
	still("meridian", ring(1, radarMeridianColour), graphics.RotateY(quarterTurn).Multiply(graphics.RotateX(quarterTurn)))
	nose := solid(&scene.Mesh{Name: "nose"})
	AppendBox(nose, [3]float32{0, 0, 0.5}, radarRingWidth, radarRingWidth, 1, radarNoseColour)
	still("nose", nose, graphics.Identity())
	still("own", ownMarker(), graphics.Identity())
	built := &Radar{Scene: sphere}
	blip := solid(ship.Build([]protocol.Layer{{Visible: true, Parts: []protocol.Part{{
		Type: "sphere", Scale: flight.Vector{radarBlipSize * sphereScale, radarBlipSize * sphereScale, radarBlipSize * sphereScale}, Color: "#ffffff"}}}}, ship.Coarse))
	stalk := Box("stalk", radarStalkWidth, 1, radarStalkWidth, radarStalkColour)
	for range radarBlips {
		blipNode := &scene.Node{Name: "blip", Mesh: WithOwnMaterial(blip), Hidden: true, Moving: true}
		stalkNode := &scene.Node{Name: "stalk", Mesh: stalk, Hidden: true, Moving: true}
		built.blips, built.stalks = append(built.blips, blipNode), append(built.stalks, stalkNode)
		sphere.Nodes = append(sphere.Nodes, blipNode, stalkNode)
	}
	return built
}

// Show places a blip for every ship, as radar.Place put it, each with its stalk to the equator.
func (shown3D *Radar) Show(blips map[string]radar.Blip) {
	drawn := 0
	for _, placed := range blips {
		if drawn == len(shown3D.blips) {
			break
		}
		x, y, z := placed.Scene()
		blip, stalk := shown3D.blips[drawn], shown3D.stalks[drawn]
		blip.Mesh.Material.BaseColor = radarEnemyColour
		if !placed.InRange {
			blip.Mesh.Material.BaseColor = radarFarColour
		}
		blip.Hidden, stalk.Hidden = false, math.Abs(y) < radarBlipSize
		blip.Transform = graphics.Translate(graphics.Vec4{float32(x), float32(y), float32(z), 0})
		stalk.Transform = graphics.Translate(graphics.Vec4{float32(x), float32(y / 2), float32(z), 0}).
			Multiply(graphics.Scale(graphics.Vec4{1, float32(math.Abs(y)), 1, 0}))
		drawn++
	}
	for index := drawn; index < len(shown3D.blips); index++ {
		shown3D.blips[index].Hidden, shown3D.stalks[index].Hidden = true, true
	}
}

// shell is the sphere's faint glass, seen from inside and out.
func shell() *scene.Mesh {
	mesh := solid(ship.Build([]protocol.Layer{{Visible: true, Parts: []protocol.Part{{
		Type: "sphere", Scale: flight.Vector{sphereScale, sphereScale, sphereScale}, Color: "#ffffff"}}}}, ship.Fine))
	mesh.Material.BaseColor, mesh.Material.Transparent, mesh.Material.DoubleSided = radarShellColour, true, true
	return mesh
}

// dimmed is colour at share of its strength over the radar's near-black background.
func dimmed(colour [3]float32, share float32) [3]float32 {
	return [3]float32{colour[0] * share, colour[1] * share, colour[2] * share}
}

// solid is a mesh drawn in its own flat colours, without light: the radar is a screen, not a
// lit object.
func solid(mesh *scene.Mesh) *scene.Mesh {
	mesh.Material = &scene.Material{Name: "radar", Kind: scene.KindUnlit, BaseColor: graphics.Vec4{1, 1, 1, 1}}
	return mesh
}

// ring is a flat circle of radius in the equator's plane, a band radarRingWidth wide.
func ring(radius float64, colour [3]float32) *scene.Mesh {
	mesh := solid(&scene.Mesh{Name: "ring"})
	inner, outer := float32(radius-radarRingWidth/2), float32(radius+radarRingWidth/2)
	for segment := range radarRingSegments {
		first := float64(segment) / radarRingSegments * 2 * math.Pi
		second := float64(segment+1) / radarRingSegments * 2 * math.Pi
		corners := [4][3]float32{
			{inner * float32(math.Cos(first)), 0, inner * float32(math.Sin(first))},
			{outer * float32(math.Cos(first)), 0, outer * float32(math.Sin(first))},
			{outer * float32(math.Cos(second)), 0, outer * float32(math.Sin(second))},
			{inner * float32(math.Cos(second)), 0, inner * float32(math.Sin(second))},
		}
		for _, index := range [6]int{0, 1, 2, 0, 2, 3} {
			mesh.Positions = append(mesh.Positions, corners[index][0], corners[index][1], corners[index][2])
			mesh.Normals = append(mesh.Normals, 0, 1, 0)
			mesh.Colors = append(mesh.Colors, colour[0], colour[1], colour[2], 1)
		}
	}
	return mesh
}

// ownMarker is a small cone at the centre, its tip ahead: our ship.
func ownMarker() *scene.Mesh {
	marker := ship.Build([]protocol.Layer{{Visible: true, Parts: []protocol.Part{{
		Type: "cone", Rotation: flight.Vector{math.Pi / 2, 0, 0},
		Scale: flight.Vector{radarMarkerSize * 0.6, radarMarkerSize, radarMarkerSize * 0.4}, Color: "#4fc3f7"}}}}, ship.Coarse)
	return solid(marker)
}
