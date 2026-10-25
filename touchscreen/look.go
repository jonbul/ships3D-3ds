// Package touchscreen is Ships 3D's touch screen, as the 3DS shows it: windows drawn pixel by
// pixel as handheld games drew theirs, floating over a blue night with stars; the menu's pages
// as tabs along their top and a little star for a cursor; the radar's readings; the title's
// prompt and the hangar's arrows. It draws into a draw list (draw2d), in the touch screen's
// pixels (Width × Height), so any program can show it: on the console's bottom screen, or in a
// window on a computer.
package touchscreen

import (
	"math"

	"github.com/akrck02/go-renderer/draw2d"

	"github.com/jonbul/ships3D-3ds/pixelart"
)

// Width and Height are the touch screen's, in pixels: the 3DS's bottom screen.
const (
	Width  = 320
	Height = 240
)

// The windows' measures, in pixels.
const (
	cornerSteps   = 3
	frameWidth    = 3
	shadowDrop    = 4
	tabsTop       = 7
	tabHeight     = 18
	tabPadding    = 6
	tabGap        = 3
	iconSize      = 12
	iconGap       = 4
	ribbonHeight  = tabsTop + tabHeight // where the tabs end
	rowHeight     = 20
	rowsTop       = ribbonHeight + 8
	textLeft      = 26
	valueLeft     = 196
	hintsHeight   = 18
	cursorHeight  = 7
	pixelTextSize = pixelart.TextSize
	labelPixels   = pixelart.LabelSize
)

// Colours is the windows' palette, "a blue moon": windows of night-blue glass with a pale blue
// edge, the stars and the nebula showing through, pale ink, and a sky blue (Gold, the accent)
// for what is chosen.
var Colours = struct {
	Screen, Window, Tab, Frame, Shadow, Ink, Soft, Gold, GoldLight, Chosen, Glow, Danger, Good, Star draw2d.Color
}{
	Screen: draw2d.MustHex("#0a1430"), Window: draw2d.MustHex("#0e1c4290"), Tab: draw2d.MustHex("#24406e"),
	Frame: draw2d.MustHex("#9cc8f2"), Shadow: draw2d.MustHex("#02060fb0"), Ink: draw2d.MustHex("#eaf4ff"),
	Soft: draw2d.MustHex("#9fb6d6"), Gold: draw2d.MustHex("#5fb0ff"), GoldLight: draw2d.MustHex("#cfe6ff"),
	Chosen: draw2d.MustHex("#2a4a80"), Glow: draw2d.MustHex("#9cc8f233"), Danger: draw2d.MustHex("#ff7088"),
	Good: draw2d.MustHex("#5fb0ff"), Star: draw2d.MustHex("#eaf4ff"),
}

// The nebula and the moon behind the menus: soft clouds of colour and a crescent, in pixels.
var nebula = []struct {
	x, y, size float32
	colour     draw2d.Color
}{
	{60, 40, 220, draw2d.MustHex("#5b3fa814")}, {250, 150, 260, draw2d.MustHex("#2f6fd014")},
	{140, 210, 160, draw2d.MustHex("#c0508a10")}, {300, 20, 120, draw2d.MustHex("#4fc3f710")},
}

const (
	moonX, moonY, moonSize = 262, 168, 96
	moonShadowShift        = 18
)

// moonCraters are where the moon's craters are, from its top left corner.
var moonCraters = []draw2d.Rectangle{{X: 20, Y: 22, Width: 18, Height: 12}, {X: 46, Y: 50, Width: 12, Height: 8}}

var (
	moonLight  = draw2d.MustHex("#dfe9f8")
	moonCrater = draw2d.MustHex("#c3d2ea")
)

// FullScreen is the touch screen's area.
var FullScreen = draw2d.Rectangle{Width: Width, Height: Height}

