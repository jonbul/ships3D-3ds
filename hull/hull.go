// Package hull is a ship's shape as the boxes of its parts, for what runs into it: the assault's
// mothership, which never moves and is large enough that a sphere around it would be wrong. It
// mirrors ships3D-api's game/hull.go, which checks the same hits on the server.
package hull

import (
	"math"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/protocol"
)

// Hull is a design's parts as boxes in the world, the design at the origin with no rotation.
type Hull struct {
	boxes []box
}

// box is one part's bounding box: its centre, half its size along its own axes, and its turn.
type box struct {
	centre   flight.Vector
	half     flight.Vector
	rotation flight.Rotation
}

// New bounds every visible part of a design drawn scale times its size, mirrored parts twice, as
// the ships are drawn (package ship).
func New(layers []protocol.Layer, scale float64) Hull {
	var shape Hull
	for _, layer := range layers {
		if !layer.Visible {
			continue
		}
		for _, part := range layer.Parts {
			shape.boxes = append(shape.boxes, boxOf(part.Position, part.Rotation, part.Scale, scale))
			if part.Mirror {
				position, rotation := part.Position, part.Rotation
				mirrored := flight.Vector{-position[0], position[1], position[2]}
				shape.boxes = append(shape.boxes, boxOf(mirrored, flight.Vector{rotation[0], -rotation[1], -rotation[2]}, part.Scale, scale))
			}
		}
	}
	return shape
}

func boxOf(position, rotation, size flight.Vector, scale float64) box {
	return box{
		centre:   position.Times(scale),
		half:     size.Times(scale / 2),
		rotation: flight.FromEulerXYZ(rotation[0], rotation[1], rotation[2]),
	}
}

// SphereHits tells whether a sphere touches any part: a ship flying into the hull.
func (shape Hull) SphereHits(centre flight.Vector, radius float64) bool {
	for _, each := range shape.boxes {
		if each.distance(centre) <= radius {
			return true
		}
	}
	return false
}

// SegmentHits tells whether the segment from one point to another passes through any part: a
// bullet's step.
func (shape Hull) SegmentHits(from, to flight.Vector) bool {
	for _, each := range shape.boxes {
		if each.segmentHits(from, to) {
			return true
		}
	}
	return false
}

// local is a point of the world in the box's own frame, its centre at the origin.
func (each box) local(point flight.Vector) flight.Vector {
	return each.rotation.Inverse().Apply(point.Subtract(each.centre))
}

// distance is how far a point is from the box, 0 inside it.
func (each box) distance(point flight.Vector) float64 {
	inside := each.local(point)
	var outside flight.Vector
	for axis := range outside {
		outside[axis] = max(0, math.Abs(inside[axis])-each.half[axis])
	}
	return outside.Length()
}

// parallel is how little a segment may move along an axis to count as running parallel to it.
const parallel = 1e-12

// segmentHits is the slab test: the segment, as a share from 0 to 1 of the way along it, must be
// between each pair of the box's faces for one same stretch.
func (each box) segmentHits(from, to flight.Vector) bool {
	start, end := each.local(from), each.local(to)
	entered, left := 0.0, 1.0
	for axis := range start {
		along := end[axis] - start[axis]
		if math.Abs(along) < parallel {
			if math.Abs(start[axis]) > each.half[axis] {
				return false
			}
			continue
		}
		near, far := (-each.half[axis]-start[axis])/along, (each.half[axis]-start[axis])/along
		if near > far {
			near, far = far, near
		}
		entered, left = max(entered, near), min(left, far)
		if entered > left {
			return false
		}
	}
	return true
}
