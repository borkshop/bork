package main

import (
	"fmt"
	"image"
	"log"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/views"
)

var (
	app   = &views.Application{}
	hud   = &hudT{}
	world = &worldT{}
)

func init() {
	hud.init()
	app.SetRootWidget(hud)
	hud.keybar.addAction('?', "Help", help)
	hud.keybar.addAction('Q', "Quit", app.Quit)
	hud.keybar.addAction('R', "Reset", reset)
	world.init(app.PostFunc)
}

func help() {
	halp := views.NewTextArea()
	halp.SetLines([]string{
		`/ Movement : vi-style keys --------------\`,
		`|   y k u  :                             |`,
		`|    \|/   : h j k l -- usual directions |`,
		`|   h-.-l  : y u b n -- for diagonals    |`,
		`|    /|\   : .       -- to stay in place |`,
		`|   b j n  :                             |`,
		`\----------------------------------------/`,
	})
	hud.showModal(halp)
}

func reset() {
	world.Clear()
	world.addRoom(image.Rect(-6, -3, 7, 4))
	world.analyze()

	player := world.addChar("player1", '@', tcell.ColorLightGreen, image.ZP)
	player.Add(wcPlayerControl)
	world.Process()
}

func main() {
	if err := func() error {
		scr, err := tcell.NewScreen()
		if err != nil {
			return err
		}
		app.SetScreen(scr)
		app.PostFunc(func() {
			hud.status.SetRight(fmt.Sprintf("colors=%v", scr.Colors()))
		})
		app.PostFunc(reset)
		return app.Run()
	}(); err != nil {
		log.Fatalln(err)
	}
}
