# 5. Ship parts as the browser shows them

> Status: building · Author's approval: proposal ☑ plan ☑

## Why

Since jonbul's versions 0.6 to 0.8, the ship designer gives parts a texture, an opacity, and two
new kinds: lights and imported `.glb` models. The game reads none of them, so ships come out
plain, and lights and models don't show at all. A ship should look the same in the game as in the
browser, as far as each platform can. This takes over proposal 3's T12 and T13.

## What

| Id | Requirement | Check |
|---|---|---|
| R1 | A light part shows as a bulb glowing in its colour, mirrored like any part | Test: a design with a light makes a glowing mesh; hangar screenshot |
| R2 | A part with an opacity is see-through by that much | Test: the part's material is transparent with that alpha; screenshot |
| R3 | A model part shows its `.glb`, fitted to its unit cube, fetched once and kept | Test: a design with a small model makes its mesh; the Lada Niva in the hangar |
| R4 | Lights shine on the ship: all of them on the desktop, and on the 3DS only as many as keep 60 fps (none, measured) | Played; STATS |
| R5 | Once go-renderer has image textures, the 8 built-in textures and `textureUrl` images tile and are tinted by the part's colour, as in the browser | Test per texture; a textured ship beside the browser |
| R6 | A part the game can't show yet is left out, and the rest of the ship still shows | Test: an unknown type doesn't break the design |

## Out of scope

- go-renderer's image textures themselves: go-renderer's proposal 2 (textures on meshes), waiting for the author. R5 waits for it.
- Imitating the textures with go-renderer's patterns: they are drawn on the desktop only, and none
  looks like the browser's.
- The designer: the game only shows ships; designing stays in the browser.
- Textures inside `.glb` models, until go-renderer reads them.

## Design

- `protocol.Part` learns the new fields: `texture`, `textureUrl`, `opacity`, `intensity`,
  `modelId`.
- In `ship`:
  - a light is a sphere in an unlit material, a child node of the ship's (go-renderer's `GlowKind`
    is shaded by the sun, and shines only in the dark);
  - opacity is a `Transparent` material with that alpha, one mesh for each twentieth of opacity;
  - a texture is an image once go-renderer has them (R5); until then a part shows its colour.
  - Parts are grouped by material, so a ship is still a few meshes and draws.
- **Real light (R4).** Each lamp is a go-renderer point light, placed by its ship every frame and
  handed to go-engine with `Scene.SetPointLights` (added for this). It reaches six design units.
  - On the desktop every ship's lamps shine, up to 64 at once.
  - On the 3DS none do: light is baked on the processor for every mesh a lamp reaches, and even
    our own ship's lamps took a New 3DS (in Azahar) from 60 frames a second to 20 while it turned.
    The bulbs still glow.
- **Models (R3).** `loaders.ReadScene` reads the `.glb` from the server's bytes. Its meshes are
  fitted to the part's unit cube, as the browser's `models.ts` does. They are fetched in the
  background, once each per run, and kept per `modelId` in memory; until one arrives the rest of
  the ship shows. A model may have 100,000 triangles on the desktop and 6,000 on the 3DS.
- Not taken: waiting for image textures before anything else. Lights and models are missing
  parts, so they come first.

## Tasks

- [x] T1 — The new fields in `protocol.Part`; unknown types left out · protocol/, ship/ · R6
- [x] T2 — Lights as glowing bulbs · ship/ · R1
- [x] T3 — Opacity · ship/ · R2
- [x] T4 — `.glb` parts: fetch, cache, fit · ship/, account/, engine/ · R3 — kept in memory only: not yet on the SD card
- [x] T5 — Real lights, per platform, measured · engine/, arena/ · R4 — go-engine pinned to its `feature/program-lights` until it is on main
- [ ] T6 — Image textures and `textureUrl`, once go-renderer has them · ship/ · R5

## Result

Not built yet.
