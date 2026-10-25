package main

import (
	"github.com/akrck02/go-engine/engine"

	"github.com/jonbul/ships3D-3ds/profile"
	"github.com/jonbul/ships3D-3ds/tilt"
	"github.com/jonbul/ships3D-3ds/touchscreen"
)

// tilted is how far the console has been tilted since it was centred, for steering by tilting:
// centred when that is switched on, or with L in the game.
var tilted tilt.Tilt

// turning is how fast the console turns this frame, in degrees a second, for the figures.
var turning engine.Turn

// unsaved is whether the settings changed since they were last saved: they are saved when the
// menu closes, as writing the SD card on every step of a volume would stall the game.
var unsaved bool

// settingsPage is the menu's page of how the ship flies, the radar shows and the game sounds.
func settingsPage() touchscreen.Page {
	change := func(apply func(settings *profile.Settings)) {
		apply(&player.Settings)
		unsaved = true
		applyVolumes()
	}
	return touchscreen.SettingsPage(player.Settings, change, tilted.Centre)
}

// saveSettings saves the settings if they changed.
func saveSettings() {
	if !unsaved {
		return
	}
	unsaved = false
	remember(func(*profile.Profile) {}) // the settings are already in the profile: this saves it
}

// applyVolumes plays the sound at the settings' volumes.
func applyVolumes() {
	playSound(func(sound *soundSet) {
		sound.SetVolumes(float64(player.Settings.Effects), float64(player.Settings.Music))
	})
}
