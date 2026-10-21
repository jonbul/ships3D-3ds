package touchscreen

import "github.com/akrck02/go-renderer/draw2d"

// pageIcons are the menu's pages' icons on their tabs, 12 × 12 pixels: '#' the tab's ink, '+'
// lighter.
var pageIcons = map[string][]string{
	PilotTitle: {
		"....####....",
		"..########..",
		".##++++++##.",
		".#+######+#.",
		"##+#....#+##",
		"##+#....#+##",
		"##+######+##",
		".##++++++##.",
		"..########..",
		"...#....#...",
		"..##....##..",
		"............",
	},
	FiguresTitle: {
		"............",
		"..........##",
		"..........##",
		"......##..##",
		"......##..##",
		"..##..##..##",
		"..##..##..##",
		"..##..##..##",
		"++++++++++++",
		"............",
		"............",
		"............",
	},
	SettingsTitle: {
		".....##.....",
		"..#.####.#..",
		".##########.",
		"..###++###..",
		".###+..+###.",
		"####+..+####",
		".###+..+###.",
		"..###++###..",
		".##########.",
		"..#.####.#..",
		".....##.....",
		"............",
	},
	AccountTitle: {
		"....####....",
		"...######...",
		"...##++##...",
		"...######...",
		"....####....",
		"............",
		"..########..",
		".##########.",
		".###++++###.",
		".##########.",
		"............",
		"............",
	},
	ControlsTitle: {
		"............",
		"............",
		"..########..",
		".##########.",
		"##+###++#+##",
		"#+++###+#+##",
		"##+####+++##",
		"############",
		"###......###",
		".##......##.",
		"............",
		"............",
	},
}

// cursorStar is the cursor: a little four-pointed star, '#' in the accent and '+' lighter.
var cursorStar = []string{
	"...#...",
	"...#...",
	"..#+#..",
	"##+++##",
	"..#+#..",
	"...#...",
	"...#...",
}

// The menu's pages' titles, by which their icons go by.
const (
	PilotTitle    = "PILOT"
	FiguresTitle  = "STATS"
	SettingsTitle = "SETUP"
	AccountTitle  = "ACCOUNT"
	ControlsTitle = "CONTROLS"
)

// The titles of the windows that are not the menu's pages.
const (
	ConnectingTitle = "CONNECTING"
	RestartTitle    = "RESTART?"
	JoiningTitle    = "JOINING"
	LobbyTitle      = "LOBBY"
)

// bitmap draws a picture drawn by hand in text with its top left at x, y: '#' in main, '+' in light.
func bitmap(drawing *draw2d.List, rows []string, x, y float32, main, light draw2d.Color) {
	scaledBitmap(drawing, rows, x, y, 1, main, light)
}

// scaledBitmap draws such a picture with each of its pixels size pixels wide.
func scaledBitmap(drawing *draw2d.List, rows []string, x, y, size float32, main, light draw2d.Color) {
	for row, line := range rows {
		for column := 0; column < len(line); column++ {
			colour := main
			switch line[column] {
			case '.':
				continue
			case '+':
				colour = light
			}
			// Runs of one colour in one rectangle: fewer to draw.
			end := column + 1
			for end < len(line) && line[end] == line[column] {
				end++
			}
			drawing.Rectangle(draw2d.Rect{X: x + float32(column)*size, Y: y + float32(row)*size, Width: float32(end-column) * size, Height: size}, 0, colour, draw2d.Color{}, 0)
			column = end - 1
		}
	}
}
