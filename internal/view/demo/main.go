package main

import (
	"image"
	"log"
	"math/rand"
	"os"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/input"
	"github.com/borkshop/bork/internal/perf"
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

	table []string
	dis   *display.Display
}

func (w *world) init() {
	w.Logs.Init(1000)
	w.perf.Init("demo", nil) // TODO pass your root ecs.Proc in here
	w.ui.Perf = &w.perf
}

func (w *world) generate(sz image.Point) {
	w.dis = display.New(image.Rectangle{Max: sz})
	for chs, i := w.table, 0; i < len(w.dis.Text.Strings); i++ {
		w.dis.Text.Strings[i] = chs[rand.Intn(len(chs))]
	}
}

func (w *world) Render(d *display.Display) error {
	hud := hud.HUD{
		Logs:  w.Logs,
		World: w.dis, // TODO render your world and pass its Display here
	}

	hud.AddRenderable(&w.ui.Dash, view.AlignRight|view.AlignBottom|view.AlignHFlush)

	// TODO call hud methods to build a basic UI, e.g.:
	hud.HeaderF("<left1")
	hud.HeaderF("<left2")
	hud.HeaderF(">right1")
	hud.HeaderF(">right2")
	hud.HeaderF("center by default")

	hud.FooterF("use h/j/k/l and y/u/b/n to change world size")
	hud.FooterF(".>right footer") // the "." forces a new line

	hud.Render(d)
	return nil
}

func (w *world) Close() error {
	// TODO shutdown any long-running resources

	return nil
}

func (w *world) HandleInput(cmd interface{}) error {
	if w.ui.Dash.HandleInput(cmd) {
		return nil
	}

	switch c := cmd.(type) {
	case input.RelativeMove:
		w.generate(w.dis.Rect.Size().Add(c.Point))
	}

	w.perf.Process()

	return nil
}

func main() {
	f, err := os.Create("debug.log")
	if err != nil {
		log.Fatalln(err)
	}
	log.SetOutput(f)

	if err := view.JustKeepRunning(func(v *view.View) (view.Client, error) {
		var w world
		w.init()

		w.table = []string{
			"_", "-",
			"=", "+",
			"/", "?",
			"\\", "|",
			",", ".",
			":", ";",
			"'", "\"",
			"<", ">",
			"[", "]",
			"{", "}",
			"(", ")",
			"!", "@", "#", "$",
			"%", "^", "&", "*",
		}

		w.Log("Hello World Of Democraft!")

		w.generate(image.Pt(64, 32))

		return &w, nil
	}); err != nil {
		log.Fatal(err)
	}
}
