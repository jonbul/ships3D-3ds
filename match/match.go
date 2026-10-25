// Package match is one game on the server, as any of the clients plays it: the connection (over
// TLS when the server is written with https://), the websocket, joining with a ship, and every
// frame the server's messages in and our ship's moves and shots out. It knows nothing of screens
// or sounds: those listen through Hooks.
package match

import (
	"strings"

	"github.com/jonbul/ships3D-3ds/account"
	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/game"
	"github.com/jonbul/ships3D-3ds/protocol"
	"github.com/jonbul/ships3D-3ds/transport"
	"github.com/jonbul/ships3D-3ds/websocket"
)

// Mode is a way to play, by the websocket path where the server takes its connections
// (ships3D-api's controllers/game.go).
type Mode string

// The modes: every pilot for themselves, and the assault, teams around a mothership.
const (
	FreeForAll Mode = "/ws"
	Assault    Mode = "/ws/assault"
)

// StateBacklogBytes is how much may wait to be sent before position updates are skipped: each
// replaces the one before, so sending a stale one behind a stalled network only adds to the pile.
const StateBacklogBytes = 2048

// How every position update and every shot begin (protocol.State, protocol.Fire).
const (
	stateStart = `{"type":"state"`
	fireStart  = `{"type":"fire"`
)

// Dial opens a connection to a host on a port, over TLS when secure. The clients pass
// transport.Dial; tests pass their own.
type Dial func(host string, port int, secure bool) (transport.Connection, error)

// Hooks are what a client hears of the match, for its log, its sounds and its figures; any may
// be nil.
type Hooks struct {
	Received func(message protocol.Incoming, bytes int)
	Sent     func(text string)
	// Fired is our own shot leaving: the client plays its sound.
	Fired func()
}

// Match is one game on the server.
type Match struct {
	Game       *game.Game
	Server     string // as the player wrote it
	connection transport.Connection
	client     *websocket.Client
	reader     protocol.Reader
	joined     bool
	shipID     string
	name       string
	hooks      Hooks
	// Skipped counts the position updates skipped while the network was behind.
	Skipped int
}

// Start connects to server (address:port, or https://name:port) to play mode, joining with the
// ship of shipID under name, as the player whose session cookie is cookie (empty for a guest).
// Joining happens once the websocket is open, in a later Exchange.
func Start(dial Dial, server string, mode Mode, cookie, shipID, name string, hooks Hooks) (*Match, error) {
	address, err := account.ParseAddress(server)
	if err != nil {
		return nil, err
	}
	connection, err := dial(address.Host, address.Port, address.Secure)
	if err != nil {
		return nil, err
	}
	var header map[string]string
	if cookie != "" {
		header = map[string]string{"Cookie": cookie}
	}
	client, err := websocket.Connect(connection, address.HostPort(), string(mode), header)
	if err != nil {
		_ = connection.Close() // the handshake's error is the one to report
		return nil, err
	}
	return &Match{Game: game.New(), Server: server, connection: connection, client: client,
		shipID: shipID, name: name, hooks: hooks}, nil
}

// Exchange moves the match on one frame of seconds, at now: what the server sent goes into the
// game, the ship flies with controls and fires when firing, and what the server must hear is
// sent. An error means the match is over: the connection is lost.
func (match *Match) Exchange(controls flight.Controls, firing bool, seconds, now float64) error {
	if err := match.connection.Flush(); err != nil {
		return err
	}
	if err := match.receive(now); err != nil {
		return err
	}
	if !match.client.Open() {
		return nil
	}
	if !match.joined {
		match.joined = true
		if err := match.send(protocol.Join(match.shipID, match.name)); err != nil {
			return err
		}
	}
	for _, text := range match.Game.Step(controls, firing, firing, seconds, now) {
		if err := match.send(text); err != nil {
			return err
		}
	}
	return nil
}

// Joining tells whether the match waits for the server's welcome.
func (match *Match) Joining() bool { return !match.Game.Welcomed }

// Pending is how many bytes wait to be sent.
func (match *Match) Pending() int { return match.connection.Pending() }

// Close leaves the match.
func (match *Match) Close() {
	_ = match.client.Close()     // leaving anyway; the server notices the socket close too
	_ = match.connection.Close() // as above
}

// receive hands the game what the server sent since the last frame.
func (match *Match) receive(now float64) error {
	messages, err := match.client.Receive()
	for _, text := range messages {
		message, states, readErr := match.reader.Read(text)
		if readErr != nil {
			continue // one bad message doesn't end the match
		}
		if match.hooks.Received != nil {
			match.hooks.Received(message, len(text))
		}
		match.Game.Handle(message, states, now)
	}
	return err
}

func (match *Match) send(text string) error {
	if strings.HasPrefix(text, stateStart) && match.connection.Pending() > StateBacklogBytes {
		match.Skipped++
		return nil
	}
	if strings.HasPrefix(text, fireStart) && match.hooks.Fired != nil {
		match.hooks.Fired()
	}
	if match.hooks.Sent != nil {
		match.hooks.Sent(text)
	}
	return match.client.Send(text)
}
