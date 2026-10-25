package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/akrck02/go-engine/engine"
)

// onlyOn are the actions bound on one target alone, and why: elsewhere a stick or nothing does it.
var onlyOn = map[string]string{
	"centre tilt": "nintendo3ds", // only a console has a tilt to centre
	"pitch":       "desktop",     // the 3DS steers with its circle pad, read as a stick
	"yaw":         "desktop",
}

// boundActions are engine.json's actions: for each, the keys bound on each target.
func boundActions(test *testing.T) map[string]map[string][]string {
	read, err := os.ReadFile("engine.json")
	if err != nil {
		test.Fatal(err)
	}
	var settings struct {
		Actions map[string]map[string][]string `json:"actions"`
	}
	if err := json.Unmarshal(read, &settings); err != nil {
		test.Fatal(err)
	}
	return settings.Actions
}

func TestEveryActionTheGameReadsIsBoundOnTheDesktopAndThe3DS(test *testing.T) {
	settings := struct {
		Actions map[string]map[string][]string
	}{boundActions(test)}
	sources, _ := filepath.Glob("*.go")
	checked := 0
	named := regexp.MustCompile(`Action\("([^"]+)"\)|\{"([^"]+)", touchscreen\.Button`)
	for _, source := range sources {
		code, err := os.ReadFile(source)
		if err != nil {
			test.Fatal(err)
		}
		for _, match := range named.FindAllStringSubmatch(string(code), -1) {
			name := match[1] + match[2]
			checked++
			for _, target := range []string{"desktop", "nintendo3ds"} {
				if only, found := onlyOn[name]; found && only != target {
					continue
				}
				if len(settings.Actions[name][target]) == 0 {
					test.Errorf("%s reads the action %q, which engine.json does not bind on %s", source, name, target)
				}
			}
		}
	}
	if checked == 0 {
		test.Fatal("no action found in the game's code: the test reads nothing")
	}
}

// flightActions are the actions read while flying: a key in two of them would do both at once, as
// the arrows once climbed and zoomed the radar together.
var flightActions = []string{"fire", "menu", "zoom in", "zoom out", "pitch", "yaw", "roll", "throttle"}

func TestNoKeyDoesTwoThingsInFlight(test *testing.T) {
	actions := boundActions(test)
	for _, target := range []string{"desktop", "nintendo3ds"} {
		owner := map[string]string{}
		for _, action := range flightActions {
			for _, key := range actions[action][target] {
				key = strings.TrimRight(key, "+-") // an axis's key, as "Up+"
				if other, taken := owner[key]; taken {
					test.Errorf("on %s, %s is bound to both %q and %q", target, key, other, action)
				}
				owner[key] = action
			}
		}
	}
}

func TestAHintLineNamesTheKeysBoundToItsActions(test *testing.T) {
	engine.BindActions(map[string]map[string][]string{
		"confirm": {"desktop": {"Enter"}}, "back": {"desktop": {"Escape"}},
	}, "desktop")
	line := &hintLine{parts: []hintPart{{[]string{"confirm"}, "OK"}, {[]string{"back"}, "BACK"}}}
	if got := line.made(&engine.Input{}); got != "ENTER OK   ESC BACK" {
		test.Errorf("the line says %q, want the keys bound, ENTER OK   ESC BACK", got)
	}
}
