package main

import (
	"github.com/akrck02/go-engine/engine"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/profile"
	"github.com/jonbul/ships3D-3ds/tilt"
)

// The keys are the engine's actions, bound per target in engine.json: the desktop's keyboard and
// pad as the browser's, the 3DS's buttons as the hand-written program's. The sticks are read here
// rather than bound, to soften them near their centre (fine aim), and so is the console's tilt.

// controlsOf reads this frame's actions as the flight model's controls, and whether the ship
// fires: the stick (or the arrows) for the nose and the turn, the roll and the throttle, and the
// console's tilt when steering by tilting. Inverting the pitch swaps the nose's way.
func controlsOf(keys *engine.Input, settings profile.Settings) (flight.Controls, bool) {
	stickX, stickY := keys.Stick()
	precision := float64(settings.Precision)
	pitch := keys.Action("pitch").Value - flight.Curve(stickY, precision)
	if settings.InvertPitch {
		pitch = -pitch
	}
	controls := flight.Controls{
		Pitch:    pitch,
		Yaw:      keys.Action("yaw").Value + flight.Curve(stickX, precision),
		Roll:     keys.Action("roll").Value,
		Throttle: keys.Action("throttle").Value,
	}
	if settings.Gyroscope {
		controls.Pitch += tilt.Axis(tilted.Pitch)
		controls.Yaw += tilt.Axis(tilted.Steer)
	}
	return controls, keys.Action("fire").Down
}

// turnTilt adds this frame's turning of the console up, and its action centres it, as in the 3DS
// program. The gyroscope's roll tips the top screen (toward the player is negative); its pitch
// turns the console like a wheel. A computer has no gyroscope: it reads still.
func turnTilt(keys *engine.Input, seconds float64) {
	turning = keys.Gyroscope()
	tilted.Turn(-turning.Roll, turning.Pitch, seconds)
	if keys.Action("centre tilt").Pressed {
		tilted.Centre()
	}
}

// menuPressed tells whether the player opened or closed the menu.
func menuPressed(keys *engine.Input) bool { return keys.Action("menu").Pressed }

// confirmed tells whether the player said yes this frame.
func confirmed(keys *engine.Input) bool { return keys.Action("confirm").Pressed }

// cancelled tells whether the player went back this frame.
func cancelled(keys *engine.Input) bool { return keys.Action("back").Pressed }
