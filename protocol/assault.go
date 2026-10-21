package protocol

import "github.com/jonbul/ships3D-3ds/flight"

// The assault (the server's /ws/assault) plays its rounds with the free for all's messages, and
// these around them: the lobby while players gather, the round's start with the teams, and who
// won. A hit on the mothership, and the damage it causes, has MothershipTarget as its target.
const (
	TypeLobby      = "lobby"
	TypeRoundStart = "roundStart"
	TypeRoundOver  = "roundOver"

	TeamAttackers = "attackers"
	TeamGunners   = "gunners"

	MothershipTarget = "mothership"
)

// Rules are the assault's rules, sent with each round. In a round the settings' MaxLife and
// FireCooldownMs are the attackers'.
type Rules struct {
	MinPlayers         int     `json:"minPlayers"`
	MaxPlayers         int     `json:"maxPlayers"`
	LobbyWaitMs        int     `json:"lobbyWaitMs"`
	AttackersPerGunner int     `json:"attackersPerGunner"`
	AttackerLife       float64 `json:"attackerLife"`
	MothershipLife     float64 `json:"mothershipLife"`
	MothershipRadius   float64 `json:"mothershipRadius"`
	AttackerCooldownMs int     `json:"attackerCooldownMs"`
	GunnerCooldownMs   int     `json:"gunnerCooldownMs"`
	GunnerTurnRate     float64 `json:"gunnerTurnRate"` // radians a second
	SpawnDistance      float64 `json:"spawnDistance"`
	RoundMs            int     `json:"roundMs"`
	ResultsMs          int     `json:"resultsMs"`
}

// Seat is a gunner's turret on the mothership, in the design's units: Up is 1 on its top and -1
// under it.
type Seat struct {
	Position flight.Vector `json:"p"`
	Up       float64       `json:"up"`
}

// Mothership is the ship the attackers attack and the gunners defend: its design, the scale
// that sizes it, and its life.
type Mothership struct {
	Ship    Ship    `json:"ship"`
	Scale   float64 `json:"scale"`
	Life    float64 `json:"life"`
	MaxLife float64 `json:"maxLife"`
}
