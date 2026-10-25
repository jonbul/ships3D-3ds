# 2. Ships 3D as a go-engine game

> Status: building · Author's approval: proposal ☑ plan ☑

## Why

Ships 3D's 3DS client (this repository) is a program written by hand:
its own loop, its own scene, its own menus. go-engine now does much of that for any game: scenes
edited in an editor, behaviours in Go, cameras, lights, physics, the bottom screen, and builds for
the desktop, Android and the 3DS from one project. Moving Ships 3D onto it makes the game easier
to change (the hangar, the arena and the HUD become scenes and behaviours), puts it on the desktop
and on phones for free, and is the first online, 3D, two-screen game to test go-engine with.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | A go-engine project, Ships 3D, opens in the editor: scenes for the title, the hangar and the arena, laid out there | The project opens with `go-engine <folder>`; the three scenes show in the editor |
| R2 | The same game runs on the desktop, the 3DS and Android, built from the editor's Build | A build for each; desktop and 3DS played against a server |
| R3 | It plays online with the browser's players, as `client3ds` does: log in, the player's ships, join, fly, fire, hits, deaths, respawns | Played against a local server with bots, and against the real one from the 3DS |
| R4 | The ships are built from their designs (the designer's parts), as the browser shows them | Test: a design's mesh has the parts' triangles; a ship seen in the hangar and the arena |
| R5 | The same experience as the browser and the 3DS today: the title with a ship crossing it, the mode, the ship in the hangar, the game, the pause menu | Played on the 3DS and the desktop, side by side with the browser |
| R6 | The arena: the chase camera, the stars and planets, the world's edge as a force field, the copilot turning the ship back, the HUD (the 3DS's life bar) | Played; the edge's behaviour checked by the existing game tests |
| R7 | On the 3DS: the 3D radar and the menu on the touch screen, stereoscopic 3D, the gyroscope | Played on a New 3DS |
| R8 | Sound: chiptune effects, music and the engine's hum, on every platform | Heard on the desktop and the 3DS |
| R9 | The game's logic is not copied: the engine game uses this repository's plain-Go packages (protocol, game, flight, ship, radar, chase, impact, chiptune, account, profile) | It lives in this repository and imports them; no copies |
| R10 | 60 frames a second on a New 3DS with a dozen players | STATS on the console |
| R11 | On the desktop, a player logs in, picks the server and plays with a mouse and keyboard, as in the browser | Played on Linux; builds for Windows and macOS |
| R12 | Releases carry the desktop game and the `.3dsx`, with their third-party notices | A release with all four files |

## Out of scope

- The assault mode, ship textures, the `crash` message and the language setting (each can follow,
  in this project or in `client3ds`).
- Removing or replacing the hand-written 3DS program (`cmd/ships3ds`): it stays as it is until the
  engine game plays as well.
- An old (non-New) 3DS.
- go-engine's own docs and versions: the go-engine session looks after those.

## Design

**Where.** In this repository, as a go-engine project in `engine/` (made with `go-engine new`):
its scenes, behaviours and `main.go`, in the same module, `github.com/jonbul/ships3D-3ds`.

**What it reuses.** The repository's plain-Go packages hold the game: the protocol, the game's
state, the flight model, ships' meshes, the radar, the camera, crash effects, sounds made in code,
login and the saved profile. Being in the same module, `engine/` imports them directly.
Alternatives: copy the packages (breaks R9), or a repository of its own (another module to keep in
step).

**What the game adds.** Behaviours (`Hangar`, `Pilot`, `Arena`, `Radar`, `Copilot`, `Engine`),
the scenes, and a network piece: on the 3DS it dials with go-3ds `network` (TCP and TLS); on the
desktop and Android, with Go's `net` and `crypto/tls`. Both satisfy the transport `websocket` and
`web` already take.

