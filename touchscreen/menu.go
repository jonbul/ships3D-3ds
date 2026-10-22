package touchscreen

import (
	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/pixelart"
)

// Row is one line of a menu page: a choice when it can be activated or adjusted, otherwise a
// line to read.
type Row struct {
	Label    string
	Value    string
	Level    int  // drawn as a bar of LevelSteps blocks when HasLevel
	HasLevel bool // the value is a level, 0 to LevelSteps
	IsSwitch bool // the value is a switch, on or off
	On       bool
	Colour   draw2d.Color // the label's colour; none is the usual ink
	Activate func()       // A, or a touch on the label
	Adjust   func(step int)
}

// Chooseable tells whether the cursor stops on the row.
func (row Row) Chooseable() bool { return row.Activate != nil || row.Adjust != nil }

// Page is a window of rows with a title; L and R go through the pages of a menu.
type Page struct {
	Title     string
	Rows      []Row
	ValueLeft float32 // where values start, in pixels; 0 is the usual place
	Small     bool    // a window in the middle of the screen rather than over all of it
}

// LevelSteps is how many blocks a level's bar has; a volume of 0.7 shows seven.
const LevelSteps = 10

// The menu's measures, in pixels.
const (
	smallWindowWidth  = 280
	smallWindowMargin = 12
	fullPageMargin    = 3 // the stars show around a page
)

// LabelCharacters is how much of a line fits across a window, in the pixel font.
const LabelCharacters = 36

// Buttons are the buttons a menu reads, as bits: a frame's pressed ones are or-ed together.
type Buttons uint16

// The buttons a menu reads; Touch is a touch of the screen.
const (
	ButtonUp Buttons = 1 << iota
	ButtonDown
	ButtonLeft
	ButtonRight
	ButtonA
	ButtonB
	ButtonL
	ButtonR
	ButtonTouch
)

// Input is what a menu reads of a frame: the buttons pressed and where the screen was touched,
// in its pixels.
type Input struct {
	Pressed            Buttons
	TouchX, TouchY     float32
	Closes, TurnsPages bool // B closes the menu; L and R turn its pages
}

// Sound is the sound a menu asks for after a press.
type Sound int

// The sounds of a menu: none, the cursor moving, a choice, going back.
const (
	Silent Sound = iota
	Moved
	Chose
	Went
)

// Menu is where the cursor is in the pages shown; it lasts while the same pages are shown.
type Menu struct {
	Page, cursor, scroll int
	visible              int     // how many rows the page's window shows, in the screen it is laid out in
	opened               string  // the first page's title when the menu was set up: other pages, other menu
	openedAt             float64 // when, in seconds: the window unfolds from then
}

// unfoldSeconds is how long a window takes to unfold when it opens.
const unfoldSeconds = 0.18

// Reset makes the menu open new on the next frame, on its first page.
func (shown *Menu) Reset() { shown.opened = "" }

// prepare puts the cursor on the page's first choice when the pages are new, or a page changed,
// and measures how many rows the page shows in screen.
func (shown *Menu) prepare(pages []Page, screen draw2d.Rect, now float64) {
	if len(pages) == 0 {
		return
	}
	if pages[0].Title != shown.opened {
		shown.opened, shown.Page, shown.openedAt = pages[0].Title, 0, now
		shown.turnTo(pages[0])
	}
	shown.Page = min(shown.Page, len(pages)-1)
	page := pages[shown.Page]
	shown.visible = visibleRows(pageArea(page, screen))
	if shown.cursor < 0 || shown.cursor >= len(page.Rows) || !page.Rows[shown.cursor].Chooseable() {
		shown.cursor = firstChoice(page)
	}
}

// turnTo puts the cursor at the top of a page just turned to.
func (shown *Menu) turnTo(page Page) {
	shown.cursor, shown.scroll = firstChoice(page), 0
}

// firstChoice is the index of the page's first chooseable row, or -1 when it has none.
func firstChoice(page Page) int {
	for index, row := range page.Rows {
		if row.Chooseable() {
			return index
		}
	}
	return -1
}

// move goes to the next choice up or down (step -1 or 1), keeping it in sight; on a page with
// nothing to choose it scrolls instead. It tells whether anything moved.
func (shown *Menu) move(page Page, step int) bool {
	if shown.cursor < 0 {
		scroll := min(max(shown.scroll+step, 0), max(len(page.Rows)-shown.visible, 0))
		moved := scroll != shown.scroll
		shown.scroll = scroll
		return moved
	}
	for index := shown.cursor + step; index >= 0 && index < len(page.Rows); index += step {
		if page.Rows[index].Chooseable() {
			shown.cursor = index
			shown.keepInSight(page)
			return true
		}
	}
	return false
}

