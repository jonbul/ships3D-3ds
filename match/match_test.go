package match

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	gorilla "github.com/gorilla/websocket"

	"github.com/jonbul/ships3D-3ds/flight"
	"github.com/jonbul/ships3D-3ds/transport"
)

const welcome = `{"type":"welcome","id":"a","settings":{"tickRate":20,"shipRadius":4,"maxSpeed":80,"worldRadius":1500},
"players":[{"id":"a","name":"Pilot","ship":{"id":"default:arrow","name":"Arrow","layers":[]},
"scale":1,"life":10,"p":[0,0,0],"q":[0,0,0,1]}]}`

const lobby = `{"type":"lobby","id":"a","players":[{"id":"a","name":"Pilot"}],"minPlayers":2,"maxPlayers":8}`

// server welcomes whoever joins, and hands the join it got to joins.
func server(test *testing.T, joins chan<- map[string]any) (host string, port int) {
	return serverOn(test, FreeForAll, welcome, joins)
}

// serverOn answers whoever joins a mode with answer, and hands the join it got to joins.
func serverOn(test *testing.T, mode Mode, answer string, joins chan<- map[string]any) (host string, port int) {
	upgrader := gorilla.Upgrader{}
	served := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != string(mode) {
			http.NotFound(writer, request)
			return
		}
		connection, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer connection.Close() //nolint:errcheck // the test's server
		for {
			_, text, err := connection.ReadMessage()
			if err != nil {
				return
			}
			var message map[string]any
			if json.Unmarshal(text, &message) == nil && message["type"] == "join" {
				joins <- message
				connection.WriteMessage(gorilla.TextMessage, []byte(answer)) //nolint:errcheck // the test fails on its own
			}
		}
	}))
	test.Cleanup(served.Close)
	address, _ := url.Parse(served.URL)
	port, _ = strconv.Atoi(address.Port())
	return address.Hostname(), port
}

func TestAPlayerJoinsWithTheirShipAndIsWelcomed(test *testing.T) {
	joins := make(chan map[string]any, 1)
	host, port := server(test, joins)
	sent := 0
	played, err := Start(transport.Dial, host+":"+strconv.Itoa(port), FreeForAll, "", "default:arrow", "Pilot",
		Hooks{Sent: func(string) { sent++ }})
	if err != nil {
		test.Fatal(err)
	}
	defer played.Close()
	now := 0.0
	for deadline := time.Now().Add(3 * time.Second); played.Joining() && time.Now().Before(deadline); {
		now += 0.016
		if err := played.Exchange(flight.Controls{}, false, 0.016, now); err != nil {
			test.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if played.Joining() {
		test.Fatal("never welcomed")
	}
	join := <-joins
	if join["shipId"] != "default:arrow" || join["name"] != "Pilot" {
		test.Fatalf("joined with %v", join)
	}
	if sent == 0 {
		test.Fatal("the Sent hook heard nothing")
	}
}

func TestAServerWrittenBadlyIsRefused(test *testing.T) {
	if _, err := Start(transport.Dial, "not a server", FreeForAll, "", "default:arrow", "Pilot", Hooks{}); err == nil {
		test.Fatal("a bad address was accepted")
	}
}

func TestTheAssaultIsJoinedOnItsOwnPathAndWaitsInTheLobby(test *testing.T) {
	joins := make(chan map[string]any, 1)
	host, port := serverOn(test, Assault, lobby, joins)
	played, err := Start(transport.Dial, host+":"+strconv.Itoa(port), Assault, "", "default:arrow", "Pilot", Hooks{})
	if err != nil {
		test.Fatal(err)
	}
	defer played.Close()
	now := 0.0
	for deadline := time.Now().Add(3 * time.Second); played.Game.Lobby == nil && time.Now().Before(deadline); {
		now += 0.016
		if err := played.Exchange(flight.Controls{}, false, 0.016, now); err != nil {
			test.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if played.Game.Lobby == nil || played.Game.Lobby.MinPlayers != 2 {
		test.Fatalf("the lobby is %+v", played.Game.Lobby)
	}
}
