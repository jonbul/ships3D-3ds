package main

import (
	"strconv"

	"github.com/akrck02/go-engine/engine"
	"github.com/akrck02/go-renderer/draw2d"
	"github.com/akrck02/go-renderer/models"
	"github.com/akrck02/go-renderer/scene"
	ui "github.com/akrck02/go-ui"

	"github.com/jonbul/ships3D-3ds/arena"
	"github.com/jonbul/ships3D-3ds/pixelart"
	"github.com/jonbul/ships3D-3ds/profile"
	"github.com/jonbul/ships3D-3ds/protocol"
	"github.com/jonbul/ships3D-3ds/ship"
	"github.com/jonbul/ships3D-3ds/touchscreen"
)

// The ship screen's measures.
const (
	shipNameCharacters  = 18
	arrowsFromTheMiddle = 150
)

// Ships is choosing the ship, as the browser and the 3DS show it: the ship turning on a platform
// floating in space, the arrows (the cross, the circle pad, or a touch) to go through the
// player's ships, Enter (A) to launch, Escape (B) back to the modes. On the 3DS the hangar is on
// the touch screen and the top screen says the ship's name, as the hand-written program does; on
// the desktop the hangar fills the window.
type Ships struct {
	world   *arena.World
	hangar  *arena.Hangar // on the desktop, in the scene played
	below   *hangarScreen // on the 3DS, on the touch screen
	turning *scene.Node
	ships   []protocol.Ship
	own     int // how many of the ships, first, are the player's own designs
	index   int
	status  string
	reload  bool      // the server or the account changed: load the ships again
	menu    touchMenu // the account, the settings and the controls, with START (Tab)
	now     float64
}

// Start builds the hangar and asks the server for its ships.
func (choice *Ships) Start(context *engine.Context) {
	choice.world = stage(context)
	choice.load()
	drawOver(func(drawing *draw2d.List, width, height float32) {
		paintScaled(drawing, width, height, choice.paint)
	})
	if !oneScreen() {
		choice.below = openHangarScreen()
		showTouchScreen(choice.paintTouchScreen)
		return
	}
	choice.hangar = arena.NewHangar()
	world := context.Scene.World()
	world.Nodes = append(world.Nodes, choice.hangar.Nodes...)
	choice.turning = &scene.Node{Name: "chosen ship", Moving: true}
	world.Nodes = append(world.Nodes, choice.turning)
	drawInterface(func(context *ui.Context) {
		if !choice.menu.open || askedAccount != nil { // the window would cover the menu's pages
			accountForm(context, func() { choice.reload = true })
		}
	})
}

// load asks the server for the player's ships and puts the last one flown first.
func (choice *Ships) load() {
	choice.ships, choice.own, choice.status = fetchShips()
	choice.index = 0
	for index, offered := range choice.ships {
		if offered.Id == player.Ship {
			choice.index = index
		}
	}
}

// fetchShips is the server's ships for the player: their own designs first (own is how many),
// then the built-in ones, or only those for a guest; when the server can't be reached, the
// Arrow, and why.
func fetchShips() (ships []protocol.Ship, own int, status string) {
	fallback := []protocol.Ship{{Id: defaultShip, Name: "Arrow"}}
	server, err := api()
	if err != nil {
		return fallback, 0, err.Error()
	}
	offered, err := server.Ships(session())
	if err != nil || len(offered.Defaults)+len(offered.Own) == 0 {
		return fallback, 0, "The server's ships didn't load"
	}
	return append(offered.Own, offered.Defaults...), len(offered.Own), ""
}

// Update turns the ship, and goes through the ships, on or back.
func (choice *Ships) Update(context *engine.Context) {
	choice.now = context.Time
	if choice.reload {
		choice.reload = false
		choice.load()
	}
	darkenSky(context.Scene)
	choice.show(context)
	choice.press(context, panelInput(context.Input))
}

// show puts the hangar and the chosen ship where this platform shows them.
func (choice *Ships) show(context *engine.Context) {
	offered := choice.ships[choice.index]
	mesh := choice.world.ShipMesh(offered.Id, offered.Layers, false)
	radius := ship.Radius(offered.Layers)
	if choice.below != nil {
		var view models.Camera
		choice.world.ShowTitle(context.Time, &view)
		choice.world.HideTitle() // the ship turns on the touch screen; the top shows its name
		look(context, view)
		choice.below.show(mesh, radius, context.Time)
		return
	}
	choice.hangar.Show(context.Time)
	choice.world.HideTitle()
	look(context, arena.HangarCamera)
	turnShip(choice.turning, mesh, radius, context.Time)
}

