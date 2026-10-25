package ship

import (
	"math"
	"slices"
	"strconv"

	"github.com/akrck02/go-renderer/scene"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/protocol"
)

// shipMaterial lights every ship like the sun lights the island: the browser's flat shading.
var shipMaterial = &scene.Material{Name: "ship", Kind: scene.LitKind, BaseColor: [4]float32{1, 1, 1, 1}, DoubleSided: true}

// glowMaterial makes a light's bulb shine with its own colour, unshaded, as the browser draws it:
// go-renderer's GlowKind would only shine in the dark, and be shaded by the sun in space.
var glowMaterial = &scene.Material{Name: "ship light", Kind: scene.UnlitKind, BaseColor: [4]float32{1, 1, 1, 1}}

// partLight is the browser's light: a bulb, drawn as a sphere sized by its scale.
const partLight = "light"

// Model is a design's meshes: its opaque body, and the parts in other materials: its lights'
// bulbs, and its see-through parts, one mesh for each opacity. A node shows the body, and
// children of it the others.
type Model struct {
	Body   *scene.Mesh
	Others []*scene.Mesh
	Lamps  []Lamp // where its lights shine from
}

// Lamp is a light part as a light: where the design puts it (a mirrored one twice), its linear
// colour and its intensity.
type Lamp struct {
	Position  flight.Vector
	Colour    [3]float32
	Intensity float64
}

// lamps are the design's visible lights, as lamps.
func lamps(layers []protocol.Layer) []Lamp {
	var found []Lamp
	for _, layer := range layers {
		if !layer.Visible {
			continue
		}
		for _, part := range layer.Parts {
			if !isLight(part) {
				continue
			}
			lamp := Lamp{Position: part.Position, Colour: linearColour(part.Color), Intensity: max(part.Intensity, 0)}
			found = append(found, lamp)
			if part.Mirror {
				lamp.Position[0] = -lamp.Position[0]
				found = append(found, lamp)
			}
		}
	}
	return found
}

// BuildModel makes a design's meshes at a detail, in the design's own units.
func BuildModel(layers []protocol.Layer, detail Detail) Model {
	model := Model{Body: Build(layers, detail), Lamps: lamps(layers)}
	if glow := build(layers, detail, "ship lights", glowMaterial, isLight); len(glow.Positions) > 0 {
		model.Others = append(model.Others, glow)
	}
	for _, step := range opacitySteps(layers) {
		sameOpacity := func(part protocol.Part) bool { return !isLight(part) && opacityStep(part) == step }
		model.Others = append(model.Others, build(layers, detail, "see-through ship", seeThrough(step), sameOpacity))
	}
	return model
}

// opacityStepsInAll is how finely opacities are told apart: parts whose opacities round to the
// same twentieth share a mesh, so a ship stays a few draws.
const opacityStepsInAll = 20

// opacityStep is a part's opacity in twentieths: opacityStepsInAll for an opaque part.
func opacityStep(part protocol.Part) int {
	if part.Opacity <= 0 || part.Opacity >= 1 {
		return opacityStepsInAll
	}
	return max(1, int(math.Round(part.Opacity*opacityStepsInAll)))
}

// opacitySteps are the see-through opacities a design's visible parts have, in the order met.
func opacitySteps(layers []protocol.Layer) []int {
	var steps []int
	for _, layer := range layers {
		if !layer.Visible {
			continue
		}
		for _, part := range layer.Parts {
			if step := opacityStep(part); step < opacityStepsInAll && !isLight(part) && !slices.Contains(steps, step) {
				steps = append(steps, step)
			}
		}
	}
	return steps
}

// seeThroughMaterials are the see-through materials, one for each opacity step, made once.
var seeThroughMaterials = map[int]*scene.Material{}

// seeThrough is the ship's material, blended at an opacity step.
func seeThrough(step int) *scene.Material {
	if material, found := seeThroughMaterials[step]; found {
		return material
	}
	alpha := float32(step) / opacityStepsInAll
	material := &scene.Material{Name: "see-through ship", Kind: scene.LitKind, BaseColor: [4]float32{1, 1, 1, alpha}, DoubleSided: true, Transparent: true}
	seeThroughMaterials[step] = material
	return material
}

// Build makes one mesh of a ship's visible parts at a detail, in the design's own units: the
// caller scales it to the game's ship size. Lights and see-through parts are left out
// (BuildModel has them), and so are parts of a type it doesn't know.
func Build(layers []protocol.Layer, detail Detail) *scene.Mesh {
	opaque := func(part protocol.Part) bool { return !isLight(part) && opacityStep(part) == opacityStepsInAll }
	return build(layers, detail, "ship", shipMaterial, opaque)
}

func isLight(part protocol.Part) bool { return part.Type == partLight }

// build makes one mesh, in material, of the visible parts kept.
func build(layers []protocol.Layer, detail Detail, name string, material *scene.Material, kept func(protocol.Part) bool) *scene.Mesh {
	mesh := &scene.Mesh{Name: name, Material: material}
	// Made to measure: growing them by appending would hold up to twice the memory for a moment,
	// which a design of hundreds of parts can't spare on the console.
	vertices := 3 * triangles(layers, detail, kept)
	mesh.Positions = make([]float32, 0, 3*vertices)
	mesh.Normals = make([]float32, 0, 3*vertices)
	mesh.Colors = make([]float32, 0, 4*vertices)
	for _, layer := range layers {
		if !layer.Visible {
			continue
		}
		for _, part := range layer.Parts {
			if !kept(part) {
				continue
			}
			addPart(mesh, part, detail, part.Position, part.Rotation)
			if part.Mirror {
				// The browser's mirror: across X = 0, which turns the other way about Y and Z.
				addPart(mesh, part, detail, flight.Vector{-part.Position[0], part.Position[1], part.Position[2]},
					flight.Vector{part.Rotation[0], -part.Rotation[1], -part.Rotation[2]})
			}
		}
	}
	return mesh
}

