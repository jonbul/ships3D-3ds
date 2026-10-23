//go:build nintendo3ds

package main

import (
	"github.com/akrck02/go-3ds/keyboard"
	"github.com/akrck02/go-engine/engine"
	ui "github.com/akrck02/go-ui"

	"github.com/jonbul/ships3D-3ds/touchscreen"
)

// status is what the account screen says after a change: why a login failed.
var status string

// accountPage is the menu's account page on the 3DS, as the hand-written program has it: the
// server and the player, with choices that open the system keyboard. The keyboard takes both
// screens, so it opens between frames (engine.BetweenFrames). changed runs when either changes,
// for the ships to load again.
func accountPage(changed func()) touchscreen.Page {
	return touchscreen.AccountPage(touchscreen.Account{
		Server:   player.Server,
		Username: player.Username,
		ChangeServer: func() {
			engine.BetweenFrames(func() {
				typed, confirmed := keyboard.Ask(keyboard.Text, "Server: address:port, or https://name:port", player.Server, maximumServerLength)
				if confirmed && typed != player.Server {
					status = errorText(useServer(typed))
					changed()
				}
			})
		},
		LogIn: func() { engine.BetweenFrames(func() { askLogin(changed) }) },
		LogOut: func() {
			engine.BetweenFrames(func() {
				logOut()
				changed()
			})
		},
		Status: touchscreen.StatusRows(status),
	})
}

// accountForm is the desktop's account window, which the 3DS doesn't show: its account is a page
// of the menu (accountPage), and the hangar never draws the window there.
func accountForm(*ui.Context, func()) {}

// askLogin asks for the email and the password with the system keyboard, and logs in.
func askLogin(changed func()) {
	email, confirmed := keyboard.Ask(keyboard.Text, "Email", player.Email, maximumEmailLength)
	if !confirmed {
		return
	}
	password, confirmed := keyboard.Ask(keyboard.Password, "Password", "", maximumPasswordLength)
	if !confirmed {
		return
	}
	status = errorText(logIn(email, password))
	changed()
}

func errorText(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