// paintTouchScreen draws the arrows over the hangar, or the menu over it when it is open.
func (choice *Ships) paintTouchScreen(drawing *draw2d.List) {
	if choice.menu.open {
		choice.menu.paint(drawing, touchscreen.FullScreen, choice.pages(), choice.now)
		return
	}
	touchscreen.ShipChoice(drawing, textMeasurer(), choice.status)
}

// pages are the ship screen's menu: the account, the settings and the controls.
func (choice *Ships) pages() []touchscreen.Page {
	return []touchscreen.Page{
		accountPage(func() { choice.reload = true }),
		settingsPage(),
		touchscreen.ControlsPage(touchscreen.ControlLines),
	}
}

// press goes through the ships, launches the chosen one, or goes back to the modes.
func (choice *Ships) press(context *engine.Context, read touchscreen.Input) {
	keys := context.Input
	if menuPressed(keys) {
		choice.menu.toggle()
		return
	}
	if choice.menu.open {
		choice.menu.press(choice.pages(), read, choice.now)
		return
	}
	touchedSide := 0
	if read.Pressed&touchscreen.ButtonTouch != 0 {
		touchedSide = touchscreen.ShipArrowTouched(read.TouchX)
	}
	switch {
	case read.Pressed&touchscreen.ButtonLeft != 0 || touchedSide < 0:
		choice.turn(-1)
	case read.Pressed&touchscreen.ButtonRight != 0 || touchedSide > 0:
		choice.turn(1)
	case confirmed(keys):
		chosen.shipID = choice.ships[choice.index].Id
		remember(func(remembered *profile.Profile) { remembered.Ship = chosen.shipID })
		playSound(func(sound *soundSet) { sound.Play(sound.Choose, 1) })
		context.ChangeScene(arenaScene)
	case cancelled(keys):
		playSound(func(sound *soundSet) { sound.Play(sound.Back, 1) })
		context.ChangeScene(modeScene)
	}
}

func (choice *Ships) turn(by int) {
	choice.index = (choice.index + by + len(choice.ships)) % len(choice.ships)
	playSound(func(sound *soundSet) { sound.Play(sound.Move, 1) })
}

// caption says whose the chosen ship is, and which of how many.
func (choice *Ships) caption() string {
	whose := "Built-in ship"
	if choice.index < choice.own {
		whose = "Your design"
	}
	return whose + "   " + strconv.Itoa(choice.index+1) + " / " + strconv.Itoa(len(choice.ships))
}

func (choice *Ships) paint(drawing *draw2d.List, width, height float32) {
	name := touchscreen.Fitted(choice.ships[choice.index].Name, shipNameCharacters)
	if choice.below != nil {
		topHeading(drawing, width, name, choice.caption())
		return
	}
	centred(drawing, name, width, 30, headingSize*1.5, draw2d.Bold, colours.ink)
	count := strconv.Itoa(choice.index+1) + " / " + strconv.Itoa(len(choice.ships))
	centred(drawing, count, width, 46, pixelart.LabelSize, draw2d.Mono, colours.soft)
	middle := height / 2
	drawing.Text("<", width/2-arrowsFromTheMiddle, middle, headingSize, draw2d.Bold, colours.accent)
	drawing.Text(">", width/2+arrowsFromTheMiddle-8, middle, headingSize, draw2d.Bold, colours.accent)
	who := "Guest"
	if player.LoggedIn() {
		who = "Playing as " + player.Username
	}
	centred(drawing, who, width, 60, pixelart.TextSize, draw2d.Regular, colours.soft)
	if choice.status != "" {
		centred(drawing, choice.status, width, height-30, pixelart.TextSize, draw2d.Regular, colours.soft)
	}
	if choice.menu.open {
		choice.menu.paint(drawing, panelIn(width, height), choice.pages(), choice.now)
		return
	}
	prompt(drawing, launchHint.shown(), width, height-20, choice.now)
	hints(drawing, shipHints.shown(), width, height)
}
