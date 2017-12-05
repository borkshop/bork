package main

import (
	"log"
	"math/rand"

	"github.com/borkshop/bork/internal/point"
	"github.com/borkshop/bork/internal/view"
	"github.com/borkshop/bork/internal/view/hud"
)

type world struct {
	hud.Logs

	table []rune
	grid  view.Grid
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

	// TODO call hud methods to build a basic UI, e.g.:
	hud.HeaderF("<left1")
	hud.HeaderF("<left2")
	hud.HeaderF(">right1")
	hud.HeaderF(">right2")
	hud.HeaderF("center by default")

	hud.FooterF("use h/j/k/l and y/u/b/n to change grid size")
	hud.FooterF(">one")
	hud.FooterF(">two")
	hud.FooterF(".>three") // the "." forces a new line

	// NOTE more advanced UI components may use:
	// hud.AddRenderable(ren view.Renderable, align view.Align)

	hud.Render(termGrid)
	return nil
}

func (w *world) Close() error {
	// TODO shutdown any long-running resources

	return nil
}

func (w *world) HandleKey(k view.KeyEvent) error {
	switch k.Ch {
	case 'h':
		w.generate(w.grid.Size.Add(point.Pt(-1, 0)))
	case 'l':
		w.generate(w.grid.Size.Add(point.Pt(1, 0)))
	case 'j':
		w.generate(w.grid.Size.Add(point.Pt(0, -1)))
	case 'k':
		w.generate(w.grid.Size.Add(point.Pt(0, 1)))
	case 'y':
		w.generate(w.grid.Size.Add(point.Pt(-1, 1)))
	case 'u':
		w.generate(w.grid.Size.Add(point.Pt(1, 1)))
	case 'b':
		w.generate(w.grid.Size.Add(point.Pt(-1, -1)))
	case 'n':
		w.generate(w.grid.Size.Add(point.Pt(1, -1)))
	}
	return nil
}

func main() {
	if err := view.JustKeepRunning(func(v *view.View) (view.Client, error) {
		var w world
		w.Logs.Init(1000)

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
