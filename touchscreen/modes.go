package touchscreen

import (
	"strings"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/pixelart"
)

// The modes' icons, drawn by hand in pixels as the browser's are: each row a line of the
// picture, '#' its colour, '+' a lighter one, '.' nothing.
var (
	// FreeForAllIcon is two ships passing each other, shooting.
	FreeForAllIcon = []string{
		"................",
		".##.............",
		".###............",
		"..#++#..........",
		"..######..+.+.+.",
		"..#++#..........",
		".###............",
		".##.............",
		".............##.",
		"............###.",
		"..........#++#..",
		".+.+.+..######..",
		"..........#++#..",
		"............###.",
		".............##.",
		"................",
	}
	// AssaultIcon is the mothership, with its turrets.
	AssaultIcon = []string{
		"................",
		"......#..#......",
		".....######.....",
		"....##++++##....",
		"..############..",
		".##+##+##+##+##.",
		"################",
		"#++############+",
		"################",
		".##############.",
		"..##.##..##.##..",
		"...#..#..#..#...",
		"................",
		"..+..........+..",
		"................",
		"................",
	}
)

// Mode is one of the game's modes, as its card shows it.
type Mode struct {
	Name, Text string
	Icon       []string // drawn by hand: '#' its colour, '+' a lighter one
	Open       bool     // it can be played; otherwise its card is dimmed
}

// The mode window's measures, in pixels.
const (
	modeWindowHeight = 150
	modeCardGap      = 8
	modeCardInset    = 10
	modeCardsTop     = 30
	modeIconPixel    = 3
	modeIconTop      = 10
	modeNameTop      = 72
	modeTextTop      = 92
	modeTextLine     = 12              // between a card's lines of text
	modeTextMargin   = 4               // on either side of a card's text
	modeHintsBelow   = 2 * hintsHeight // the hints, under the window
)

// ModeTitle is the mode window's title.
const ModeTitle = "GAME MODE"

// ModeHint is the mode window's buttons' hints.
const ModeHint = "< > CHOOSE  A OK  B BACK"

// modeArea is where the mode window is: across the screen, in its middle.
func modeArea() draw2d.Rect {
	area := pixelart.Inset(FullScreen, fullPageMargin)
	return draw2d.Rect{X: area.X, Y: (Height - modeWindowHeight) / 2, Width: area.Width, Height: modeWindowHeight}
}

// modeCard is where the card of the mode at index is, of count.
func modeCard(index, count int) draw2d.Rect {
	area := modeArea()
	width := (area.Width - 2*modeCardInset - float32(count-1)*modeCardGap) / float32(count)
	return draw2d.Rect{X: area.X + modeCardInset + float32(index)*(width+modeCardGap), Y: area.Y + modeCardsTop, Width: width, Height: modeWindowHeight - modeCardsTop - modeCardInset}
}

// ModeChoice draws the modes as cards in a window over the night, with the star on the chosen
// one: its icon in the accent, the others dimmed.
func ModeChoice(drawing *draw2d.List, measurer draw2d.Measurer, modes []Mode, chosen int, now float64) {
	Starfield(drawing, now)
	area := modeArea()
	window(drawing, area)
	ribbon := draw2d.Rect{X: area.X + 8, Y: area.Y + tabsTop, Width: tabWidth(ModeTitle, true), Height: tabHeight}
	tab(drawing, ribbon, ModeTitle, true)
	for index, mode := range modes {
		paintModeCard(drawing, measurer, modeCard(index, len(modes)), mode, index == chosen, now)
	}
	hints(drawing, draw2d.Rect{X: area.X, Y: area.Y + area.Height + modeHintsBelow, Width: area.Width}, ModeHint)
}

func paintModeCard(drawing *draw2d.List, measurer draw2d.Measurer, card draw2d.Rect, mode Mode, chosen bool, now float64) {
	ink, iconColour := Colours.Soft, Colours.Tab
	if mode.Open {
		iconColour = Colours.Soft
	}
	if chosen {
		pixelart.Stepped(drawing, card, 1, Colours.Chosen)
		shine(drawing, card, now)
		cursor(drawing, card.X+4, card.Y+4, now)
		ink, iconColour = Colours.Ink, Colours.Gold
	}
	iconWidth := float32(len(mode.Icon[0])) * modeIconPixel
	scaledBitmap(drawing, mode.Icon, card.X+(card.Width-iconWidth)/2, card.Y+modeIconTop, modeIconPixel, iconColour, Colours.Ink)
	name := card.X + (card.Width-measurer.MeasureText(mode.Name, labelPixels, draw2d.Mono, 1).Width)/2
	drawing.Text(mode.Name, name, card.Y+modeNameTop, labelPixels, draw2d.Mono, ink)
	lines := cardLines(measurer, mode.Text, card.Width-2*modeTextMargin)
	top := card.Y + modeTextTop - float32(len(lines)-1)*modeTextLine/2
	for index, line := range lines {
		x := card.X + (card.Width-measurer.MeasureText(line, pixelTextSize, draw2d.Regular, 1).Width)/2
		drawing.Text(line, x, top+float32(index)*modeTextLine, pixelTextSize, draw2d.Regular, Colours.Soft)
	}
}

// cardLines breaks a card's text into lines no wider than width, between words.
func cardLines(measurer draw2d.Measurer, text string, width float32) []string {
	var lines []string
	line := ""
	for _, word := range strings.Fields(text) {
		longer := word
		if line != "" {
			longer = line + " " + word
		}
		if line != "" && measurer.MeasureText(longer, pixelTextSize, draw2d.Regular, 1).Width > width {
			lines = append(lines, line)
			longer = word
		}
		line = longer
	}
	return append(lines, line)
}

// ModeTouched is the index of the mode whose card a touch at x, y is on, of count, or -1.
func ModeTouched(x, y float32, count int) int {
	for index := range count {
		card := modeCard(index, count)
		if x >= card.X && x < card.X+card.Width && y >= card.Y && y < card.Y+card.Height {
			return index
		}
	}
	return -1
}
