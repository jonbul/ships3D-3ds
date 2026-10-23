package main

import "github.com/jonbul/ships3D-3ds/sounds"

// soundSet is the game's sounds (package sounds).
type soundSet = sounds.Sounds

// gameSound is the game's one sound, for every scene: opened once, with the music playing.
var gameSound *soundSet

// openSound opens the sound and starts the music; without it the game plays on in silence.
func openSound() {
	sound, err := sounds.New()
	if err != nil {
		println("no sound: " + err.Error()) // the debug log
	}
	gameSound = sound
	applyVolumes()
}

// playSound plays something on the game's sound, once it is open.
func playSound(play func(sound *soundSet)) {
	if gameSound != nil {
		play(gameSound)
	}
}