// window draws a window over area: a hard shadow, a white edge and the pale inside.
func window(drawing *draw2d.List, area draw2d.Rectangle) {
	shadow := area
	shadow.X, shadow.Y = shadow.X+shadowDrop, shadow.Y+shadowDrop
	pixelart.Stepped(drawing, shadow, cornerSteps, Colours.Shadow)
	pixelart.Stepped(drawing, area, cornerSteps, Colours.Frame)
	pixelart.Stepped(drawing, pixelart.Inset(area, frameWidth), cornerSteps, Colours.Window)
}

// tab draws one page's tab: its icon, and its name too when it is open, in the accent.
func tab(drawing *draw2d.List, area draw2d.Rectangle, name string, open bool) {
	fill, ink, light := Colours.Tab, Colours.Soft, Colours.GoldLight
	if open {
		fill, ink, light = Colours.Gold, Colours.Screen, Colours.Frame
	}
	pixelart.Stepped(drawing, area, 1, fill)
	x := area.X + tabPadding
	if icon, found := pageIcons[name]; found {
		bitmap(drawing, icon, x, area.Y+(area.Height-iconSize)/2, ink, light)
		x += iconSize + iconGap
	}
	if open || pageIcons[name] == nil {
		drawing.Text(name, x, area.Y+area.Height-5, labelPixels, draw2d.Mono, ink)
	}
}

// tabWidth is how wide a page's tab is: an icon, and the name when open or when it has no icon.
// The figures' face is 8 pixels a letter, so no measuring is needed.
func tabWidth(name string, open bool) float32 {
	width := float32(2 * tabPadding)
	_, hasIcon := pageIcons[name]
	if hasIcon {
		width += iconSize
	}
	if open || !hasIcon {
		width += float32(len(name) * labelPixels)
		if hasIcon {
			width += iconGap
		}
	}
	return width
}

// The cursor's bobbing: how far, in pixels, and how fast, in turns a second.
const (
	cursorBob      = 2
	cursorBobSpeed = 1.6
)

// cursor draws the little star beside the chosen row, bobbing to and fro.
func cursor(drawing *draw2d.List, x, y float32, now float64) {
	bob := float32(math.Round(cursorBob * math.Sin(now*cursorBobSpeed*2*math.Pi)))
	bitmap(drawing, cursorStar, x+bob, y, Colours.Gold, Colours.GoldLight)
}

// The light that sweeps along the chosen row: its width in pixels and its passes a second.
const (
	sweepWidth  = 36
	sweepSpeed  = 0.7
	sweepBlocks = 6
)

// shine draws a light sweeping along a band, in blocks fading to its sides.
func shine(drawing *draw2d.List, band draw2d.Rectangle, now float64) {
	travel := band.Width + sweepWidth
	start := band.X - sweepWidth + float32(math.Mod(now*sweepSpeed, 1))*travel
	block := float32(sweepWidth) / sweepBlocks
	for index := range sweepBlocks {
		left := max(start+float32(index)*block, band.X)
		right := min(start+float32(index+1)*block, band.X+band.Width)
		if right <= left {
			continue
		}
		strength := 1 - abs(float32(index)-(sweepBlocks-1)/2.0)/(sweepBlocks/2.0)
		colour := Colours.Glow
		colour[3] *= strength
		drawing.Rectangle(draw2d.Rectangle{X: left, Y: band.Y, Width: right - left, Height: band.Height}, 0, colour, draw2d.Color{}, 0)
	}
}

// hints draws the buttons' hints along the bottom of area, in the figures' face.
func hints(drawing *draw2d.List, area draw2d.Rectangle, text string) {
	drawing.Text(text, area.X+12, area.Y+area.Height-8, labelPixels, draw2d.Mono, Colours.Soft)
}

// pixelText draws a line of text with its top at top.
func pixelText(drawing *draw2d.List, text string, x, top float32, colour draw2d.Color) {
	drawing.Text(text, x, top+pixelTextSize-3, pixelTextSize, draw2d.Regular, colour)
}

// The starfield behind the menus: how many stars, and how fast the nearest drift, in pixels a
// second; farther layers drift slower and dimmer.
const (
	menuStarCount     = 60
	menuStarLayers    = 3
	menuStarSpeed     = 18
	menuStarTwinkling = 3 // radians a second of a star's twinkle
)

