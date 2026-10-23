# Ships 3D for the Nintendo 3DS

A client for the [Ships 3D game server](https://github.com/jonbul/ships3D-api), for the New Nintendo 3DS (homebrew, `.3dsx`): it
logs in, lists your ships and joins the same games as the browser, over the same websocket and
HTTPS API. The game is a [go-engine](https://github.com/akrck02/go-engine) project, `engine/`:
one game for the 3DS and the desktop (Linux, Windows, macOS), with scenes editable in go-engine's
editor. It started as a program written by hand for the console; git history keeps it.

**Status: experiment**, written with an AI assistant and pending review. The plans are in
[docs/proposals/](docs/proposals/).

| Title | Hangar | Game | At the edge |
| --- | --- | --- | --- |
| ![Title](docs/images/title.png) | ![Hangar](docs/images/hangar.png) | ![Game](docs/images/game.png) | ![At the edge](docs/images/edge.png) |

## Building

```bash
tinygo build -target=nintendo3ds -ldflags "-X runtime.heapMegabytes=64" -o ships3d.3dsx ./engine
go build -tags release -o ships3d ./engine   # the desktop game; without -tags release, with go-engine's editor
```

Every version tag builds them all on GitHub and puts them on its release
(`.github/workflows/release.yml`). The desktop game needs OpenGL's, X11's and ALSA's headers on
Linux (see the workflow).

It needs [TinyGo for 3DS](https://github.com/akrck02/tinygo-3ds) and devkitPro's libraries
(libctru, citro3d, and mbedtls from the portlibs for TLS). The console's pieces come from Go
modules: [go-3ds](https://github.com/akrck02/go-3ds) (the app loop, sockets, TLS, the SD card,
the keyboard), [go-renderer](https://github.com/akrck02/go-renderer) (3D on the console's GPU),
[go-ui](https://github.com/akrck02/go-ui) and [go-sound](https://github.com/akrck02/go-sound).

`runtime.heapMegabytes=64` leaves C its share of the console's memory: without it TinyGo takes
all it can, and the socket service doesn't find the megabyte it needs. The game starts on the
real server (`https://jonbul.ddns.net:3000`); the ACCOUNT page changes it, and a server written
with `https://` is talked to over TLS.

Copy `ships3d.3dsx` to `/3ds/` on the SD card (ftpd works) and open it from the Homebrew
Launcher. Sound needs the DSP's firmware, `sdmc:/3ds/dspfirm.cdc` (the DSP1 homebrew dumps it).

## Playing

The title, then the hangar: the cross (or a touch on either side) goes through the ships, A
flies the one turning on its platform, START opens the menu. Login and server are remembered on
the SD card, in `sdmc:/3ds/ships3d/profile.json`.

| Control | What it does |
| --- | --- |
| Circle pad | Nose and turn |
| R or A | Fire; A flies again after dying |
| C-stick, or left and right on the cross | Roll |
| ZR / ZL (also X / B) | Faster / slower |
| Tilting the console | Steers too, with *Steer by tilting* on; L centres it |
| Up / down on the cross | Radar zoom |
| START | The menu: PILOT, STATS, SETUP and ACCOUNT pages, L and R between them |

The top screen is the game in 3D, with your life at the top left; the touch screen is a 3D radar.
Near the world's edge the force field shows, and the copilot turns you back if you keep going.

## The code

| Package | What it holds |
| --- | --- |
| `websocket`, `web` | A small websocket client and HTTP over the console's sockets |
| `protocol`, `account` | The server's messages, login and the list of ships |
| `game`, `flight` | The game's state and the same flight model as the server's |
| `ship`, `hull`, `radar`, `chase`, `impact`, `tilt` | Ships' meshes, the mothership's shape, the radar, the camera, crash effects, the gyroscope |
| `chiptune`, `sounds`, `figures`, `profile` | Sounds made in code and played, frame figures, what is remembered |
| `transport`, `match` | Connections on every platform, and a match on the server frame by frame |
| `arena`, `hud`, `pixelart` | The space the game is seen in (sky, ships, hangar, radar, effects), the HUD, the pixel-art look |
| `touchscreen` | The touch screen: the menus and their pages, the radar's readings, the title's prompt, the hangar's arrows |
| `engine` | The go-engine game: its scenes and behaviours |

All but `engine` are plain Go, tested on a computer with `go test ./...` (the CI runs them).
