package main

import (
	"github.com/jonbul/ships3D-3ds/account"
	"github.com/jonbul/ships3D-3ds/profile"
	"github.com/jonbul/ships3D-3ds/transport"
)

// player is what is remembered of the player: the server, the session when logged in, the last
// ship and the settings. It is the same profile the 3DS program keeps, read at start (main) and
// saved when it changes (remember).
var player profile.Profile

// loadPlayer reads the saved profile, or starts one with the real server.
func loadPlayer() {
	saved, err := readProfile()
	if err != nil {
		player = profile.New(realServer)
		return
	}
	player = profile.Read(saved, realServer)
}

// remember changes the profile and saves it.
func remember(change func(remembered *profile.Profile)) {
	change(&player)
	if err := saveProfile(player.Bytes()); err != nil {
		println(err.Error()) // the debug log: playing on matters more than remembering
	}
}

// api is the player's server's HTTP side, for logging in and the list of ships.
func api() (account.Server, error) {
	address, err := account.ParseAddress(player.Server)
	if err != nil {
		return account.Server{}, err
	}
	return account.Server{Host: address.HostPort(), Dial: func() (account.Connection, error) {
		return transport.Dial(address.Host, address.Port, address.Secure)
	}}, nil
}

// session is the player's session on the server: empty for a guest.
func session() account.Session {
	return account.Session{Token: player.Token, Username: player.Username}
}

// logIn logs the player in and remembers the session; the password isn't kept.
func logIn(email, password string) error {
	server, err := api()
	if err != nil {
		return err
	}
	logged, err := server.LogIn(email, password)
	if err != nil {
		return err
	}
	remember(func(remembered *profile.Profile) {
		remembered.Email, remembered.Token, remembered.Username = email, logged.Token, logged.Username
	})
	return nil
}

// logOut forgets the session, on the server too when it answers.
func logOut() {
	if server, err := api(); err == nil {
		_ = server.LogOut(session()) // the session is forgotten here whatever the server says
	}
	remember(func(remembered *profile.Profile) { remembered.Token, remembered.Username = "", "" })
}

// useServer changes the server; a session belongs to the server it was made on.
func useServer(server string) error {
	if _, err := account.ParseAddress(server); err != nil {
		return err
	}
	remember(func(remembered *profile.Profile) {
		remembered.Server, remembered.Token, remembered.Username = server, "", ""
	})
	return nil
}

// The most characters the account's fields take.
const (
	maximumServerLength   = 64 // a name or an address and its port
	maximumEmailLength    = 100
	maximumPasswordLength = 64
)
