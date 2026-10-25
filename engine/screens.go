package main

import (
	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/camera"
	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/arena"
	"github.com/jonbul/ships3D-3ds/match"
	"github.com/jonbul/ships3D-3ds/touchscreen"
)

// The scenes, in the order a player goes through them.
const (
	titleScene = "scenes/title.scene.json"
	modeScene  = "scenes/mode.scene.json"
	shipScene  = "scenes/ships.scene.json"
	arenaScene = "scenes/arena.scene.json"
)

// chosen is what the player has picked, carried from one scene to the next.
var chosen = struct {
	shipID string
	mode   match.Mode
}{shipID: defaultShip, mode: match.FreeForAll}

// stage gets a scene ready for one of the game's screens: the arena's space built in it, the
// camera seeing as far as the sky.
func stage(context *engine.Context) *arena.World {
	world := arena.NewWorldIn(context.Scene.World())
	world.Wanted = models.want
	world.LampsShine = engine.RunningOn() != engine.Nintendo3DSTarget
	context.Scene.SetPointLights(nil) // the scene before may have left its lamps
	shownBackdrop = backdrop(context)
	engine.SetSecondScreen(nil) // a screen that shows a scene there (the hangar, the radar) puts it back
	if camera, found := context.Scene.MainCamera(); found {
		if settings, isCamera := camera.Camera(); isCamera {
			settings.Near, settings.Far = arena.CameraNear, arena.CameraFar
			settings.Focus = arena.FocalDistance // the 3D screen's plane, as the hand-written program sets it
		}
	}
	return world
}

// look puts the scene's main camera where view is, seeing as wide.
func look(context *engine.Context, view camera.Lens) {
	if camera, found := context.Scene.MainCamera(); found {
		aim(camera, view)
	}
}

// Title is the first screen, as the 3DS opens: the name large over space, a ship crossing it with
// a barrel roll, a prompt that blinks.
type Title struct {
	world *arena.World
	now   float64
}

// Start builds the title's space; on the 3DS the touch screen shows the prompt.
func (title *Title) Start(context *engine.Context) {
	title.world = stage(context)
	drawInterface(nil)
	drawOver(func(drawing *draw2d.List, width, height float32) {
		paintScaled(drawing, width, height, title.paint)
	})
	if !oneScreen() {
		showTouchScreen(func(drawing *draw2d.List) { touchscreen.TitlePrompt(drawing, textMeasurer(), title.now) })
	}
}

// Update flies the title's ship, and goes on to the modes when the player says so.
func (title *Title) Update(context *engine.Context) {
	title.now = context.Time
	darkenSky(context.Scene)
	var view camera.Lens
	title.world.ShowTitle(context.Time, &view)
	look(context, view)
	if confirmed(context.Input) || touchInput(context.Input).Pressed&touchscreen.ButtonTouch != 0 {
		playSound(func(sound *soundSet) { sound.Play(sound.Choose, 1) })
		context.ChangeScene(modeScene)
	}
}

func (title *Title) paint(drawing *draw2d.List, width, height float32) {
	topHeading(drawing, width, "Ships 3D", platformCaption)
	if oneScreen() {
		prompt(drawing, startHint.shown(), width, height*0.75, title.now)
	}
}

// The modes a player chooses between, and the server's way to play each, in the same order.
var (
	modes = []touchscreen.Mode{
		{Name: "FREE FOR ALL", Text: "Every pilot for themselves", Icon: touchscreen.FreeForAllIcon, Open: true},
		{Name: "ASSAULT", Text: "Teams around a mothership", Icon: touchscreen.AssaultIcon, Open: true},
	}
	playedModes = []match.Mode{match.FreeForAll, match.Assault}
)

// Modes is choosing the game's mode: a window with a big icon for each, the star on the one
// chosen; the arrows (the cross) choose, Enter (A) goes on, Escape (B) back to the title.
type Modes struct {
	world  *arena.World
	chosen int
	now    float64
}

// Start builds the space behind the window; on the 3DS the modes are on the touch screen.
func (choice *Modes) Start(context *engine.Context) {
	choice.world = stage(context)
	drawInterface(nil)
	drawOver(func(drawing *draw2d.List, width, height float32) {
		paintScaled(drawing, width, height, choice.paint)
	})
	if !oneScreen() {
		showTouchScreen(func(drawing *draw2d.List) {
			ground(drawing, choice.now)
			touchscreen.ModeChoice(drawing, textMeasurer(), touchscreen.FullScreen, modes, choice.chosen, modeHints.shown(), choice.now)
		})
	}
}

// Update moves the choice, and goes on or back.
func (choice *Modes) Update(context *engine.Context) {
	choice.now = context.Time
	darkenSky(context.Scene)
	var view camera.Lens
	choice.world.ShowTitle(context.Time, &view)
	choice.world.HideTitle()
	look(context, view)
	keys := context.Input
	read := panelInput(keys)
	touched := choice.touched(read)
	if read.Pressed&(touchscreen.ButtonLeft|touchscreen.ButtonRight) != 0 {
		choice.chosen = (choice.chosen + 1) % len(modes)
		playSound(func(sound *soundSet) { sound.Play(sound.Move, 1) })
	}
	switch {
	case (confirmed(keys) || touched) && modes[choice.chosen].Open:
		chosen.mode = playedModes[choice.chosen]
		playSound(func(sound *soundSet) { sound.Play(sound.Choose, 1) })
		context.ChangeScene(shipScene)
	case cancelled(keys):
		playSound(func(sound *soundSet) { sound.Play(sound.Back, 1) })
		context.ChangeScene(titleScene)
	}
}

// touched chooses the mode whose card a touch landed on, and tells whether one did.
func (choice *Modes) touched(read touchscreen.Input) bool {
	if read.Pressed&touchscreen.ButtonTouch == 0 {
		return false
	}
	index := touchscreen.ModeTouched(panel(), read.TouchX, read.TouchY, len(modes))
	if index < 0 {
		return false
	}
	choice.chosen = index
	return true
}

func (choice *Modes) paint(drawing *draw2d.List, width, height float32) {
	if !oneScreen() {
		mode := modes[choice.chosen]
		topHeading(drawing, width, "Game mode", mode.Name)
		return
	}
	touchscreen.ModeChoice(drawing, textMeasurer(), panelIn(width, height), modes, choice.chosen, modeHints.shown(), choice.now)
}
