package main

import (
	"log/slog"

	"github.com/akrck02/go-engine/engine"

	"github.com/jonbul/ships3D-3ds/arena"
	"github.com/jonbul/ships3D-3ds/ship"
)

// Imported models (parts of type "model") are fetched from the server in the background, once
// each, and the designs that show one are built again when it arrives: until then their other
// parts show, as the browser shows a box in its place.

// Most triangles a model may have: the server allows a hundred thousand, which the desktop draws,
// but a 3DS holds a few ships' worth of them at most.
const (
	modelTrianglesOnTheDesktop = 100_000
	modelTrianglesOnThe3DS     = 6_000
)

// modelArrival is a model fetched and read, or why it couldn't be.
type modelArrival struct {
	id    string
	shape ship.Shape
	err   error
}

// modelFetcher fetches each model once, and hands the game what arrived, a frame at a time.
type modelFetcher struct {
	asked   map[string]bool
	arrived chan modelArrival
}

// models is the game's fetcher: models outlive the scenes, as the designs do.
var models = modelFetcher{asked: map[string]bool{}, arrived: make(chan modelArrival, 16)}

// want starts fetching the models not asked for yet.
func (fetcher *modelFetcher) want(ids []string) {
	for _, id := range ids {
		if fetcher.asked[id] {
			continue
		}
		fetcher.asked[id] = true
		go fetcher.fetch(id)
	}
}

// fetch fetches and reads one model; it runs on its own goroutine.
func (fetcher *modelFetcher) fetch(id string) {
	server, err := api()
	if err != nil {
		fetcher.arrived <- modelArrival{id: id, err: err}
		return
	}
	file, err := server.Model(id)
	if err != nil {
		fetcher.arrived <- modelArrival{id: id, err: err}
		return
	}
	maximum := modelTrianglesOnTheDesktop
	if engine.RunningOn() == engine.Nintendo3DSTarget {
		maximum = modelTrianglesOnThe3DS
	}
	shape, err := ship.ReadModel(file, maximum)
	fetcher.arrived <- modelArrival{id: id, shape: shape, err: err}
}

// receive adds what arrived since the last frame to the ships, and has world build again the
// designs that show it. A model that failed stays out, as the browser keeps its placeholder.
func (fetcher *modelFetcher) receive(world *arena.World) {
	for {
		select {
		case arrival := <-fetcher.arrived:
			if arrival.err != nil {
				slog.Warn("a ship's model could not be shown", "model", arrival.id, "error", arrival.err)
				continue
			}
			ship.AddModel(arrival.id, arrival.shape)
			world.ForgetModel(arrival.id)
		default:
			return
		}
	}
}
