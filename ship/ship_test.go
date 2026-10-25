package ship

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/akrck02/go-renderer/scene"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/protocol"
)

var everyType = []string{partBox, partSphere, partCylinder, partCone, partWedge, partTorus, partDome, partCapsule, partPyramid, partTube, partPrism, partFrustum}

// slack is how far a corner may stray past the cube from rounding.
const slack = 1e-9

func TestEachPrimitiveFitsTheUnitCube(t *testing.T) {
	for _, kind := range everyType {
		triangles := primitive(kind, Fine)
		if len(triangles) == 0 {
			t.Errorf("%s has no triangles", kind)
		}
		var low, high point
		for _, corners := range triangles {
			for _, corner := range corners {
				for axis := range 3 {
					low[axis], high[axis] = min(low[axis], corner[axis]), max(high[axis], corner[axis])
				}
			}
		}
		for axis := range 3 {
			if low[axis] < -0.5-slack || high[axis] > 0.5+slack {
				t.Errorf("%s spans %v to %v, outside the unit cube", kind, low, high)
			}
		}
	}
}

func TestEachPrimitiveFillsItsHeight(t *testing.T) {
	for _, kind := range everyType {
		if kind == partTorus {
			continue // it lies in XY, a ring around Z
		}
		low, high := math.Inf(1), math.Inf(-1)
		for _, corners := range primitive(kind, Fine) {
			for _, corner := range corners {
				low, high = min(low, corner[1]), max(high, corner[1])
			}
		}
		if math.Abs(low+0.5) > slack || math.Abs(high-0.5) > slack {
			t.Errorf("%s goes from %v to %v in height, want -0.5 to 0.5", kind, low, high)
		}
	}
}

func boxPart(position flight.Vector, mirror bool) protocol.Part {
	return protocol.Part{Type: partBox, Position: position, Scale: flight.Vector{1, 1, 1}, Color: "#ffffff", Mirror: mirror}
}

func TestAMirroredPartIsReflectedAcrossX(t *testing.T) {
	mesh := Build([]protocol.Layer{{Visible: true, Parts: []protocol.Part{boxPart(flight.Vector{3, 0, 0}, true)}}}, Fine)
	lowest, highest := float32(math.Inf(1)), float32(math.Inf(-1))
	for index := 0; index < len(mesh.Positions); index += 3 {
		lowest, highest = min(lowest, mesh.Positions[index]), max(highest, mesh.Positions[index])
	}
	if lowest != -3.5 || highest != 3.5 {
		t.Fatalf("x from %v to %v, want a box at each side", lowest, highest)
	}
}

func TestHiddenLayersAreNotBuilt(t *testing.T) {
	mesh := Build([]protocol.Layer{{Visible: false, Parts: []protocol.Part{boxPart(flight.Vector{}, false)}}}, Fine)
	if mesh.VertexCount() != 0 {
		t.Fatalf("%d vertices from a hidden layer", mesh.VertexCount())
	}
}

func TestTheRadiusIsTheServersBoundingRadius(t *testing.T) {
	layers := []protocol.Layer{{Visible: true, Parts: []protocol.Part{{Position: flight.Vector{3, 4, 0}, Scale: flight.Vector{2, 0, 0}}}}}
	if got := Radius(layers); got != 6 {
		t.Fatalf("got %v, want 5 + 2/2", got)
	}
}

func TestColoursAreReadAsSRGB(t *testing.T) {
	if got := linearColour("#ff0080"); got[0] != 1 || got[1] != 0 || math.Abs(float64(got[2])-0.2158605) > 1e-5 {
		t.Fatalf("got %v", got)
	}
	if got := linearColour("not a colour"); got == ([3]float32{}) {
		t.Fatal("a bad colour came out black, want the fallback grey")
	}
}

func TestTheCoarseDetailHasFarFewerTriangles(t *testing.T) {
	for _, kind := range []string{partSphere, partTorus, partCapsule, partCylinder} {
		fine, coarse := len(primitive(kind, Fine)), len(primitive(kind, Coarse))
		if coarse*2 > fine {
			t.Errorf("%s: %d coarse triangles for %d fine, want under half", kind, coarse, fine)
		}
	}
}

func TestTheCoarseDetailFillsTheSameCube(t *testing.T) {
	for _, kind := range everyType {
		low, high := math.Inf(1), math.Inf(-1)
		for _, corners := range primitive(kind, Coarse) {
			for _, corner := range corners {
				low, high = min(low, corner[0]), max(high, corner[0])
			}
		}
		if low < -0.5-slack || high > 0.5+slack {
			t.Errorf("%s spans %v to %v across, outside the unit cube", kind, low, high)
		}
	}
}

// heavyDesign is 400 spheres, all mirrored: the most the server lets a design have.
func heavyDesign() []protocol.Layer {
	parts := make([]protocol.Part, 400)
	for index := range parts {
		parts[index] = protocol.Part{Type: partSphere, Scale: flight.Vector{1, 1, 1}, Color: "#ffffff", Mirror: true}
	}
	return []protocol.Layer{{Visible: true, Parts: parts}}
}

func TestTrianglesCountsWhatBuildMakes(t *testing.T) {
	layers := heavyDesign()[:1]
	layers[0].Parts = layers[0].Parts[:3]
	if got, want := Triangles(layers, Coarse), Build(layers, Coarse).VertexCount()/3; got != want {
		t.Fatalf("Triangles says %d, Build made %d", got, want)
	}
}

