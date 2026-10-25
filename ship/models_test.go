package ship

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"testing"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/protocol"
)

// oneTriangleModel is the smallest .glb a designer could import: one red triangle, 2 wide and 4
// high, on a node moved 5 along X, written as glTF lays a .glb out (a header, then a JSON chunk
// and a binary chunk, each padded to four bytes).
func oneTriangleModel() []byte {
	var positions bytes.Buffer
	for _, value := range []float32{0, 0, 0, 2, 0, 0, 0, 4, 0} {
		binary.Write(&positions, binary.LittleEndian, value) //nolint:errcheck // a bytes.Buffer takes every write
	}
	json := []byte(`{"asset":{"version":"2.0"},"scene":0,"scenes":[{"nodes":[0]}],` +
		`"nodes":[{"mesh":0,"translation":[5,0,0]}],` +
		`"meshes":[{"primitives":[{"attributes":{"POSITION":0},"material":0}]}],` +
		`"materials":[{"pbrMetallicRoughness":{"baseColorFactor":[1,0,0,1]}}],` +
		`"accessors":[{"bufferView":0,"componentType":5126,"count":3,"type":"VEC3","min":[0,0,0],"max":[2,4,0]}],` +
		`"bufferViews":[{"buffer":0,"byteLength":36}],"buffers":[{"byteLength":36}]}`)
	for len(json)%4 != 0 {
		json = append(json, ' ')
	}
	var file bytes.Buffer
	chunk := func(kind string, content []byte) {
		binary.Write(&file, binary.LittleEndian, uint32(len(content))) //nolint:errcheck // as above
		file.WriteString(kind)
		file.Write(content)
	}
	file.WriteString("glTF")
	binary.Write(&file, binary.LittleEndian, uint32(2))                                //nolint:errcheck // as above
	binary.Write(&file, binary.LittleEndian, uint32(12+8+len(json)+8+positions.Len())) //nolint:errcheck // as above
	chunk("JSON", json)
	chunk("BIN\x00", positions.Bytes())
	return file.Bytes()
}

func TestAModelIsFittedIntoTheUnitCubeAndKeepsItsColour(t *testing.T) {
	shape, err := ReadModel(oneTriangleModel(), 10)
	if err != nil {
		t.Fatal(err)
	}
	if shape.Triangles() != 1 {
		t.Fatalf("%d triangles, want 1", shape.Triangles())
	}
	want := triangle{{-0.25, -0.5, 0}, {0.25, -0.5, 0}, {-0.25, 0.5, 0}}
	for corner := range want {
		for axis := range 3 {
			if math.Abs(shape.triangles[0][corner][axis]-want[corner][axis]) > 1e-6 {
				t.Fatalf("the triangle is %v, want %v: centred, and 1 along its longest side", shape.triangles[0], want)
			}
		}
	}
	if shape.colours[0][0] != [3]float32{1, 0, 0} {
		t.Errorf("the colour is %v, want the model's own red", shape.colours[0][0])
	}
}

func TestAModelTooBigForTheConsoleIsRefused(t *testing.T) {
	if _, err := ReadModel(oneTriangleModel(), 0); !errors.Is(err, ErrModelTooBig) {
		t.Fatalf("got %v, want %v", err, ErrModelTooBig)
	}
	if _, err := ReadModel([]byte("not a model"), 10); !errors.Is(err, ErrModelUnreadable) {
		t.Fatalf("got %v, want %v", err, ErrModelUnreadable)
	}
}

func TestAModelPartShowsOnceItsModelIsAdded(t *testing.T) {
	part := protocol.Part{Type: partModel, ModelId: "test-triangle", Position: flight.Vector{0, 10, 0}, Scale: flight.Vector{2, 2, 2}, Color: "#ffffff"}
	layers := []protocol.Layer{{Visible: true, Parts: []protocol.Part{part, boxPart(flight.Vector{}, false)}}}
	box := len(primitive(partBox, Fine))
	if missing := MissingModels(layers); len(missing) != 1 || missing[0] != part.ModelId {
		t.Fatalf("missing %v, want the part's model", missing)
	}
	if got := len(Build(layers, Fine).Positions) / 9; got != box {
		t.Fatalf("before the model arrives the design has %d triangles, want the box's %d", got, box)
	}
	shape, err := ReadModel(oneTriangleModel(), 10)
	if err != nil {
		t.Fatal(err)
	}
	AddModel(part.ModelId, shape)
	defer delete(models, part.ModelId)
	mesh := Build(layers, Fine)
	if got := len(mesh.Positions) / 9; got != box+1 || len(MissingModels(layers)) != 0 || !ShowsModel(layers, part.ModelId) {
		t.Fatalf("with the model the design has %d triangles, want %d", got, box+1)
	}
	if top := mesh.Positions[7]; math.Abs(float64(top)-11) > 1e-5 {
		t.Errorf("the model's top is at %v, want 11: half its scale of 2 above its place", top)
	}
	if red := mesh.Colors[:3]; red[0] != 1 || red[1] != 0 {
		t.Errorf("the model is coloured %v, want its own red rather than the part's white", red)
	}
}

func TestEveryErrorHasItsOwnCode(t *testing.T) {
	codes := map[error]string{ErrModelUnreadable: "ship.model-unreadable", ErrModelTooBig: "ship.model-too-big"}
	for err, code := range codes {
		if err.Error() != code {
			t.Errorf("%v has the code %q, want %q", err, err.Error(), code)
		}
	}
}
