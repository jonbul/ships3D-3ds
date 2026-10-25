package main

import (
	"math"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/scene"

	"github.com/jonbul/ships3D-3ds/arena"
	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/ship"
)

// The hangar's turning ship: how big it shows, in metres, and how fast it turns, in turns a
// second.
const (
	hangarShipRadius   = 7
	hangarTurnsASecond = 0.12
)

// hangarScreen is the hangar on the second screen, as the 3DS program shows the ship to choose:
// a platform floating in space, lit by spotlights, with the ship turning on it.
type hangarScreen struct {
	scene  *engine.Scene
	camera *engine.Node
	hangar *arena.Hangar
	ship   *scene.Node
}

// openHangarScreen builds the hangar's scene and puts it on the second screen.
func openHangarScreen() *hangarScreen {
	shown, camera := secondScreenScene(arena.HangarCamera)
	built := arena.NewHangar()
	ship := &scene.Node{Name: "chosen ship", Hidden: true, Moving: true}
	world := shown.World()
	world.Nodes = append(append(world.Nodes, built.Nodes...), ship)
	engine.SetSecondScreen(shown)
	return &hangarScreen{scene: shown, camera: camera, hangar: built, ship: ship}
}

// show lights the hangar as space is, with its far dark behind, and turns the ship at now
// seconds; with no model (a ship the server hasn't described) only the hangar.
func (shown *hangarScreen) show(model ship.Model, radius, now float64) {
	environment := &shown.scene.World().Environment
	arena.LightSpace(environment)
	environment.SkyHorizon, environment.FogColor = arena.HangarBack, arena.HangarBack
	shown.hangar.Show(now)
	turnShip(shown.ship, model, radius, now)
	if shown.camera != nil {
		aim(shown.camera, arena.HangarCamera)
	}
}

// turnShip puts a ship on the hangar's platform, turning at now seconds, as big as the hangar
// shows ships; with no body, or no size, it hides it.
func turnShip(node *scene.Node, model ship.Model, radius, now float64) {
	node.Hidden = model.Body == nil || radius <= 0
	if node.Hidden {
		return
	}
	arena.Wear(node, model)
	turn := flight.FromEulerXYZ(0, now*hangarTurnsASecond*2*math.Pi, 0)
	node.Transform = arena.TransformOf(arena.HangarShipPlace, turn, hangarShipRadius/radius)
}
