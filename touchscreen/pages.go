package touchscreen

import (
	"math"
	"sort"
	"strconv"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/figures"
	"github.com/jonbul/ships3D-3ds/game"
	"github.com/jonbul/ships3D-3ds/profile"
)

// The buttons' hints along the bottom of the windows.
const (
	MenuHint    = "A OK  B BACK  L/R PAGE"
	ChoiceHint  = "A OK  B BACK"
	WaitingHint = "PLEASE WAIT"
)

// StatusRows are lines to read (the connection's status, the server's complaint), wrapped to the
// window; empty texts are left out.
func StatusRows(texts ...string) []Row {
	rows := []Row{}
	for _, text := range texts {
		if text == "" {
			continue
		}
		for _, line := range Wrapped(text, LabelCharacters) {
			rows = append(rows, Row{Label: line, Colour: Colours.Soft})
		}
	}
	return rows
}

// RestartWarning asks before a change of server or account, which starts the game again.
func RestartWarning(confirm, cancel func()) Page {
	return Page{Title: RestartTitle, Small: true, Rows: []Row{
		{Label: "You will leave the game"},
		{Label: "and connect again."},
		{},
		{Label: "Restart", Colour: Colours.Danger, Activate: confirm},
		{Label: "Cancel", Activate: cancel},
	}}
}

// Pilot is what the game's own page offers: the ways out of the game.
type Pilot struct {
	Resume, ChooseShip, Quit func()
}

// PilotPage is the game's own page: life and score, the ways out, and the others' scores.
func PilotPage(state *game.Game, offered Pilot) Page {
	rows := []Row{
		{Label: "Life", Value: strconv.Itoa(int(state.Player.Life)) + " / " + strconv.Itoa(int(state.Settings.MaxLife))},
		{Label: "Kills / deaths", Value: strconv.Itoa(state.Player.Kills) + " / " + strconv.Itoa(state.Player.Deaths)},
		{Label: "Resume", Activate: offered.Resume},
		{Label: "Choose another ship", Activate: offered.ChooseShip},
	}
	if offered.Quit != nil {
		rows = append(rows, Row{Label: "Quit the game", Colour: Colours.Danger, Activate: offered.Quit})
	}
	if len(state.Remotes) > 0 {
		rows = append(rows, Row{Label: "OTHER PILOTS", Colour: Colours.Gold})
	}
	return Page{Title: PilotTitle, Rows: append(rows, remoteRows(state)...)}
}

// remoteRows are the other players' scores, best first.
func remoteRows(state *game.Game) []Row {
	type score struct {
		name          string
		kills, deaths int
	}
	scores := []score{}
	for _, remote := range state.Remotes {
		scores = append(scores, score{remote.Player.Name, remote.Player.Kills, remote.Player.Deaths})
	}
	sort.Slice(scores, func(first, second int) bool { return scores[first].kills > scores[second].kills })
	rows := make([]Row, 0, len(scores))
	for _, entry := range scores {
		rows = append(rows, Row{Label: Fitted(entry.name, 20), Value: strconv.Itoa(entry.kills) + " / " + strconv.Itoa(entry.deaths)})
	}
	return rows
}

// settingsValueLeft is where the settings' values start, in pixels.
const settingsValueLeft = 186

