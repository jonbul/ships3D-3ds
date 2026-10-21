package main

import (
	"github.com/akrck02/go-engine/engine"

	"github.com/jonbul/ships3D-3ds/arena"
	"github.com/jonbul/ships3D-3ds/game"
	"github.com/jonbul/ships3D-3ds/radar"
)

// radarScreen is the 3D radar on the second screen (the 3DS's bottom one), as in the 3DS program:
// a scene under the interface, its camera above and behind the sphere. The desktop has no second
// screen: there the same scene is placed in a corner of the window, as the browser's radar is
// (paintCornerRadar), or the flat map when it is chosen.
type radarScreen struct {
	scene   *engine.Scene
	camera  *engine.Node
	sphere    *arena.Radar
	scanner *radar.Scanner
	blips   map[string]radar.Blip
	shown   bool // on the second screen: not while the flat map is chosen
}

// openRadar builds the radar's scene and puts it on the second screen.
func openRadar() *radarScreen {
	bottom, camera := secondScreenScene(arena.RadarView)
	shown := &radarScreen{scene: bottom, camera: camera, sphere: arena.NewRadarIn(bottom.World()), scanner: radar.NewScanner(), blips: map[string]radar.Blip{}, shown: true}
	engine.SetSecondScreen(bottom)
	return shown
}

// show places the ships as far as the zoom (an index in radar.Ranges) reaches: on the 3D sphere, as
// the browser shows them, and for the flat map, whose sweep sees them as it passes.
func (shown *radarScreen) show(state *game.Game, now float64, zoom int, flat bool) {
	if flat == shown.shown {
		shown.shown = !flat
		shown.putOnScreen()
	}
	clear(shown.blips)
	for id, remote := range state.Remotes {
		if !remote.Player.Dead {
			shown.blips[id] = radar.Place(remote.Shown.Position, state.Own.Position, state.Own.Rotation, radar.Ranges[zoom])
		}
	}
	shown.scanner.Update(now, shown.blips)
	shown.sphere.Show(shown.blips)
	if shown.camera != nil {
		aim(shown.camera, arena.RadarView)
	}
}

// putOnScreen puts the sphere on the second screen when it shows, or takes it off.
func (shown *radarScreen) putOnScreen() {
	if shown.shown {
		engine.SetSecondScreen(shown.scene)
		return
	}
	engine.SetSecondScreen(nil)
}

// close takes the radar off the second screen, as the match ends.
func (*radarScreen) close() { engine.SetSecondScreen(nil) }
