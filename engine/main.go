// Ships 3D, made with go-engine: the same game as the browser's and the 3DS program's, built for
// the desktop and the 3DS from one project. The editor keeps the scenes in scenes/; the game's
// code is the Go beside this file and the repository's packages (arena, match, game…). Play it
// with `go run ./engine`.
package main

import (
	"embed"
	"log"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/jonbul/ships3D-3ds/pixelart"
)

// drawnOver paints what the scene playing now shows over its 3D view (the HUD, a menu); each
// scene's behaviour sets it as it starts (drawOver).
var drawnOver func(drawing *draw2d.List, width, height float32)

// drawOver makes paint what shows over the 3D view, until another scene's behaviour changes it.
func drawOver(paint func(drawing *draw2d.List, width, height float32)) { drawnOver = paint }

// drawnInterface is the scene playing now's go-ui interface (the hangar's account), or none: on
// the 3DS the bottom screen, on the desktop over the window.
var drawnInterface func(context *ui.Context)

// gameTheme is go-ui's dark theme in the game's look.
var gameTheme = pixelart.Theme(ui.DarkTheme())

// drawInterface makes paint the interface, until another scene's behaviour changes it.
func drawInterface(paint func(context *ui.Context)) { drawnInterface = paint }

// paintInterface is go-engine's game interface, every frame: what the scene playing shows.
func paintInterface(context *ui.Context, _ engine.Menu) {
	context.Theme = gameTheme
	if drawnInterface != nil {
		drawnInterface(context)
	}
}

// paintOverlay is go-engine's 2D layer over the 3D, every frame: what the scene playing paints.
func paintOverlay(drawing *draw2d.List, width, height float32) {
	windowWidth, windowHeight = width, height
	if drawnOver != nil {
		drawnOver(drawing, width, height)
	}
}

// files are the project's settings, scenes and assets, carried inside the game.
//
//go:embed engine.json scenes all:assets
var files embed.FS

func main() {
	engine.SetFonts(pixelart.Fonts())
	loadPlayer()
	engine.SetOverlay(paintOverlay)
	engine.SetInterface(paintInterface)
	openSound()
	// Each behaviour the scenes use is registered here, by its type's name.
	engine.Register(Title{}, Modes{}, Ships{}, Pilot{})
	if err := engine.Run(files); err != nil {
		log.Fatal(err)
	}
}