// keepInSight scrolls so that the cursor's row shows, and the lines just above it too.
func (shown *Menu) keepInSight(page Page) {
	top := shown.cursor
	for top > 0 && !page.Rows[top-1].Chooseable() && shown.cursor-top < shown.visible-1 {
		top-- // a heading above a choice comes into sight with it
	}
	if top < shown.scroll {
		shown.scroll = top
	}
	if shown.cursor >= shown.scroll+shown.visible {
		shown.scroll = shown.cursor - shown.visible + 1
	}
}

// Press acts on a frame's buttons and touch over pages laid out in screen (a touch is in its
// pixels): the cross moves, A activates, left and right adjust, L and R turn pages, B closes. It
// gives back the sound to play and whether B asked to close.
func (shown *Menu) Press(pages []Page, input Input, screen draw2d.Rect, now float64) (sound Sound, closed bool) {
	shown.prepare(pages, screen, now)
	if len(pages) == 0 {
		return Silent, false
	}
	page := pages[shown.Page]
	pressed := input.Pressed
	switch {
	case pressed&ButtonB != 0 && input.Closes:
		return Went, true
	case pressed&(ButtonL|ButtonR) != 0 && input.TurnsPages && len(pages) > 1:
		step := 1
		if pressed&ButtonL != 0 {
			step = -1
		}
		return shown.turnPage(pages, step), false
	case pressed&ButtonUp != 0:
		return movedSound(shown.move(page, -1)), false
	case pressed&ButtonDown != 0:
		return movedSound(shown.move(page, 1)), false
	case pressed&ButtonTouch != 0:
		return shown.touch(pages, input, screen), false
	case shown.cursor < 0:
		return Silent, false
	case pressed&ButtonA != 0:
		return activate(page.Rows[shown.cursor]), false
	case pressed&ButtonLeft != 0:
		return adjust(page.Rows[shown.cursor], -1), false
	case pressed&ButtonRight != 0:
		return adjust(page.Rows[shown.cursor], 1), false
	}
	return Silent, false
}

// turnPage goes step pages on, round the ends.
func (shown *Menu) turnPage(pages []Page, step int) Sound {
	shown.Page = (shown.Page + step + len(pages)) % len(pages)
	shown.turnTo(pages[shown.Page])
	return Moved
}

func movedSound(moved bool) Sound {
	if moved {
		return Moved
	}
	return Silent
}

func activate(row Row) Sound {
	if row.Activate == nil {
		return adjust(row, 1) // a toggle: A flips it as the cross does
	}
	row.Activate()
	return Chose
}

func adjust(row Row, step int) Sound {
	if row.Adjust == nil {
		return Silent
	}
	row.Adjust(step)
	return Moved
}

// touch acts on a touch: the ribbon's tabs turn the page, a row is chosen and activated, a touch
// on a value's left half lowers it and on its right half raises it.
func (shown *Menu) touch(pages []Page, input Input, screen draw2d.Rect) Sound {
	page := pages[shown.Page]
	area := pageArea(page, screen)
	if input.TouchY < area.Y+ribbonHeight {
		for index, place := range tabPlaces(area, pages, shown.Page) {
			if input.TurnsPages && index != shown.Page && input.TouchX >= place.X && input.TouchX < place.X+place.Width {
				return shown.turnPage(pages, index-shown.Page)
			}
		}
		return Silent
	}
	index := shown.scroll + int((input.TouchY-area.Y-rowsTop)/rowHeight)
	if input.TouchY < area.Y+rowsTop || index >= len(page.Rows) || index >= shown.scroll+shown.visible || !page.Rows[index].Chooseable() {
		return Silent
	}
	shown.cursor = index
	row := page.Rows[index]
	valueStart := area.X + valueLeftOf(page)
	if row.Adjust != nil && input.TouchX >= valueStart {
		step := 1
		if input.TouchX < valueStart+(area.X+area.Width-valueStart)/2 {
			step = -1
		}
		return adjust(row, step)
	}
	return activate(row)
}

// pageArea is where a page's window is in screen: all of the touch screen, or in a taller screen as
// tall as its rows need; a small page is a small window in the middle. None is taller than the
// screen.
func pageArea(page Page, screen draw2d.Rect) draw2d.Rect {
	whole := pixelart.Inset(screen, fullPageMargin)
	if !page.Small && screen.Height <= Height {
		return whole // a page fills the touch screen
	}
	height := min(float32(rowsTop+len(page.Rows)*rowHeight+hintsHeight+smallWindowMargin), whole.Height)
	if !page.Small {
		return draw2d.Rect{X: whole.X, Y: whole.Y + (whole.Height-height)/2, Width: whole.Width, Height: height}
	}
	width := min(smallWindowWidth, whole.Width)
	return draw2d.Rect{X: screen.X + (screen.Width-width)/2, Y: screen.Y + (screen.Height-height)/2, Width: width, Height: height}
}

// visibleRows is how many rows fit in a page's window.
func visibleRows(area draw2d.Rect) int {
	return max(int((area.Height-rowsTop-hintsHeight)/rowHeight), 1)
}