// SettingsPage offers how the ship flies, the radar shows and the game sounds. change applies a
// change to the settings (and remembers it); tilted runs when steering by tilting is switched,
// to centre it on how the console is held.
func SettingsPage(settings profile.Settings, change func(apply func(settings *profile.Settings)), tilted func()) Page {
	adjusted := func(apply func(settings *profile.Settings, step int)) func(int) {
		return func(step int) { change(func(settings *profile.Settings) { apply(settings, step) }) }
	}
	flipped := func(apply func(settings *profile.Settings)) func(int) {
		return func(int) { change(apply) }
	}
	return Page{Title: SettingsTitle, ValueLeft: settingsValueLeft, Rows: []Row{
		{Label: "Up raises the nose", IsSwitch: true, On: settings.InvertPitch, Adjust: flipped(func(settings *profile.Settings) { settings.InvertPitch = !settings.InvertPitch })},
		{Label: "Radar", Value: radarName(settings.FlatRadar), Adjust: flipped(func(settings *profile.Settings) { settings.FlatRadar = !settings.FlatRadar })},
		{Label: "Steer by tilting", IsSwitch: true, On: settings.Gyroscope, Adjust: flipped(func(settings *profile.Settings) {
			settings.Gyroscope = !settings.Gyroscope
			tilted()
		})},
		{Label: "Fine aim", HasLevel: true, Level: level(settings.Precision), Adjust: adjusted(func(settings *profile.Settings, step int) { stepped(&settings.Precision, step) })},
		{Label: "Effects", HasLevel: true, Level: level(settings.Effects), Adjust: adjusted(func(settings *profile.Settings, step int) { stepped(&settings.Effects, step) })},
		{Label: "Music", HasLevel: true, Level: level(settings.Music), Adjust: adjusted(func(settings *profile.Settings, step int) { stepped(&settings.Music, step) })},
		{},
		{Label: "Fine aim: the circle pad softer"},
		{Label: "near its centre. L centres the tilt."},
	}}
}

// stepped moves a level of 0 to 1 by step tenths, within its ends.
func stepped(value *float32, step int) {
	*value = float32(min(max(math.Round(float64(*value*LevelSteps))+float64(step), 0), LevelSteps)) / LevelSteps
}

// level is a value of 0 to 1 in its bar's blocks.
func level(value float32) int { return int(value*LevelSteps + 0.5) }

func radarName(flat bool) string {
	if flat {
		return "FLAT"
	}
	return "3D"
}

// Account is what the account's page shows and offers.
type Account struct {
	Server, Username    string // no username is a guest
	ChangeServer, LogIn func()
	LogOut              func()
	Status              []Row // lines under it: why a login failed, the connection's status
}

// AccountPage shows the server and the account, each with its change.
func AccountPage(account Account) Page {
	rows := []Row{
		{Label: "SERVER", Colour: Colours.Gold},
		{Label: Fitted(account.Server, LabelCharacters-2)},
		{Label: "Change server", Activate: account.ChangeServer},
		{Label: "ACCOUNT", Colour: Colours.Gold},
	}
	if account.Username == "" {
		rows = append(rows, Row{Label: "Guest"}, Row{Label: "Log in", Activate: account.LogIn})
	} else {
		rows = append(rows, Row{Label: Fitted(account.Username, LabelCharacters-2)}, Row{Label: "Log out", Activate: account.LogOut})
	}
	return Page{Title: AccountTitle, Rows: append(rows, account.Status...)}
}

// ControlLines are what each control does on the 3DS, for the Controls page.
var ControlLines = [][2]string{
	{"Circle pad", "Down raises the nose; turn"},
	{"R or A", "Fire; A flies again after dying"},
	{"C-stick", "Roll, faster and slower"},
	{"Left / Right", "Roll"},
	{"ZR / ZL", "Faster / slower (also X / B)"},
	{"Tilting", "Nose and turn, if on in Setup"},
	{"L", "Centre the tilt"},
	{"Left / Right", "Ship to fly, on the ship screen"},
	{"Up / Down", "Radar zoom"},
	{"3D slider", "Depth"},
	{"START", "Options"},
}

// controlNameWidth is the Controls page's first column, in pixels.
const controlNameWidth = 92

// ControlsPage lists the controls in two columns: each line's control and what it does.
func ControlsPage(lines [][2]string) Page {
	rows := make([]Row, 0, len(lines))
	for _, line := range lines {
		rows = append(rows, Row{Label: line[0], Value: line[1], Colour: Colours.Gold})
	}
	return Page{Title: ControlsTitle, Rows: rows, ValueLeft: controlNameWidth + textLeft}
}

