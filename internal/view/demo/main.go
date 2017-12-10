package main

import (
	"image"
	"log"
	"math/rand"

	"github.com/borkshop/bork/internal/input"
	"github.com/borkshop/bork/internal/perf"
	"github.com/borkshop/bork/internal/point"
	"github.com/borkshop/bork/internal/view"
	"github.com/borkshop/bork/internal/view/hud"
)

type ui struct {
	hud.Logs
	perf.Dash
}

type world struct {
	perf perf.Perf
	ui

	table []rune
	grid  view.Grid
}

func (w *world) init() {
	w.Logs.Init(1000)
	w.perf.Init("demo", nil) // TODO pass your root ecs.Proc in here
	w.ui.Perf = &w.perf
}

func (w *world) generate(sz point.Point) {
	w.grid = view.MakeGrid(sz)
	for chs, i := w.table, 0; i < len(w.grid.Data); i++ {
		w.grid.Data[i].Ch = chs[rand.Intn(len(chs))]
	}
}

func (w *world) Render(termGrid view.Grid) error {
	hud := hud.HUD{
		Logs:  w.Logs,
		World: w.grid, // TODO render your world grid and pass it here
	}

	hud.AddRenderable(&w.ui.Dash, view.AlignRight|view.AlignBottom|view.AlignHFlush)

	// TODO call hud methods to build a basic UI, e.g.:
	hud.HeaderF("<left1")
	hud.HeaderF("<left2")
	hud.HeaderF(">right1")
	hud.HeaderF(">right2")
	hud.HeaderF("center by default")

	hud.FooterF("use h/j/k/l and y/u/b/n to change grid size")
	hud.FooterF(".>right footer") // the "." forces a new line

	hud.Render(termGrid)
	return nil
}

func (w *world) Close() error {
	// TODO shutdown any long-running resources

	return nil
}

func (w *world) HandleKey(k view.KeyEvent) error {
	handled := w.ui.Dash.HandleKey(k)

	if !handled {
		if pt, ok := input.ParseMove(k.Ch, image.ZP); ok {
			w.generate(w.grid.Size.Add(point.Point(pt)))
			handled = true
		}
	}

	w.perf.Process()

	return nil
}

func main() {
	if err := view.JustKeepRunning(func(v *view.View) (view.Client, error) {
		var w world
		w.init()

		w.table = []rune{
			'_', '-',
			'=', '+',
			'/', '?',
			'\\', '|',
			',', '.',
			':', ';',
			'"', '\'',
			'<', '>',
			'[', ']',
			'{', '}',
			'(', ')',
			'!', '@', '#', '$',
			'%', '^', '&', '*',
		}

		w.Log("Hello World Of Democraft!")

		w.generate(point.Pt(64, 32))

		return &w, nil
	}); err != nil {
		log.Fatal(err)
	}
}
