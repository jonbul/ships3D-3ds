package arena

import (
	"math"
	"testing"

	"github.com/akrck02/go-renderer/scene"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/ship"
)

func TestALampShinesWhereItsShipCarriesIt(test *testing.T) {
	world := &World{lamps: make([]scene.PointLight, 0, maximumLamps)}
	model := ship.Model{Lamps: []ship.Lamp{{Position: flight.Vector{1, 0, 0}, Colour: [3]float32{1, 0.5, 0}, Intensity: 2}}}
	world.AddLamps(model, TransformOf(flight.Vector{10, 0, 0}, flight.NoRotation, 3), 3)
	lamps := world.Lamps()
	if len(lamps) != 1 {
		test.Fatalf("%d lamps, want 1", len(lamps))
	}
	lamp := lamps[0]
	if math.Abs(float64(lamp.Position[0])-13) > 1e-5 || lamp.Color[3] != 2 || lamp.Outer != lampOuter*3 {
		test.Errorf("the lamp is %+v, want at x 13 (the ship at 10, the lamp 1 out at a scale of 3), intensity 2, reaching %v", lamp, lampOuter*3)
	}
}

func TestLampsBeyondTheirRoomStayDarkAndNothingIsAllocated(test *testing.T) {
	world := &World{lamps: make([]scene.PointLight, 0, maximumLamps)}
	model := ship.Model{Lamps: make([]ship.Lamp, 8)}
	place := TransformOf(flight.Vector{}, flight.NoRotation, 1)
	allocations := testing.AllocsPerRun(10, func() {
		world.ClearLamps()
		for range 20 { // twenty ships of eight lamps
			world.AddLamps(model, place, 1)
		}
	})
	if len(world.Lamps()) != maximumLamps || allocations > 0 {
		test.Errorf("%d lamps and %v allocations a frame, want %d and none", len(world.Lamps()), allocations, maximumLamps)
	}
}
