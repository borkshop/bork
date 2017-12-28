package main

import (
	"bytes"
	"fmt"
	"image"
	"log"

	"github.com/borkshop/bork/internal/ecs"
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

type fieldWriter struct {
	*bytes.Buffer
	any bool
}

func (fw *fieldWriter) Printf(mess string, args ...interface{}) {
	if fw.any {
		fw.WriteRune(' ')
	}
	fw.any = true
	fmt.Fprintf(fw.Buffer, mess, args...)
}

func main() {
	const showPlayerEntID = false
	world.AddProcFunc(func() {
		var buf bytes.Buffer
		for it := world.Iter(wcPlayerControl.All()); it.Next(); {
			ent := it.Entity()
			if buf.Len() > 0 {
				buf.WriteRune(' ')
			}
			buf.WriteRune('<')
			fw := fieldWriter{Buffer: &buf}
			if showPlayerEntID {
				fw.Printf("[%v]", ent.ID())
			}
			if pt, def := world.pos.Get(ent); def {
				fw.Printf("@%v", pt)
			}
			if move := world.moves.GetPendingMove(ent); move != ecs.NilEntity {
				fw.Printf("mag:%v", world.moves.Mag(move))
			}
			buf.WriteRune('>')
		}
		hud.status.SetLeft(buf.String())
	})

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
