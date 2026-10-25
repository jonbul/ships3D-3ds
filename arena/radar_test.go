package arena

import (
	"testing"

	"github.com/jonbul/ships3D-3ds/radar"
)

func TestTheRadarShowsEveryShipWithItsStalkAsTheBrowsers(t *testing.T) {
	shown := NewRadar()
	shown.Show(map[string]radar.Blip{
		"above": {Ahead: 0.5, Above: 0.4, InRange: true},
		"level": {Right: 0.3, InRange: true},
	})
	blips, stalks := 0, 0
	for index := range shown.blips {
		if !shown.blips[index].Hidden {
			blips++
		}
		if !shown.stalks[index].Hidden {
			stalks++
		}
	}
	if blips != 2 || stalks != 1 {
		t.Fatalf("%d blips and %d stalks shown, want both ships and the stalk of the one above", blips, stalks)
	}
	for _, part := range []string{"shell", "equator", "meridian", "nose", "own"} {
		if !hasNode(shown, part) {
			t.Errorf("the radar has no %s, as the browser's has", part)
		}
	}
}

func hasNode(shown *Radar, name string) bool {
	for _, node := range shown.Scene.Nodes {
		if node.Name == name {
			return true
		}
	}
	return false
}
