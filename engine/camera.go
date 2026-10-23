package main

import (
	"log/slog"
	"math"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-engine/scenefile"
	"github.com/akrck02/go-engine/space"
	"github.com/akrck02/go-renderer/camera"
)

// degreesPerRadian turns go-renderer's angles into the engine's.
const degreesPerRadian = 180 / math.Pi

// aim puts a camera where a go-renderer camera (arena's views) is, seeing as wide.
func aim(camera *engine.Node, view camera.Lens) {
	eye := vectorOf(view.Position[0], view.Position[1], view.Position[2])
	target := vectorOf(view.Target[0], view.Target[1], view.Target[2])
	up := vectorOf(view.Up[0], view.Up[1], view.Up[2])
	camera.Transform.Position = eye
	camera.Transform.Rotation = space.RotationLookingAt(target.Subtract(eye), up)
	if settings, isCamera := camera.Camera(); isCamera {
		settings.FieldOfView = fieldOfView(view) * degreesPerRadian
	}
}

// rendererFieldOfView is go-renderer's field of view when a camera leaves it at 0, in radians.
const rendererFieldOfView = math.Pi / 4

// fieldOfView is how wide a go-renderer camera sees, in radians, as go-renderer reads it.
func fieldOfView(view camera.Lens) float64 {
	if view.FieldOfViewRadians > 0 {
		return float64(view.FieldOfViewRadians)
	}
	return rendererFieldOfView
}

func vectorOf(x, y, z float32) space.Vector {
	return space.NewVector(float64(x), float64(y), float64(z))
}

// secondScreenScene makes a scene for the second screen (the radar, the hangar), with a camera
// seeing as view does; the camera is nil if it could not be made.
func secondScreenScene(view camera.Lens) (*engine.Scene, *engine.Node) {
	shown := engine.NewScene(files)
	settings := &scenefile.Camera{Main: true, Near: float64(view.Near), Far: float64(view.Far)}
	camera, err := shown.NewNode(scenefile.Node{Name: "Camera", Camera: settings})
	if err != nil {
		slog.Warn("the second screen's camera could not be made: it stays empty", "error", err)
		return shown, nil
	}
	shown.Root.Add(camera)
	return shown, camera
}
