package main

import (
	"log"

	"github.com/borkshop/bork/internal/ecs"
	"github.com/borkshop/bork/internal/perf"
	"github.com/borkshop/bork/internal/view"
	"github.com/borkshop/bork/internal/view/hud"
)

type ui struct {
	hud.Logs
	perf.Dash
}

// const (
// 	XXX ecs.ComponentType = 1 << iota
// )

// const (
// 	XXX = YYY | ZZZ
// )

type world struct {
	perf perf.Perf
	ui
	grid view.Grid

	ecs.System
	// TODO: your state here

	// TODO: if you're going to want to position things (and who doesn't),
	// you'll probably want:
	// eps eps.EPS
}

func (w *world) init() {
	w.Logs.Init(1000)
	w.perf.Init("skeleton", w)
	w.ui.Perf = &w.perf
}

func (w *world) Render(termGrid view.Grid) error {
	hud := hud.HUD{
		Logs:  w.Logs,
		World: w.grid, // TODO render your world grid and pass it here
	}

	hud.AddRenderable(&w.ui.Dash, view.AlignRight|view.AlignBottom|view.AlignHFlush)

	// TODO add more hud elements

	hud.Render(termGrid)
	return nil
}

func (w *world) Close() error {
	// TODO shutdown any long-running resources

	return nil
}

func (w *world) HandleKey(k view.KeyEvent) error {
	// TODO do something with it

	return nil
}

func main() {
	if err := view.JustKeepRunning(func(v *view.View) (view.Client, error) {
		var w world
		w.init()

		return &w, nil
	}); err != nil {
		log.Fatal(err)
	}
}
