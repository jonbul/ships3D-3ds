package hull

import (
	"math"
	"testing"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/protocol"
)

// slab is one part: a box 10 wide, 2 high and 4 deep at the origin, drawn twice its size.
func slab(rotation flight.Vector, mirror bool, position flight.Vector) []protocol.Layer {
	return []protocol.Layer{{Visible: true, Parts: []protocol.Part{{Type: "box", Position: position, Rotation: rotation, Scale: flight.Vector{10, 2, 4}, Mirror: mirror}}}}
}

func TestASphereTouchingAPartHitsTheHull(t *testing.T) {
	shape := New(slab(flight.Vector{}, false, flight.Vector{}), 2) // 20 × 4 × 8 in the world
	cases := []struct {
		name   string
		centre flight.Vector
		hits   bool
	}{
		{"inside", flight.Vector{0, 0, 0}, true},
		{"just over the top", flight.Vector{0, 2.9, 0}, true},
		{"above the top", flight.Vector{0, 3.1, 0}, false},
		{"past the end", flight.Vector{11.1, 0, 0}, false},
	}
	for _, each := range cases {
		if hits := shape.SphereHits(each.centre, 1); hits != each.hits {
			t.Errorf("%s: hits %v, want %v", each.name, hits, each.hits)
		}
	}
}

func TestATurnedPartIsHitWhereItTurned(t *testing.T) {
	shape := New(slab(flight.Vector{0, 0, math.Pi / 2}, false, flight.Vector{}), 1) // stood up: 2 wide, 10 high
	if !shape.SphereHits(flight.Vector{0, 4.5, 0}, 0.1) || shape.SphereHits(flight.Vector{4.5, 0, 0}, 0.1) {
		t.Fatal("the turned part is not where it turned")
	}
}

func TestAMirroredPartIsHitOnBothSides(t *testing.T) {
	shape := New(slab(flight.Vector{}, true, flight.Vector{20, 0, 0}), 1)
	if !shape.SphereHits(flight.Vector{20, 0, 0}, 0.1) || !shape.SphereHits(flight.Vector{-20, 0, 0}, 0.1) {
		t.Fatal("the mirrored part is missing a side")
	}
}

func TestAHiddenLayerIsNoHull(t *testing.T) {
	layers := slab(flight.Vector{}, false, flight.Vector{})
	layers[0].Visible = false
	if New(layers, 1).SphereHits(flight.Vector{}, 1) {
		t.Fatal("a hidden layer was hit")
	}
}

func TestABulletPassingThroughAPartHitsTheHull(t *testing.T) {
	shape := New(slab(flight.Vector{}, false, flight.Vector{}), 1)
	if !shape.SegmentHits(flight.Vector{0, -10, 0}, flight.Vector{0, 10, 0}) {
		t.Fatal("a bullet through the middle missed")
	}
	if shape.SegmentHits(flight.Vector{0, 5, -10}, flight.Vector{0, 5, 10}) {
		t.Fatal("a bullet passing above hit")
	}
}