func valueLeftOf(page Page) float32 {
	if page.ValueLeft > 0 {
		return page.ValueLeft
	}
	return valueLeft
}

// Paint draws the page the menu is on, laid out in screen: its rows, cursor, tabs and hints.
// What is behind it (the 3DS's night, or the game) is the caller's.
func (shown *Menu) Paint(drawing *draw2d.List, screen draw2d.Rect, pages []Page, hint string, now float64) {
	shown.prepare(pages, screen, now)
	if len(pages) == 0 {
		return
	}
	page := pages[shown.Page]
	area := pageArea(page, screen)
	if unfolded := (now - shown.openedAt) / unfoldSeconds; unfolded < 1 {
		// The window unfolds from its middle line, as the old handhelds opened theirs.
		height := max(float32(unfolded)*area.Height, 8)
		window(drawing, draw2d.Rect{X: area.X, Y: area.Y + (area.Height-height)/2, Width: area.Width, Height: height})
		return
	}
	window(drawing, area)
	for index, place := range tabPlaces(area, pages, shown.Page) {
		tab(drawing, place, pages[index].Title, index == shown.Page)
	}
	for line := range min(shown.visible, len(page.Rows)-shown.scroll) {
		index := shown.scroll + line
		paintRow(drawing, area, page, page.Rows[index], area.Y+rowsTop+float32(line*rowHeight), index == shown.cursor, now)
	}
	if shown.scroll > 0 {
		paintScrollMark(drawing, area, area.Y+rowsTop-6, -1)
	}
	if shown.scroll+shown.visible < len(page.Rows) {
		paintScrollMark(drawing, area, area.Y+area.Height-hintsHeight-4, 1)
	}
	hints(drawing, area, hint)
}

// paintRow draws one row; the chosen one on a band, with the cursor before it.
func paintRow(drawing *draw2d.List, area draw2d.Rect, page Page, row Row, top float32, chosen bool, now float64) {
	if chosen {
		band := draw2d.Rect{X: area.X + frameWidth + 6, Y: top, Width: area.Width - 2*frameWidth - 12, Height: rowHeight - 2}
		pixelart.Stepped(drawing, band, 1, Colours.Chosen)
		shine(drawing, band, now)
		cursor(drawing, area.X+12, top+(rowHeight-2-cursorHeight)/2, now)
	}
	ink := Colours.Soft
	if row.Chooseable() {
		ink = Colours.Ink
	}
	if row.Colour != (draw2d.Color{}) {
		ink = row.Colour
	}
	pixelText(drawing, row.Label, area.X+textLeft, top, ink)
	valueX := area.X + valueLeftOf(page)
	switch {
	case row.HasLevel:
		energyBar(drawing, valueX, top, row.Level, LevelSteps)
		return
	case row.IsSwitch:
		pixelSwitch(drawing, valueX, top, row.On)
		return
	}
	valueInk := Colours.Soft
	if chosen && row.Adjust != nil {
		valueInk = Colours.Gold
	}
	pixelText(drawing, row.Value, valueX, top, valueInk)
}

// tabPlaces are where the pages' tabs go along the window's top, one after another, the open
// one wider.
func tabPlaces(area draw2d.Rect, pages []Page, open int) []draw2d.Rect {
	places := make([]draw2d.Rect, len(pages))
	x := area.X + 8
	for index, page := range pages {
		width := tabWidth(page.Title, index == open)
		places[index] = draw2d.Rect{X: x, Y: area.Y + tabsTop, Width: width, Height: tabHeight}
		x += width + tabGap
	}
	return places
}

// paintScrollMark draws a small triangle pointing up (-1) or down (1): more rows that way.
func paintScrollMark(drawing *draw2d.List, area draw2d.Rect, y float32, direction int) {
	const halfWidth = 4
	x := area.X + area.Width - 18
	for row := range halfWidth {
		width := float32(2 * (row + 1))
		rowY := y + float32(row)
		if direction > 0 {
			rowY = y + float32(halfWidth-1-row)
		}
		drawing.Rectangle(draw2d.Rect{X: x - width/2, Y: rowY, Width: width, Height: 1}, 0, Colours.Gold, draw2d.Color{}, 0)
	}
}

// Fitted shortens a line to what fits in characters, with an ellipsis.
func Fitted(text string, characters int) string {
	runes := []rune(text)
	if len(runes) <= characters {
		return text
	}
	return string(runes[:characters-1]) + "…"
}

// Wrapped breaks text into lines of at most characters, between words where it can.
func Wrapped(text string, characters int) []string {
	lines := []string{}
	runes := []rune(text)
	for len(runes) > characters {
		cut := characters
		for space := characters; space > characters/2; space-- {
			if runes[space] == ' ' {
				cut = space
				break
			}
		}
		lines = append(lines, string(runes[:cut]))
		runes = runes[cut:]
		for len(runes) > 0 && runes[0] == ' ' {
			runes = runes[1:]
		}
	}
	return append(lines, string(runes))
}