// Starfield fills the screen with the night: a nebula, a crescent moon and stars drifting left,
// the near ones faster.
func Starfield(drawing *draw2d.List, now float64) {
	drawing.Rectangle(FullScreen, 0, Colours.Screen, draw2d.Color{}, 0)
	for _, cloud := range nebula {
		for ring := float32(1); ring >= 0.2; ring -= 0.2 {
			size := cloud.size * ring
			drawing.Rectangle(draw2d.Rectangle{X: cloud.x - size/2, Y: cloud.y - size/2, Width: size, Height: size}, size/2, cloud.colour, draw2d.Color{}, 0)
		}
	}
	drawing.Rectangle(draw2d.Rectangle{X: moonX, Y: moonY, Width: moonSize, Height: moonSize}, moonSize/2, moonLight, draw2d.Color{}, 0)
	for _, crater := range moonCraters {
		drawing.Rectangle(draw2d.Rectangle{X: moonX + crater.X, Y: moonY + crater.Y, Width: crater.Width, Height: crater.Height}, crater.Height/2, moonCrater, draw2d.Color{}, 0)
	}
	shade := draw2d.Rectangle{X: moonX - moonShadowShift, Y: moonY - moonShadowShift/2, Width: moonSize, Height: moonSize}
	drawing.Rectangle(shade, moonSize/2, Colours.Screen, draw2d.Color{}, 0) // the dark side cuts a crescent
	for index := range menuStarCount {
		paintStar(drawing, index, now)
	}
}

// paintStar draws one of the starfield's stars where it has drifted to by now.
func paintStar(drawing *draw2d.List, index int, now float64) {
	layer := index % menuStarLayers
	depth := float32(menuStarLayers-layer) / menuStarLayers // 1 is the nearest
	seed := uint32(index)*2654435761 + 12345
	baseX := float32(seed%Width) - float32(now)*menuStarSpeed*depth
	x := float32(math.Mod(float64(baseX), Width))
	if x < 0 {
		x += Width
	}
	y := float32((seed / Width) % Height)
	colour := Colours.Star
	colour[3] = depth * float32(0.55+0.45*math.Sin(now*menuStarTwinkling+float64(index)))
	size := float32(1)
	if layer == 0 && index%4 == 0 {
		size = 2
	}
	drawing.Rectangle(draw2d.Rectangle{X: x, Y: y, Width: size, Height: size}, 0, colour, draw2d.Color{}, 0)
}

// pixelSwitch draws an on / off switch: a round track with the knob to its side, blue when on.
func pixelSwitch(drawing *draw2d.List, x, top float32, on bool) {
	const width, height, knob = 26, 12, 10
	y := top + (rowHeight-2-height)/2
	track, knobX, label := Colours.Tab, x+1, "OFF"
	if on {
		track, knobX, label = Colours.Good, x+width-knob-1, "ON"
	}
	drawing.Rectangle(draw2d.Rectangle{X: x, Y: y, Width: width, Height: height}, height/2, track, draw2d.Color{}, 0)
	drawing.Rectangle(draw2d.Rectangle{X: knobX, Y: y + 1, Width: knob, Height: height - 2}, knob/2, Colours.Frame, draw2d.Color{}, 0)
	drawing.Text(label, x+width+8, y+height-2, labelPixels, draw2d.Mono, Colours.Soft)
}

// energyBar draws a level as round cells, lit to the level.
func energyBar(drawing *draw2d.List, x, top float32, level, steps int) {
	const width, gap, height = 7, 2, 10
	y := top + (rowHeight-2-height)/2
	for cell := range steps {
		colour := Colours.Tab
		if cell < level {
			colour = Colours.Gold
		}
		drawing.Rectangle(draw2d.Rectangle{X: x + float32(cell*(width+gap)), Y: y, Width: width, Height: height}, width/2, colour, draw2d.Color{}, 0)
	}
}

func abs(value float32) float32 {
	if value < 0 {
		return -value
	}
	return value
}