**What go-engine needs**, asked of the go-engine session, which makes them in the engine (the
author's choice); the game uses the versions it releases:

| Gap | Why Ships needs it |
|---|---|
| A node whose mesh is made by the game's code (`Node.SetMesh`) | Ships are built from their designs at run time, not imported `.glb` files |
| Nodes marked as moving (go-renderer's `Node.Moving`) | Ships turn every frame; baked lighting stalls the 3DS |
| A 3D scene on the bottom screen as well as a 2D one | The 3D radar and the hangar's turning ship |
| A 2D layer over the 3D top screen, on both eyes | The life bar and the copilot |
| Sound through go-sound on every platform | R8 |
| The 3DS's gyroscope, C-stick and START in `engine.Input` | Controls the 3DS client has today |

If a gap can't be closed in go-engine soon, the game draws that piece itself with go-renderer, as
`client3ds` does, and the proposal says so.

**Alternatives not taken.** Keep extending `client3ds` alone: it works, but it stays 3DS-only and
every screen is code. Rewrite the logic inside the engine's project: duplicates what is tested.

## Decided by the author

1. It lives in jonbul's repository for the 3DS client, ships3D-3ds, in `engine/`.
2. What go-engine lacks is asked of the go-engine session.
3. The experience is the same everywhere: the browser's and the 3DS's, which already match.

4. The desktop build is for players too: the login and server screens work with a mouse and a
   keyboard, and releases carry the desktop game (Linux, Windows, macOS) beside the `.3dsx`.

## Tasks

Each task leaves both games working (the hand-written `cmd/ships3ds` and `engine/`), with its
tests, and is a commit on `feature/go-engine`.

- [x] T1 — ask the go-engine session for what Ships needs (the table above), with why and how it is used; agree on versions · R2 · its answer written here
  - Answer (2026-10-21): v0.0.3 has keyboard input (held and pressed, the 3DS's buttons mapped to keys), a camera node behaviours can move (far plane fixed at 2000: a setting is missing) and a go-ui-only bottom screen. Everything else is missing and not planned yet: it needs a go-engine proposal the author approves. The engine's 3DS loop runs on go-ui's window.Run, not go-3ds's app, so whether go-sound gets its turn each frame is still to check. Until there is a plan, the engine's session advises building those parts on go-renderer directly.
- [x] T2 — a network transport for every platform: go-3ds `network` on the 3DS, Go's `net` and `crypto/tls` elsewhere, behind the interface `websocket` and `web` take · `transport/` · R3 · tests: `a message goes to a test server and back`, over TLS too
- [x] T3 — the project: `engine/` made with `go-engine new`, its `main.go` registering the game's behaviours, scenes for the title, the mode, the hangar and the arena · R1 · opens in the editor
- [x] T4 — ships from their designs as nodes (needs the engine's meshes from code) · R4 · test: `a design's node has its parts' triangles`
  - Done without waiting for the engine: the arena (package arena, shared with the 3DS program) builds its nodes inside the engine's own go-renderer scene (`Scene.World()`), ships Moving, as the 3DS program does. The engine's `Node.SetMesh` (proposal 9) can take over once released.
- [x] T5 — the arena: joining, flying, firing, hits, deaths and respawns through the existing `game` package · R3 · played against a local server with bots
- [x] T6 — the world around: stars, planets, dust, the edge's grid, the force field, the copilot and its autopilot, the crash's effects · R6 · played
- [x] T7 — the HUD: the 3DS's life bar, the copilot's box, the pause menu · R5, R6 · played
- [x] T8 — the title, the mode and the hangar, as the browser and the 3DS show them · R5 · played on the desktop and the 3DS
- [~] T9 — on the 3DS: the radar and the menus on the touch screen, stereoscopic 3D, the gyroscope and the C-stick · R7 · played on a New 3DS
- [x] T10 — sound: the chiptune effects, music and engine through go-sound · R8 · heard on the desktop and the 3DS
- [x] T11 — logging in and choosing the server on the desktop (mouse and keyboard) and the 3DS (its keyboard) · R11 · played
- [x] T12 — releases with the desktop games and the `.3dsx` and their notices; the README's how to play · R12 · a release — tried on the fork with test tags: the three desktop games and both `.3dsx` on one release
- [ ] T13 — frame rate on a New 3DS with a dozen players · R10 · numbers written here

## Result

Not started.
