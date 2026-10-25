package protocol

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/jonbul/ships3D-3ds/flight"
)

func TestAWelcomeIsRead(t *testing.T) {
	text := `{"type":"welcome","id":"a","settings":{"tickRate":20,"shipRadius":4,"maxSpeed":80},
	"players":[{"id":"a","name":"Guest","ship":{"id":"default:arrow","name":"Arrow","layers":[{"name":"Hull","visible":true,
	"parts":[{"type":"box","position":[0,0,0],"rotation":[0,0,0],"scale":[1,2,3],"color":"#ffffff","mirror":true}]}]},
	"scale":0.5,"life":10,"p":[1,2,3],"q":[0,0,0,1]}]}`
	message, _, err := (&Reader{}).Read(text)
	if err != nil {
		t.Fatal(err)
	}
	if message.Settings.ShipRadius != 4 || message.Settings.MaxSpeed != 80 {
		t.Fatalf("settings %+v", message.Settings)
	}
	part := message.Players[0].Ship.Layers[0].Parts[0]
	if part.Scale != (flight.Vector{1, 2, 3}) || !part.Mirror || message.Players[0].Position != (flight.Vector{1, 2, 3}) {
		t.Fatalf("player %+v", message.Players[0])
	}
}

func TestASnapshotGivesItsMovedPlayers(t *testing.T) {
	message, states, err := (&Reader{}).Read(`{"type":"snapshot","players":[{"id":"b","p":[1,0,0],"q":[0,0,0,1],"v":[0,0,5]}]}`)
	if err != nil || message.Type != TypeSnapshot || len(states) != 1 || states[0].Velocity[2] != 5 {
		t.Fatalf("got %+v %+v %v", message, states, err)
	}
}

func TestOutgoingMessagesUseTheServersFieldNames(t *testing.T) {
	cases := map[string][]string{
		Join("default:arrow", "3DS"):                               {"type", "shipId", "name"},
		State(flight.Vector{}, flight.NoRotation, flight.Vector{}): {"type", "p", "q", "v"},
		Fire("1", flight.Vector{}, flight.Forward):                 {"type", "id", "p", "d"},
		Hit("1", "b"): {"type", "bulletId", "target"},
		Respawn():     {"type"},
	}
	for text, fields := range cases {
		var decoded map[string]any
		if err := json.Unmarshal([]byte(text), &decoded); err != nil {
			t.Fatal(err)
		}
		for _, field := range fields {
			if _, found := decoded[field]; !found {
				t.Errorf("%s has no %q", text, field)
			}
		}
	}
}

func TestBrokenJsonHasItsCode(t *testing.T) {
	if _, _, err := (&Reader{}).Read("{not json"); !errors.Is(err, ErrUnreadable) {
		t.Fatalf("got %v, want %v", err, ErrUnreadable)
	}
	if ErrUnreadable.Error() != "protocol.unreadable-message" {
		t.Fatalf("code %q", ErrUnreadable.Error())
	}
}

func TestASnapshotWithTypeLastIsStillRead(t *testing.T) {
	_, states, err := (&Reader{}).Read(`{"players":[{"id":"b","p":[1,0,0],"q":[0,0,0,1],"v":[0,0,0]}],"type":"snapshot"}`)
	if err != nil || len(states) != 1 || states[0].Id != "b" {
		t.Fatalf("got %+v, %v", states, err)
	}
}

func TestACrashNamesTheShipRunInto(t *testing.T) {
	if message := Crash("them"); message != `{"type":"crash","with":"them"}` {
		t.Fatalf("crash is %s", message)
	}
}

func TestALobbyIsRead(t *testing.T) {
	text := `{"type":"lobby","id":"b","players":[{"id":"a","name":"Ann","guest":true},{"id":"b","name":"Ben","guest":false}],"minPlayers":2,"maxPlayers":8,"startsInMs":5000}`
	message, _, err := (&Reader{}).Read(text)
	if err != nil {
		t.Fatal(err)
	}
	if len(message.Players) != 2 || message.Players[1].Name != "Ben" || message.MinPlayers != 2 || message.MaxPlayers != 8 {
		t.Fatalf("lobby %+v", message)
	}
	if message.StartsInMs == nil || *message.StartsInMs != 5000 {
		t.Fatalf("starts in %v, want 5000 ms", message.StartsInMs)
	}
}

func TestALobbyThatIsNotCountingDownHasNoStart(t *testing.T) {
	message, _, err := (&Reader{}).Read(`{"type":"lobby","id":"a","players":[],"minPlayers":2,"maxPlayers":8}`)
	if err != nil || message.StartsInMs != nil {
		t.Fatalf("starts in %v (%v), want none", message.StartsInMs, err)
	}
}

func TestARoundStartGivesTheTeamsTheSeatsAndTheMothership(t *testing.T) {
	text := `{"type":"roundStart","id":"a","settings":{"tickRate":20},"rules":{"mothershipLife":200,"roundMs":180000,"gunnerTurnRate":1.5},
	"mothership":{"ship":{"id":"mothership","name":"Mothership","layers":[]},"scale":2,"life":200,"maxLife":200},
	"players":[{"id":"a","name":"Ann","team":"attackers","hits":3},{"id":"b","name":"Ben","team":"gunners","seat":{"p":[1,2,3],"up":-1}}],"endsInMs":180000}`
	message, _, err := (&Reader{}).Read(text)
	if err != nil {
		t.Fatal(err)
	}
	if message.Rules.MothershipLife != 200 || message.Rules.GunnerTurnRate != 1.5 || message.Mothership.MaxLife != 200 || message.Mothership.Scale != 2 || message.EndsInMs != 180000 {
		t.Fatalf("round %+v", message)
	}
	attacker, gunner := message.Players[0], message.Players[1]
	if attacker.Team != TeamAttackers || attacker.Hits != 3 || attacker.Seat != nil {
		t.Fatalf("attacker %+v", attacker)
	}
	if gunner.Team != TeamGunners || gunner.Seat == nil || gunner.Seat.Position != (flight.Vector{1, 2, 3}) || gunner.Seat.Up != -1 {
		t.Fatalf("gunner %+v", gunner)
	}
}

func TestARoundOverSaysWhoWon(t *testing.T) {
	message, _, err := (&Reader{}).Read(`{"type":"roundOver","winner":"gunners","returnInMs":8000}`)
	if err != nil || message.Winner != TeamGunners || message.ReturnInMs != 8000 {
		t.Fatalf("round over %+v (%v)", message, err)
	}
}
