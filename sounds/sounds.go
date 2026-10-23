package sounds

import (
	"time"

	sound "github.com/akrck02/go-sound"

	"github.com/jonbul/ships3D-3ds/chiptune"
	"github.com/jonbul/ships3D-3ds/game"
)

// Package sounds is the game's sound, through go-sound (the console's DSP): chiptune effects and music worked out in
// code (package chiptune). Each effect has a few players taken in turn, so a burst of shots
// overlaps instead of cutting itself off. On a console without the DSP's firmware
// (sdmc:/3ds/dspfirm.cdc, dumped once with the DSP1 homebrew) the game plays on in silence.
const (
	soundLatency     = 60 * time.Millisecond // short enough for a shot to sound with it
	playersPerEffect = 4
	OtherShotsVolume = 0.45 // others' shots, quieter than ours
)

// Effect is one sound and its players.
type Effect struct {
	players []*sound.Player
	next    int
}

// Sounds are the game's effects and music; all of it nil when the console has no sound.
type Sounds struct {
	context                       *sound.Context
	Laser, Hit, Explosion, Shield *Effect
	Move, Choose, Back            *Effect
	music                         *sound.Player
	engine                        *sound.Player
	engineSound                   *chiptune.Engine
	effectsVolume, musicVolume    float64
	edgeHeard                     float64 // the crash into the edge the force field's sound was played for
}

// stream gives go-sound a chiptune sound or the music.
type stream struct {
	source interface {
		Format() (sampleRate, channels int)
		Read(samples []float32) (int, error)
		Rewind() error
	}
}

func (played stream) Format() sound.Format {
	rate, channels := played.source.Format()
	return sound.Format{SampleRate: rate, Channels: channels}
}
func (played stream) Read(samples []float32) (int, error) { return played.source.Read(samples) }
func (played stream) Rewind() error                       { return played.source.Rewind() }

// New opens the console's sound, or returns silent sounds and why there is none.
func New() (*Sounds, error) {
	all := &Sounds{effectsVolume: 1, musicVolume: 1}
	context, err := sound.NewContext(sound.Options{Latency: soundLatency, Name: "Ships 3D"})
	if err != nil {
		return all, err
	}
	all.context = context
	for target, make := range map[**Effect]func() *chiptune.Sound{
		&all.Laser: chiptune.Laser, &all.Hit: chiptune.Hit, &all.Explosion: chiptune.Explosion, &all.Shield: chiptune.Shield,
		&all.Move: chiptune.Move, &all.Choose: chiptune.Choose, &all.Back: chiptune.Back,
	} {
		played := &Effect{}
		made := make()
		for range playersPerEffect {
			player, err := context.NewPlayer(stream{made.Copy()})
			if err != nil {
				return all, err
			}
			played.players = append(played.players, player)
		}
		*target = played
	}
	music, err := context.NewPlayer(stream{chiptune.NewMusic(1)})
	if err != nil {
		return all, err
	}
	all.music = music
	all.engineSound = chiptune.NewEngine(1)
	engine, err := context.NewPlayer(stream{all.engineSound})
	if err != nil {
		return all, err
	}
	all.engine = engine
	return all, nil
}

// Play starts an effect at a share of the effects' volume, on the player whose turn it is.
func (all *Sounds) Play(played *Effect, volume float64) {
	if all.context == nil || played == nil {
		return
	}
	player := played.players[played.next]
	played.next = (played.next + 1) % len(played.players)
	player.Rewind() //nolint:errcheck // a buffer rewinds without fail
	player.SetVolume(volume * all.effectsVolume)
	player.Play()
}

// PlayMusic starts the music, looping, or stops it at no volume.
func (all *Sounds) PlayMusic() {
	if all.music == nil {
		return
	}
	all.music.SetVolume(all.musicVolume * musicLevel)
	if all.musicVolume == 0 {
		all.music.Pause()
		return
	}
	if !all.music.IsPlaying() {
		all.music.SetLoop(true) //nolint:errcheck // the music rewinds without fail
		all.music.Play()
	}
}

// engineLevel is the engine's share of the effects' volume: always there, under the shots.
const engineLevel = 0.8

// RunEngine plays the engine at a throttle from 0 to 1 while flying, and stops it otherwise.
func (all *Sounds) RunEngine(flying bool, throttle float64) {
	if all.engine == nil {
		return
	}
	if !flying || all.effectsVolume == 0 {
		if all.engine.IsPlaying() {
			all.engine.Pause()
		}
		return
	}
	all.engineSound.Throttle = throttle
	all.engine.SetVolume(all.effectsVolume * engineLevel)
	if !all.engine.IsPlaying() {
		all.engine.Play()
	}
}

// musicLevel keeps the music under the effects.
const musicLevel = 0.55

// SetVolumes takes the settings' volumes, from 0 to 1.
func (all *Sounds) SetVolumes(effects, music float64) {
	all.effectsVolume, all.musicVolume = effects, music
	all.PlayMusic()
}

// MusicSeconds is how much of the music has been heard, for the log: it stops growing when the DSP
// stops being fed.
func (all *Sounds) MusicSeconds() float64 {
	if all.music == nil {
		return 0
	}
	return all.music.Position().Seconds()
}

// Hear plays the sounds of what happened in the game this frame: hits, blasts, and the force
// field when the ship runs into the world's edge anew.
func (all *Sounds) Hear(events []game.Event, state *game.Game) {
	for _, event := range events {
		switch event.Kind {
		case game.ShipHit:
			all.Play(all.Hit, 1)
		case game.ShipDestroyed:
			all.Play(all.Explosion, 1)
		}
	}
	if hit := state.EdgeHitAt; hit != 0 && hit != all.edgeHeard {
		all.edgeHeard = hit
		all.Play(all.Shield, 1)
	}
}
