# 4. A layout for one screen

> Status: built, pending review · Author's approval: proposal ☑ plan ☑ (an experiment: built in one go)

## Why

The desktop game shows the 3DS's touch screen pasted over the game: a 320 × 240 panel in the
middle for the modes and the menu, and the touch screen's whole picture, shrunk, as the radar in
a corner. That makes sense on a console with a second screen. On one big screen it wastes the
room and hides the game behind boxes. The pieces are right; the way they are laid out is not.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | The touch screen's pieces (the menu's pages, the mode cards, the radar's readings) lay out in any area they are given, not only in 320 × 240 | Tests: a page in a larger area shows more rows; a mode card is found where it was drawn |
| R2 | On one screen (the desktop, and later the web) the modes and the menu are laid out for the window: wider, with more rows, over the game, without the 3DS's starry ground | Screenshots on the desktop at 1280 × 720 and in a narrow window |
| R3 | On one screen the radar floats in a corner, as the browser's does: the sphere, its range under it (`RADAR · 600 u`), the kills with the HUD, and the keys in a line of hints | Screenshot in a match |
| R4 | The 3DS keeps its two screens as they are | Azahar screenshots of the modes, the menu and the radar |
| R5 | The keys are the engine's actions, bound per target in `engine.json`, and the layout is chosen by `engine.RunningOn()`, not by build files | Test: every action the game reads is bound on the desktop and the 3DS |

## Out of scope

- go-engine's surfaces (`SetSurface`): a surface can paint draw2d (`context.Paint` gives its area
  and draw list), so the touch screen and the one-screen menus can move onto surfaces, placed per
  target in engine.json. Left for a later change: this one keeps their layout in the game.
- The web build: it will take the one-screen layout when it exists.

## Design

- `touchscreen`: `Menu.Paint` and `Menu.Press`, `ModeChoice` and `ModeTouched` take the screen
  they lay out in. How many rows show comes from its height. The starry ground is painted by
  whoever owns a whole screen (the 3DS), not by the pieces.
- `engine/layout.go`: `oneScreen()` tells the layout from `engine.RunningOn()`. On one screen the
  overlay is painted at a smaller pixel-art scale, so there are more of its pixels to lay out, and
  the menus get a window sized for it. A click lands in those same pixels.
- The radar on one screen: the sphere is placed in the corner (`engine.PlaceSecondScreen`), and
  the readings move out of the touch screen's picture, into the HUD and the hints.
- Actions: `engine.json` binds `fire`, `menu`, `confirm`, `back`, the cross (which also zooms the radar), the
  menus' cross and pages, and the flight's axes, for `desktop` and `nintendo3ds`. The game reads
  `keys.Action`. The gyroscope and the curve of the sticks stay in code, since they are not
  buttons.

## Tasks

- [x] T1 — The pieces lay out in the screen they are given · touchscreen/ · R1 R4
- [x] T2 — One screen's layout: the modes, the menu, the lobby and the account over the game · engine/ · R2
- [x] T3 — One screen's radar and readings · engine/, hud/ · R3
- [x] T4 — Actions in engine.json, and the layout from RunningOn · engine/ · R5

## Result

Built and played on Linux (1280 × 720, 900 × 960, 700 × 600) and in Azahar:

- On one screen the modes and the menu are windows over the game, sized to the window and to
  their rows, with a keyboard's hints. The 3DS's screens are as they were.
- The radar's sphere floats over the game in the bottom right corner, with no background
  (go-engine's `PlaceSecondScreenOver`, which go-engine's session added for this). Its zoom is above
  it, the speed beside it, the score under the life bar, the keys along the bottom.
- The keys are engine.json's actions on both targets. A test checks that every action the code
  reads is bound, and another that no key does two things in flight: the arrows once climbed and
  zoomed the radar at once, so on the desktop the radar zooms with + and −, as in the browser.
- The layout is chosen by `engine.RunningOn()`: the build files no longer say whether there is a
  touch screen.

Not done: the surfaces (see Out of scope), a pad's sticks were not tried, and nothing was played
on a real console. go-engine is taken from its main (66a7e73), which has the floating second screen.