// shapeOf is what a part is drawn with: its model's triangles and colours (none until the model
// is added), a sphere for a light's bulb, or its own primitive; colours is nil for the part's
// own colour.
func shapeOf(part protocol.Part, detail Detail) (triangles []triangle, colours [][3][3]float32) {
	switch {
	case part.Type == partModel:
		shape := models[part.ModelId]
		return shape.triangles, shape.colours
	case isLight(part):
		return primitive(partSphere, detail), nil
	}
	return primitive(part.Type, detail), nil
}

// Triangles is how many triangles BuildModel makes of a design at a detail.
func Triangles(layers []protocol.Layer, detail Detail) int {
	return triangles(layers, detail, func(protocol.Part) bool { return true })
}

func triangles(layers []protocol.Layer, detail Detail, kept func(protocol.Part) bool) int {
	count := 0
	for _, layer := range layers {
		if !layer.Visible {
			continue
		}
		for _, part := range layer.Parts {
			if !kept(part) {
				continue
			}
			shape, _ := shapeOf(part, detail)
			triangles := len(shape)
			if part.Mirror {
				triangles *= 2
			}
			count += triangles
		}
	}
	return count
}

// Within is the finest detail at which a design has at most maximum triangles, or the plainest
// when even that has more.
func Within(layers []protocol.Layer, maximum int) Detail {
	for _, detail := range Levels {
		if Triangles(layers, detail) <= maximum {
			return detail
		}
	}
	return Levels[len(Levels)-1]
}

// addPart places a primitive's triangles with the part's scale, then rotation, then position.
func addPart(mesh *scene.Mesh, part protocol.Part, detail Detail, position, rotation flight.Vector) {
	turn := flight.FromEulerXYZ(rotation[0], rotation[1], rotation[2])
	colour := linearColour(part.Color)
	shape, colours := shapeOf(part, detail)
	for each, corners := range shape {
		var placed triangle
		for index, corner := range corners {
			scaled := flight.Vector{corner[0] * part.Scale[0], corner[1] * part.Scale[1], corner[2] * part.Scale[2]}
			placed[index] = point(turn.Apply(scaled).Add(position))
		}
		normal := faceNormal(placed)
		for index, corner := range placed {
			if colours != nil {
				colour = colours[each][index]
			}
			mesh.Positions = append(mesh.Positions, float32(corner[0]), float32(corner[1]), float32(corner[2]))
			mesh.Normals = append(mesh.Normals, normal[0], normal[1], normal[2])
			mesh.Colors = append(mesh.Colors, colour[0], colour[1], colour[2], 1)
		}
	}
}

// faceNormal is the unit vector out of a triangle's face, so each face is lit flat.
func faceNormal(corners triangle) [3]float32 {
	first := flight.Vector(corners[1]).Subtract(flight.Vector(corners[0]))
	second := flight.Vector(corners[2]).Subtract(flight.Vector(corners[0]))
	cross := flight.Vector{
		first[1]*second[2] - first[2]*second[1],
		first[2]*second[0] - first[0]*second[2],
		first[0]*second[1] - first[1]*second[0],
	}
	length := cross.Length()
	if length == 0 {
		return [3]float32{0, 1, 0} // a triangle squashed to a line: any normal will do
	}
	return [3]float32{float32(cross[0] / length), float32(cross[1] / length), float32(cross[2] / length)}
}

// fallbackColour is a part's grey when its colour can't be read.
const fallbackColour = "#9aa7b8"

// linearColour reads "#rrggbb" in the screen's sRGB and returns its linear colour, as three.js
// does with a material's colour.
func linearColour(hex string) [3]float32 {
	value, err := strconv.ParseUint(trimHash(hex), 16, 32)
	if err != nil || len(hex) != len(fallbackColour) {
		value, _ = strconv.ParseUint(fallbackColour[1:], 16, 32) // a constant that always reads
	}
	var colour [3]float32
	for index := range colour {
		channel := float64((value>>(16-8*index))&0xFF) / 0xFF
		colour[index] = float32(sRGBToLinear(channel))
	}
	return colour
}

func trimHash(hex string) string {
	if len(hex) > 0 && hex[0] == '#' {
		return hex[1:]
	}
	return hex
}

// sRGBToLinear is the sRGB transfer curve undone.
func sRGBToLinear(channel float64) float64 {
	if channel < 0.04045 {
		return channel / 12.92
	}
	return math.Pow((channel+0.055)/1.055, 2.4)
}

// Radius is how far the ship reaches from its centre, as the server measures it
// (models.Project3d.BoundingRadius), which sets its scale in the game.
func Radius(layers []protocol.Layer) float64 {
	radius := 0.0
	for _, layer := range layers {
		if !layer.Visible {
			continue
		}
		for _, part := range layer.Parts {
			radius = max(radius, part.Position.Length()+part.Scale.Length()/2)
		}
	}
	return radius
}