// The frame budget the figures are judged against: the screen's 60 frames a second.
const (
	targetFramesPerSecond   = 60
	frameBudgetMilliseconds = 1000.0 / targetFramesPerSecond
	bytesPerKilobyte        = 1024
	bytesPerMegabyte        = 1024 * 1024
	figuresValueLeft        = 150
)

// Motion is how the console turns, for the figures: the gyroscope's rates, in degrees a second,
// and the tilt it has added up, in degrees.
type Motion struct {
	Roll, Pitch, Yaw        float64
	TiltPitch, TiltSteering float64
}

// FiguresPage shows how the game runs: the frame, the drawing, the network and the tilt. A
// figure within its budget shows gold, one over it red. tickRate is the server's snapshots a
// second.
func FiguresPage(shown figures.Second, tickRate float64, motion Motion) Page {
	judged := func(good bool) draw2d.Color {
		if good {
			return Colours.Gold
		}
		return Colours.Danger
	}
	snapshotBudget := 2 * 1000 / float32(max(tickRate, 1))
	return Page{Title: FiguresTitle, ValueLeft: figuresValueLeft, Rows: []Row{
		{Label: "Frames", Value: strconv.Itoa(shown.FramesPerSecond) + " a second", Colour: judged(shown.FramesPerSecond >= targetFramesPerSecond-2)},
		{Label: "Processor", Value: tenths(shown.ProcessorMilliseconds) + " ms", Colour: judged(shown.ProcessorMilliseconds < frameBudgetMilliseconds)},
		{Label: "Graphics", Value: tenths(shown.GPUMilliseconds) + " ms", Colour: judged(shown.GPUMilliseconds < frameBudgetMilliseconds)},
		{Label: "Slowest frame", Value: tenths(shown.SlowestFrameMilliseconds) + " ms", Colour: judged(shown.SlowestFrameMilliseconds < 2*frameBudgetMilliseconds)},
		{Label: "Network", Value: tenths(shown.NetworkMilliseconds) + " ms"},
		{Label: "Drawing", Value: tenths(shown.DrawingMilliseconds) + " ms"},
		{Label: "Heap", Value: tenths(float32(shown.HeapBytes)/bytesPerMegabyte) + " MB"},
		{Label: "Triangles", Value: strconv.Itoa(shown.Triangles)},
		{Label: "Draws", Value: strconv.Itoa(shown.DrawCalls)},
		{Label: "Bakes", Value: strconv.Itoa(shown.BakesPerSecond) + " a second"},
		{Label: "Received", Value: strconv.Itoa(shown.MessagesIn) + " msg, " + tenths(float32(shown.BytesIn)/bytesPerKilobyte) + " KB/s"},
		{Label: "Sent", Value: strconv.Itoa(shown.MessagesOut) + " msg, " + tenths(float32(shown.BytesOut)/bytesPerKilobyte) + " KB/s"},
		{Label: "Since snapshot", Value: strconv.Itoa(int(shown.SnapshotDelayMilliseconds)) + " ms", Colour: judged(shown.SnapshotDelayMilliseconds < snapshotBudget)},
		{Label: "Players", Value: strconv.Itoa(shown.Players)},
		{Label: "Gyroscope", Value: signed(motion.Roll) + " " + signed(motion.Pitch) + " " + signed(motion.Yaw) + " deg/s"},
		{Label: "Tilt / steer", Value: signed(motion.TiltPitch) + " / " + signed(motion.TiltSteering)},
	}}
}

// tenths writes a number with one decimal.
func tenths(value float32) string {
	whole := int(value*10 + 0.5)
	return strconv.Itoa(whole/10) + "." + strconv.Itoa(whole%10)
}

// signed writes a whole number with its sign.
func signed(value float64) string {
	rounded := int(math.Round(value))
	if rounded > 0 {
		return "+" + strconv.Itoa(rounded)
	}
	return strconv.Itoa(rounded)
}
