package main

import (
	"strings"

	"github.com/akrck02/go-engine/engine"
)

// The keys' hints are made from what engine.json binds to each action, as the player last played
// (a keyboard, a pad, the 3DS's buttons), so they stay right when a player rebinds or picks up a
// pad. A line is made again only when the kind of input changes, so painting one allocates
// nothing.

// hintKeys is the input read this frame (touchInput keeps it), which the hints are made from.
var hintKeys *engine.Input

// hintPart is one hint of a line: the actions' keys, then what they do.
type hintPart struct {
	actions []string // more than one are written together, as "LEFT/RIGHT"
	does    string
}

// hintLine is a line of hints, and the kind of input it was last made for.
type hintLine struct {
	parts []hintPart
	kind  engine.InputKind
	text  string
}

// The game's lines of hints.
var (
	menuHints    = &hintLine{parts: []hintPart{{[]string{"confirm"}, "OK"}, {[]string{"back"}, "BACK"}, {[]string{"previous page", "next page"}, "PAGE"}}}
	modeHints    = &hintLine{parts: []hintPart{{[]string{"left", "right"}, "CHOOSE"}, {[]string{"confirm"}, "OK"}, {[]string{"back"}, "BACK"}}}
	joiningHints = &hintLine{parts: []hintPart{{[]string{"back"}, "SHIPS"}}}
	matchHints   = &hintLine{parts: []hintPart{{[]string{"fire"}, "FIRE"}, {[]string{"up", "down"}, "ZOOM"}, {[]string{"menu"}, "MENU"}}}
	shipHints    = &hintLine{parts: []hintPart{{[]string{"left", "right"}, "SHIP"}, {[]string{"back"}, "BACK"}, {[]string{"menu"}, "MENU"}}}
	launchHint   = &hintLine{parts: []hintPart{{[]string{"confirm"}, "LAUNCH"}}}
	startHint    = &hintLine{parts: []hintPart{{[]string{"confirm"}, "START"}}}
)

// hintGap is the room between two hints of a line.
const hintGap = "   "

// shown is the line's text for the input played last, made again when that changed.
func (line *hintLine) shown() string {
	if hintKeys == nil {
		return ""
	}
	if kind := hintKeys.LastInputKind(); line.text == "" || kind != line.kind {
		line.kind, line.text = kind, line.made(hintKeys)
	}
	return line.text
}

// made writes the line: each part's keys, as the player reads them, and what they do.
func (line *hintLine) made(keys *engine.Input) string {
	var text strings.Builder
	for index, part := range line.parts {
		if index > 0 {
			text.WriteString(hintGap)
		}
		for each, action := range part.actions {
			if each > 0 {
				text.WriteString("/")
			}
			text.WriteString(strings.ToUpper(keys.ActionHint(action)))
		}
		text.WriteString(" " + part.does)
	}
	return text.String()
}
