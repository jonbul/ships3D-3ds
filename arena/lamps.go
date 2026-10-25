package arena

import (
	"github.com/akrck02/go-renderer/graphics"
	"github.com/akrck02/go-renderer/scene"

	"github.com/jonbul/ships3D-3ds/ship"
)

// Ships' lamps: the light parts of their designs, shining on what is near them. The browser's
// lamps fade with the square of the distance and never end; these end at lampOuter, so the few
// lights a draw takes (go-renderer's four) are the ones that reach it.
const (
	lampInner = 0.5 // design units: full light within
	lampOuter = 6   // design units: no light beyond
	// maximumLamps is how many lamps a frame lights, made room for once so the list never grows:
	// the browser allows eight a ship.
	maximumLamps = 64
)

// ClearLamps starts the frame's lamps again.
func (world *World) ClearLamps() { world.lamps = world.lamps[:0] }

// AddLamps adds a model's lamps, placed by its node's transform, scale being how large a unit of
// the design is in the world. Past maximumLamps the rest stay dark.
func (world *World) AddLamps(model ship.Model, transform graphics.Matrix, scale float64) {
	for _, lamp := range model.Lamps {
		if len(world.lamps) == cap(world.lamps) {
			return
		}
		place := transform.Apply(graphics.Vector{float32(lamp.Position[0]), float32(lamp.Position[1]), float32(lamp.Position[2]), 1})
		world.lamps = append(world.lamps, scene.PointLight{
			Position: place,
			Color:    graphics.Vector{lamp.Colour[0], lamp.Colour[1], lamp.Colour[2], float32(lamp.Intensity)},
			Inner:    float32(lampInner * scale),
			Outer:    float32(lampOuter * scale),
		})
	}
}

// Lamps are the frame's lamps, for the engine to light the scene with.
func (world *World) Lamps() []scene.PointLight { return world.lamps }
