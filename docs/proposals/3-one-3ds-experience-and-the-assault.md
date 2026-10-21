# 3. One 3DS experience, and the assault

> Status: building · Author's approval: proposal ☑ plan ☑

## Why

The repository has two 3DS programs that no longer look alike. The hand-written one
(`cmd/ships3ds`) uses both screens everywhere: a moving background and "TOUCH OR PRESS A" under
the title, the hangar's ship turning on the touch screen, the radar with kills, zoom and the menu
under the game. The go-engine one (`engine/`) leaves the bottom screen empty, except for a plain
list in the hangar.

Meanwhile the server and the browser (jonbul's versions 0.6.0 to 0.8.0) gained things neither 3DS
program knows:

- the assault mode, with lobbies, teams, gunners in turrets and a mothership;
- crashes between ships;
- textures, lights and opacity on ship parts;
- imported `.glb` models as parts.

A player should get the same game on the console as in the browser, whichever program they run.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | The engine game on the 3DS shows exactly what the hand-written program shows on both screens: title, hangar, arena, pause and settings, the same pixel-art and sounds | Azahar screenshots of both, screen by screen, side by side |
| R2 | The engine game's touch screen works as the hand-written one's: touch to start, arrows on the hangar, the radar's zoom, the menu | Played in Azahar with the mouse as the stylus |
| R3 | On the 3DS, the account (server, log in, log out) is the hand-written program's pixel-art screen with the system keyboard, not a go-ui list | Screenshot; logging in on Azahar against a local server |
| R4 | The mode screen offers the assault: join its lobby, see the players and the countdown, play as attacker or gunner, and see the round's result, as the browser does | Played against a local server with 4 bots; tests on the lobby and round messages |
| R5 | In the assault: the mothership at the centre, the turrets, the teams' colours, the mothership's life on the HUD, and a gunner aiming from their turret | Played; a test builds the mothership's mesh from the server's design |
| R6 | Crashes: running into a ship (or the mothership) sends `crash`, and both programs show the explosion | Test: a crash is reported once; played with a bot |
| R7 | Ship parts show their textures, lights and opacity as far as the console can, and `.glb` parts show their model | Test: a design with each kind of part makes its mesh; the Lada Niva and a textured ship in the hangar |
| R8 | The desktop game is the same as the 3DS's top screen, plus the radar and the account where the desktop has room for them | Played on Linux side by side with Azahar |
| R9 | Still 60 frames a second on a New 3DS, also in an assault with 20 players | STATS on the console |

## Out of scope

- Changing the server or the browser: the 3DS follows their protocol as it is.
- Android: it builds from the same project, but this change is not played or checked there.
- New features the browser doesn't have.

## Decisions

- **One 3DS program.** Once the engine game matches the hand-written one and plays at 60 fps on a
  real New 3DS, `cmd/ships3ds` is removed (git history keeps it). The assault and the new parts are
  built only into the engine game.
- **The match comes first** (R1–R3), then the assault and the rest (R4–R7).

## Design

**The touch screen becomes a package.** Today the hand-written program's bottom screen is in
`cmd/ships3ds`: pixel-art windows (pixelart.go), the menu and its pages (menu.go, panel.go,
bitmap.go), the map (map.go), the figures (figures.go) and the turning ship (preview.go). These
move into a package `touchscreen/`, which draws with go-ui and draw2d only. The hand-written
program and the engine game both use it, so the two can be compared screen by screen until the
hand-written one goes. The top screen's pieces are already shared (`arena`, `hud`, `pixelart`).

**How the engine game shows it.** On the 3DS:

- the touch screen's pages are drawn through `engine.SetInterface`, which go-engine lays out on the
  second screen;
- the 3D under them (the hangar's turning ship, the radar) is a scene of its own,
  `engine.SetSecondScreen`, as the radar already is.

The screens follow the hand-written program:

- **Title:** space and the name on top; the moving background and "TOUCH OR PRESS A" below.
- **Mode:** the space on top; the two modes as pixel-art cards on the touch screen.
- **Hangar:** the ship's name on top; the ship turning on its platform below, with the arrows. The
  account is a page of the menu.
- **Arena:** the game on top; the radar, kills, zoom and the menu below.

On the desktop, the same screens are put on one: the top screen's picture, with the touch
screen's pages in a window over it (R8).

**The assault.** It is a second connection, `/ws/assault`, in `match` (the free for all stays
`/ws`). `protocol` learns its messages, as the browser's `protocol.ts` has them. `game` learns:

- the lobby, the teams and the round;
- the gunner's turret (a fixed place on the hull; aiming turns the view);
- the mothership's life.

`arena` builds the mothership from the server's design (it comes with the round, as the browser
reads it). Behaviours `Lobby` and `Assault` join the scenes.

**Crashes** are checked in `game` against the hit spheres the server uses, and reported once, with
`crash {with}`.

**The new parts.** `ship` learns textures (go-renderer materials with an image), opacity (blended
materials), lights (go-renderer point lights, at most a few per scene on the 3DS) and `.glb`
models (go-renderer's glTF loader, fetched from the server once and kept).

**Not taken:** keeping both 3DS programs and building everything twice. The author chose one.

## Tasks

- [x] T1 — Move the touch screen into `touchscreen/`; `cmd/ships3ds` uses it unchanged · touchscreen/, cmd/ships3ds · R1 · test: `TestEveryMenuPageHasATitleAndAnIcon`; Azahar screenshots unchanged
- [x] T2 — Engine game, title and mode: the bottom screen as the hand-written one's, touch to start · engine/screens.go · R1 R2 · screenshots side by side
- [x] T3 — Engine game, hangar: the name on top, the turning ship and the arrows below, touch on the arrows · engine/ships.go · R1 R2 · screenshots side by side — the own designs come first, as in the browser (the hand-written program put them last)
- [x] T4 — Engine game, account as a menu page with the system keyboard · engine/account_nintendo3ds.go · R3 · logged in on Azahar — the pages seen and turned in Azahar; the system keyboard is still to try on a console
- [x] T5 — Engine game, arena: the radar with kills, zoom and the menu (pause, settings, figures) on the touch screen · engine/pilot.go · R1 R2 · screenshots; played
- [x] T6 — The desktop: the touch screen's pages in a window over the top screen's picture · engine/ · R8 · played on Linux
- [x] T7 — 60 fps on a New 3DS (the author plays it); then remove `cmd/ships3ds` and update the README and the release · R1 R9 · STATS — both programs at 60 fps in a match on a New 3DS XL (one player), read from their logs through Rosalina
- [x] T8 — Crashes · game/, protocol/ · R6 · tests: `TestRunningIntoAShipIsReportedAgainOnlyEveryQuarterSecond`, `TestPassingCloseToAShipIsNoCrash`, `TestADeadShipCannotBeRunInto` — as the browser: crash spheres of 0.7 ship radii, reported again every 0.25 s until the server agrees; the mothership waits for T10; not yet played against a bot
- [x] T9 — The assault's protocol and lobby · protocol/, match/, game/ · R4 · tests on the lobby, round and result messages — also played: four clients on a local ships3D-api got a round with three attackers and a gunner in a turret
- [x] T10 — The mothership, turrets, teams and the gunner · hull/, arena/, game/, engine/ · R5 · test: `TestTheMothershipIsBuiltFromItsDesign`; played with 4 bots — played as an attacker on the desktop and in Azahar with three bots on a local ships3D-api; the gunner's view is tested but not yet played (the teams are dealt at random); Azahar dropped to 53 fps for a moment with the mothership in sight (T14)
- [x] T11 — The assault's screens: lobby, countdown, result; the mothership's life on the HUD · engine/, hud/ · R4 R5 · screenshots beside the browser
  Played on the desktop against three bots: the lobby lists who waits and counts down, the round
  opens with our role, and the HUD shows the mothership's life and the time left (a gunner has no
  life bar). The result screen is tested but not yet seen in play: a round lasts five minutes.
- [ ] T12 — Textures, opacity and lights on parts · ship/ · R7 · tests per kind of part
- [ ] T13 — `.glb` parts · ship/ · R7 · test: a design with a model part makes its mesh
- [ ] T14 — 60 fps in an assault with 20 players on a New 3DS · R9 · STATS

## Result

Not built yet.
