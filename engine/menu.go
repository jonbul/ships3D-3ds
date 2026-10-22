package main

import (
	"github.com/akrck02/go-renderer/draw2d"
	ui "github.com/akrck02/go-ui"

	"github.com/jonbul/ships3D-3ds/touchscreen"
)

// touchMenu is the menu that START opens on the touch screen, over the ship screen or the game,
// as in the hand-written program: pages with tabs that L and R (or a touch) go through.
type touchMenu struct {
	open  bool
	shown touchscreen.Menu
}

// toggle opens or closes the menu, with its sound; closing it saves the settings.
func (menu *touchMenu) toggle() {
	menu.open = !menu.open
	if menu.open {
		menu.shown.Reset() // it opens new, on its first page
		playSound(func(sound *soundSet) { sound.Play(sound.Choose, 1) })
		return
	}
	askedAccount = nil
	saveSettings()
	playSound(func(sound *soundSet) { sound.Play(sound.Back, 1) })
}

// askedAccount is set while the menu's account page asked for the account window, which a
// keyboard and a mouse fill (the desktop's, accountForm): it runs when the account changes. The
// window closes with the menu.
var askedAccount func()

// askedAccountForm is the account window, while the account page asks for it.
func askedAccountForm(context *ui.Context) {
	if askedAccount != nil {
		accountForm(context, askedAccount)
	}
}

// press acts on the menu's pages; B closes it.
func (menu *touchMenu) press(pages []touchscreen.Page, read touchscreen.Input, now float64) {
	read.Closes, read.TurnsPages = true, true
	sound, closed := menu.shown.Press(pages, read, panel(), now)
	if closed {
		menu.toggle()
		return
	}
	playMenuSound(sound)
}

// paint draws the menu's page in screen, over the night on a 3DS's touch screen.
func (menu *touchMenu) paint(drawing *draw2d.List, screen draw2d.Rectangle, pages []touchscreen.Page, now float64) {
	ground(drawing, now)
	menu.shown.Paint(drawing, screen, pages, menuHints.shown(), now)
}
