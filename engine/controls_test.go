package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// onlyOn are the actions bound on one target alone, and why: elsewhere a stick or nothing does it.
var onlyOn = map[string]string{
	"centre tilt": "nintendo3ds", // only a console has a tilt to centre
	"pitch":       "desktop",     // the 3DS steers with its circle pad, read as a stick
	"yaw":         "desktop",
}

func TestEveryActionTheGameReadsIsBoundOnTheDesktopAndThe3DS(test *testing.T) {
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
