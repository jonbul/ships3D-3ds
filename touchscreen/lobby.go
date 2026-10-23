package touchscreen

import (
	"math"
	"strconv"

	"github.com/jonbul/ships3D-3ds/game"
)

// LobbyPage is the assault's waiting room, as the browser shows it: how many wait out of how
// many, when the round starts (or how many more it needs), and who waits, ownID marked as you.
func LobbyPage(lobby game.Lobby, ownID string, now float64, back func()) Page {
	rows := []Row{{Label: strconv.Itoa(len(lobby.Players)) + " / " + strconv.Itoa(lobby.MaxPlayers) + " players"}}
	rows = append(rows, StatusRows(lobbyStatus(lobby, now))...)
	rows = append(rows, Row{})
	for _, waiting := range lobby.Players {
		row := Row{Label: Fitted(waiting.Name, LabelCharacters-4)}
		if waiting.Id == ownID {
			row.Value = "YOU"
		}
		rows = append(rows, row)
	}
	rows = append(rows, Row{}, Row{Label: "Back to the ships", Activate: back})
	return Page{Title: LobbyTitle, Small: true, Rows: rows}
}

// lobbyStatus says when the round starts, or that it waits for more players.
func lobbyStatus(lobby game.Lobby, now float64) string {
	if lobby.CountingDown {
		return "Starting in " + strconv.Itoa(int(math.Ceil(lobby.StartsAt-now))) + " s"
	}
	return "Waiting for at least " + strconv.Itoa(lobby.MinPlayers) + " players…"
}
