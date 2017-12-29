package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"os"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/cops/terminal"
	"github.com/borkshop/bork/internal/cops/text"
	"github.com/borkshop/bork/internal/rectangle"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func run() (err error) {
	term := terminal.New(os.Stdin.Fd())
	defer func() {
		err = term.Restore()
	}()
	err = term.SetRaw()
	if err != nil {
		return
	}

	bounds, err := term.Bounds()
	if err != nil {
		return err
	}

	front := display.New(bounds)

	front.Fill(front.Bounds(), "/", color.RGBA{192, 0, 0, 255}, color.RGBA{30, 20, 40, 255})

	msg := "Press any key to continue..."
	msgbox := text.Bounds(msg)
	inset := rectangle.MiddleCenter(msgbox, bounds)
	outset := rectangle.Outset(inset, 4, 2)
	panel := display.New(outset)
	// Fill the panel with a translucent background color.
	draw.Draw(panel.Background, outset, &image.Uniform{color.NRGBA{63, 63, 127, 127}}, image.ZP, draw.Over)
	// Draw our text in the panel.
	text.Write(panel, inset, msg, display.Colors[7])
	display.Draw(front, outset, panel, outset.Min, draw.Over)

	var buf []byte
	cur := display.Start
	buf, cur = cur.Hide(buf)
	buf, cur = cur.Clear(buf)
	buf, cur = cur.Home(buf)
	buf, cur = display.Render(buf, cur, front, display.Model24)
	buf, cur = cur.Home(buf)
	_, err = os.Stdout.Write(buf)
	if err != nil {
		return err
	}
	buf = buf[0:0]

	var input [1]byte
	_, err = os.Stdin.Read(input[0:1])
	if err != nil {
		return err
	}

	buf, cur = cur.Home(buf)
	buf, cur = cur.Clear(buf)
	buf, cur = cur.Show(buf)
	_, err = os.Stdout.Write(buf)
	buf = buf[0:0]

	return err
}
