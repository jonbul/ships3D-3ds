package touchscreen

import (
	"testing"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/figures"
	"github.com/jonbul/ships3D-3ds/game"
	"github.com/jonbul/ships3D-3ds/profile"
)

func nothing() {}

// menuPages are the menu's pages, as the games show them.
func menuPages() []Page {
	state := game.New()
	return []Page{
		PilotPage(state, Pilot{Resume: nothing, ChooseShip: nothing, Quit: nothing}),
		FiguresPage(figures.Second{}, 20, Motion{}),
		SettingsPage(profile.Settings{}, func(func(*profile.Settings)) {}, nothing),
		AccountPage(Account{Server: "example.com:3000", ChangeServer: nothing, LogIn: nothing, LogOut: nothing}),
		ControlsPage(ControlLines),
	}
}

func TestEveryMenuPageHasATitleAndAnIcon(test *testing.T) {
	for _, page := range menuPages() {
		if page.Title == "" {
			test.Fatalf("a page has no title: %+v", page)
		}
		if _, found := pageIcons[page.Title]; !found {
			test.Errorf("the page %q has no icon", page.Title)
		}
	}
}

func TestTheCursorSkipsLinesThatCannotBeChosen(test *testing.T) {
	chosen := ""
	page := Page{Title: "TEST", Rows: []Row{
		{Label: "a heading"},
		{Label: "first", Activate: func() { chosen = "first" }},
		{Label: "a line to read"},
		{Label: "second", Activate: func() { chosen = "second" }},
	}}
	var shown Menu
	shown.Press([]Page{page}, Input{Pressed: ButtonDown}, 0)
	sound, _ := shown.Press([]Page{page}, Input{Pressed: ButtonA}, 0)
	if chosen != "second" || sound != Chose {
		test.Errorf("down then A chose %q with sound %d, want second with a choice's sound", chosen, sound)
	}
}

func TestBClosesOnlyAMenuThatCloses(test *testing.T) {
	pages := menuPages()
	var shown Menu
	if _, closed := shown.Press(pages, Input{Pressed: ButtonB}, 0); closed {
		test.Error("B closed a menu that doesn't close")
	}
	if sound, closed := shown.Press(pages, Input{Pressed: ButtonB, Closes: true}, 0); !closed || sound != Went {
		test.Errorf("B gave closed %v and sound %d, want closed and the going back's sound", closed, sound)
	}
}

func TestLAndRTurnThePagesRoundTheEnds(test *testing.T) {
	pages := menuPages()
	var shown Menu
	shown.Press(pages, Input{Pressed: ButtonL, TurnsPages: true}, 0)
	if shown.Page != len(pages)-1 {
		test.Errorf("L from the first page went to page %d, want the last, %d", shown.Page, len(pages)-1)
	}
}

func TestATouchOnALevelsLeftHalfLowersIt(test *testing.T) {
	steps := 0
	page := Page{Title: "TEST", Rows: []Row{{Label: "Music", HasLevel: true, Adjust: func(step int) { steps += step }}}}
	var shown Menu
	area := pageArea(page)
	top := area.Y + rowsTop + rowHeight/2
	shown.Press([]Page{page}, Input{Pressed: ButtonTouch, TouchX: area.X + valueLeft + 1, TouchY: top}, 0)
	shown.Press([]Page{page}, Input{Pressed: ButtonTouch, TouchX: area.X + area.Width - 2, TouchY: top}, 0)
	shown.Press([]Page{page}, Input{Pressed: ButtonTouch, TouchX: area.X + area.Width - 2, TouchY: top}, 0)
	if steps != 1 {
		test.Errorf("one touch on the left and two on the right added %d steps, want 1", steps)
	}
}

func TestASettingsLevelStopsAtItsEnds(test *testing.T) {
	settings := profile.Settings{Music: 1}
	page := SettingsPage(settings, func(apply func(*profile.Settings)) { apply(&settings) }, nothing)
	for _, row := range page.Rows {
		if row.Label == "Music" {
			row.Adjust(1)
		}
	}
	if settings.Music != 1 {
		test.Errorf("raising the music at its loudest made it %v, want 1", settings.Music)
	}
}

func TestLongLinesWrapBetweenWords(test *testing.T) {
	lines := Wrapped("the server could not be reached", 12)
	for _, line := range lines {
		if len([]rune(line)) > 12 {
			test.Errorf("the line %q is longer than 12 characters", line)
		}
	}
	if lines[0] != "the server" {
		test.Errorf("the first line is %q, want it cut at a space: %q", lines[0], "the server")
	}
}

func TestTheHangarsArrowsAreTheScreensThirds(test *testing.T) {
	cases := []struct {
		x    float32
		want int
	}{{10, -1}, {Width / 2, 0}, {Width - 10, 1}}
	for _, each := range cases {
		if got := ShipArrowTouched(each.x); got != each.want {
			test.Errorf("a touch at %v is on arrow %d, want %d", each.x, got, each.want)
		}
	}
}

func TestAMenuPaintsItsWindowOverTheNight(test *testing.T) {
	var shown Menu
	var drawing draw2d.List
	shown.Paint(&drawing, menuPages(), MenuHint, unfoldSeconds)
	if len(drawing.Commands) == 0 {
		test.Fatal("the menu drew nothing")
	}
}

func TestATouchOnAModesCardChoosesIt(test *testing.T) {
	card := modeCard(1, 2)
	if got := ModeTouched(card.X+card.Width/2, card.Y+card.Height/2, 2); got != 1 {
		test.Errorf("a touch in the second card's middle chose mode %d, want 1", got)
	}
	if got := ModeTouched(1, 1, 2); got != -1 {
		test.Errorf("a touch in the screen's corner chose mode %d, want none", got)
	}
}

// eightWide measures every character 8 pixels wide.
type eightWide struct{}

func (eightWide) MeasureText(text string, size float32, weight draw2d.Weight, scale float32) draw2d.TextMetrics {
	return draw2d.TextMetrics{Width: 8 * float32(len(text))}
}

func (eightWide) HasIcon(string) bool { return false }

func TestAModesTextIsWrappedToFitItsCard(t *testing.T) {
	lines := cardLines(eightWide{}, "Teams around a mothership", 100)
	if len(lines) != 2 || lines[0] != "Teams around" || lines[1] != "a mothership" {
		t.Fatalf("lines %q, want two that fit 100 pixels", lines)
	}
	if lines := cardLines(eightWide{}, "Short", 100); len(lines) != 1 {
		t.Fatalf("lines %q, want one", lines)
	}
}
