//go:build !nintendo3ds

package main

// platformCaption is the title's caption, under the game's name.
const platformCaption = "a space dogfight"

// paintsSky is true on a computer: its renderer paints a day's sky wherever nothing covers the
// far plane, with the sun's glow, so the game darkens it (darkenSky) and the backdrop covers it.
const paintsSky = true
