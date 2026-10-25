package ship

import (
	"bytes"
	"fmt"
	"math"
	"slices"

	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/loaders"
	"github.com/akrck02/go-renderer/scene"

	"github.com/jonbul/ships3D-3ds/protocol"
)

// partModel is the browser's imported model: a part showing a .glb file the server keeps.
const partModel = "model"

// Shape is an imported model made ready to place like a primitive: its triangles fitted into the
// unit cube centred on the origin along its longest side, keeping its proportions (the browser's
// fitUnitCube, in src/shared/models.ts), and each corner's colour, since a model keeps its own
// look rather than the part's colour.
type Shape struct {
	triangles []triangle
	colours   [][3][3]float32 // linear, one for each corner of each triangle
}

// Triangles is how many triangles the shape has.
func (shape Shape) Triangles() int { return len(shape.triangles) }

// models are the shapes read so far, by the model's id: parts whose model isn't here yet are
// left out, and their design is built again once it is (AddModel).
var models = map[string]Shape{}

// AddModel keeps a model's shape for the parts that show it.
func AddModel(id string, shape Shape) { models[id] = shape }

// MissingModels are the ids of the models a design's visible parts show that haven't been added.
func MissingModels(layers []protocol.Layer) []string {
	var missing []string
	for _, layer := range layers {
		if !layer.Visible {
			continue
		}
		for _, part := range layer.Parts {
			if _, found := models[part.ModelId]; part.Type == partModel && !found && !slices.Contains(missing, part.ModelId) {
				missing = append(missing, part.ModelId)
			}
		}
	}
	return missing
}

// ShowsModel tells whether a design has a visible part showing a model.
func ShowsModel(layers []protocol.Layer, id string) bool {
	for _, layer := range layers {
		for _, part := range layer.Parts {
			if layer.Visible && part.Type == partModel && part.ModelId == id {
				return true
			}
		}
	}
	return false
}

// ReadModel reads a .glb file into a shape. One with more than maximumTriangles is refused, so a
// console doesn't run out of memory on a model made for a computer.
func ReadModel(file []byte, maximumTriangles int) (Shape, error) {
	read, err := loaders.ReadScene(bytes.NewReader(file))
	if err != nil {
		return Shape{}, fmt.Errorf("%w: %w", ErrModelUnreadable, err)
	}
	var shape Shape
	tooMany := false
	read.Walk(func(node *scene.Node, world graphics.Matrix) {
		if node.Mesh == nil || tooMany {
			return
		}
		tooMany = !appendMesh(&shape, node.Mesh, world, maximumTriangles)
	})
	if tooMany {
		return Shape{}, ErrModelTooBig
	}
	fit(shape.triangles)
	return shape, nil
}

// appendMesh adds a mesh's triangles, placed by its world transform; false when that would pass
// maximum triangles.
func appendMesh(shape *Shape, mesh *scene.Mesh, world graphics.Matrix, maximum int) bool {
	corners := len(mesh.Positions) / 3
	if len(mesh.Indices) > 0 {
		corners = len(mesh.Indices)
	}
	if shape.Triangles()+corners/3 > maximum {
		return false
	}
	base := [3]float32{1, 1, 1}
	if mesh.Material != nil {
		base = [3]float32{mesh.Material.BaseColor[0], mesh.Material.BaseColor[1], mesh.Material.BaseColor[2]}
	}
	for first := 0; first+2 < corners; first += 3 {
		var placed triangle
		var colours [3][3]float32
		for index := range placed {
			vertex := first + index
			if len(mesh.Indices) > 0 {
				vertex = int(mesh.Indices[vertex])
			}
			position := world.Apply(graphics.Vector{mesh.Positions[3*vertex], mesh.Positions[3*vertex+1], mesh.Positions[3*vertex+2], 1})
			placed[index] = point{float64(position[0]), float64(position[1]), float64(position[2])}
			colours[index] = base
			if len(mesh.Colors) >= 4*(vertex+1) {
				for channel := range colours[index] {
					colours[index][channel] *= mesh.Colors[4*vertex+channel]
				}
			}
		}
		shape.triangles = append(shape.triangles, placed)
		shape.colours = append(shape.colours, colours)
	}
	return true
}

// fit moves and scales triangles so their box is centred on the origin and its longest side is 1.
func fit(triangles []triangle) {
	if len(triangles) == 0 {
		return
	}
	lowest, highest := triangles[0][0], triangles[0][0]
	for _, corners := range triangles {
		for _, corner := range corners {
			for axis := range corner {
				lowest[axis], highest[axis] = math.Min(lowest[axis], corner[axis]), math.Max(highest[axis], corner[axis])
			}
		}
	}
	longest := math.Max(highest[0]-lowest[0], math.Max(highest[1]-lowest[1], highest[2]-lowest[2]))
	if longest == 0 {
		return
	}
	for index := range triangles {
		for corner := range triangles[index] {
			for axis := range 3 {
				centre := (lowest[axis] + highest[axis]) / 2
				triangles[index][corner][axis] = (triangles[index][corner][axis] - centre) / longest
			}
		}
	}
}