func TestAHeavyDesignGetsAPlainerDetail(t *testing.T) {
	small := []protocol.Layer{{Visible: true, Parts: []protocol.Part{boxPart(flight.Vector{}, false)}}}
	if Within(small, 1000) != Fine {
		t.Fatal("a small design didn't keep the fine detail")
	}
	heavy := heavyDesign()
	detail := Within(heavy, 20000)
	if detail == Fine || Triangles(heavy, detail) > 20000 && detail != Minimal {
		t.Fatalf("got a detail of %d triangles", Triangles(heavy, detail))
	}
}

func TestAPartOfAnUnknownTypeIsLeftOutAndTheRestStillShows(t *testing.T) {
	unknown := boxPart(flight.Vector{3, 0, 0}, true)
	unknown.Type = "hologram"
	layers := []protocol.Layer{{Visible: true, Parts: []protocol.Part{unknown, boxPart(flight.Vector{}, false)}}}
	alone := []protocol.Layer{{Visible: true, Parts: []protocol.Part{boxPart(flight.Vector{}, false)}}}
	if got, want := len(Build(layers, Fine).Positions), len(Build(alone, Fine).Positions); got != want {
		t.Fatalf("the design made %d positions, want the box's %d", got, want)
	}
}

func TestTheBrowsersNewFieldsAreRead(t *testing.T) {
	var part protocol.Part
	text := `{"type":"light","color":"#fff1c4","intensity":2,"opacity":0.5,"texture":"carbon","textureUrl":"https://example.com/a.png","modelId":"m1","mirror":true}`
	if err := json.Unmarshal([]byte(text), &part); err != nil {
		t.Fatal(err)
	}
	if part.Intensity != 2 || part.Opacity != 0.5 || part.Texture != "carbon" || part.TextureURL == "" || part.ModelId != "m1" {
		t.Fatalf("read %+v", part)
	}
}

func TestALightIsABulbGlowingInItsColourMirroredLikeAnyPart(t *testing.T) {
	light := protocol.Part{Type: partLight, Position: flight.Vector{2, 0, 0}, Scale: flight.Vector{0.4, 0.4, 0.4}, Color: "#ff0000", Intensity: 2, Mirror: true}
	layers := []protocol.Layer{{Visible: true, Parts: []protocol.Part{boxPart(flight.Vector{}, false), light}}}
	model := BuildModel(layers, Fine)
	if len(model.Others) != 1 || model.Others[0].Material.Kind != scene.UnlitKind {
		t.Fatal("the light made no glowing mesh")
	}
	glow := model.Others[0]
	sphere := len(primitive(partSphere, Fine))
	if got := len(glow.Positions) / 9; got != 2*sphere {
		t.Errorf("the bulbs have %d triangles, want two spheres of %d", got, sphere)
	}
	if glow.Colors[0] != 1 || glow.Colors[1] != 0 {
		t.Errorf("the bulb's colour is %v, want red", glow.Colors[:3])
	}
	box := Build([]protocol.Layer{{Visible: true, Parts: []protocol.Part{boxPart(flight.Vector{}, false)}}}, Fine)
	if got, want := len(model.Body.Positions), len(box.Positions); got != want {
		t.Errorf("the body has %d positions, want only the box's %d", got, want)
	}
	if got, want := Triangles(layers, Fine), len(box.Positions)/9+2*sphere; got != want {
		t.Errorf("Triangles counts %d, want the box's and the bulbs' %d", got, want)
	}
}

func TestAnOpaqueDesignWithoutLightsIsOneMesh(t *testing.T) {
	if model := BuildModel([]protocol.Layer{{Visible: true, Parts: []protocol.Part{boxPart(flight.Vector{}, false)}}}, Fine); len(model.Others) != 0 {
		t.Fatalf("a plain design made %d meshes besides its body", len(model.Others))
	}
}

func TestASeeThroughPartIsBlendedAtItsOpacity(t *testing.T) {
	glass, frosted, alsoGlass := boxPart(flight.Vector{3, 0, 0}, false), boxPart(flight.Vector{-3, 0, 0}, false), boxPart(flight.Vector{0, 3, 0}, false)
	glass.Opacity, frosted.Opacity, alsoGlass.Opacity = 0.3, 0.7, 0.31
	layers := []protocol.Layer{{Visible: true, Parts: []protocol.Part{boxPart(flight.Vector{}, false), glass, frosted, alsoGlass}}}
	model := BuildModel(layers, Fine)
	box := len(primitive(partBox, Fine))
	if len(model.Body.Positions)/9 != box {
		t.Fatalf("the body has %d triangles, want only the opaque box's %d", len(model.Body.Positions)/9, box)
	}
	if len(model.Others) != 2 {
		t.Fatalf("%d see-through meshes, want one for 0.3 (with 0.31) and one for 0.7", len(model.Others))
	}
	for index, want := range []struct {
		alpha     float32
		triangles int
	}{{0.3, 2 * box}, {0.7, box}} {
		mesh := model.Others[index]
		if !mesh.Material.Transparent || math.Abs(float64(mesh.Material.BaseColor[3]-want.alpha)) > 1e-6 || len(mesh.Positions)/9 != want.triangles {
			t.Errorf("mesh %d: transparent %v, alpha %v, %d triangles; want alpha %v and %d", index, mesh.Material.Transparent, mesh.Material.BaseColor[3], len(mesh.Positions)/9, want.alpha, want.triangles)
		}
	}
}
