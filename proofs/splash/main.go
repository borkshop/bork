package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"os"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/cops/terminal"
	"github.com/borkshop/bork/internal/cops/text"
	"github.com/borkshop/bork/internal/input"
	"github.com/borkshop/bork/internal/rectangle"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}

var (
	white   = color.RGBA{192, 198, 187, 255}
	yellow  = color.RGBA{213, 179, 42, 255}
	smog    = color.RGBA{20, 185, 255, 255}
	blue    = color.RGBA{2, 50, 145, 255}
	asphalt = color.RGBA{29, 33, 48, 255}
)

func run() error {

	term := terminal.New(os.Stdout.Fd())
	defer term.Restore()
	term.SetRaw()

	bounds, err := term.Bounds()
	if err != nil {
		return err
	}

	front := display.New(bounds)

	var buf []byte
	cur := display.Start
	buf, cur = cur.Home(buf)
	buf, cur = cur.Clear(buf)
	buf, cur = cur.Hide(buf)

	commands, mute := input.Channel(os.Stdin)
	defer mute()

Loop:
	for {
		splash(front)

		buf, cur = display.Render(buf, cur, front, display.Model24)
		buf, cur = cur.Reset(buf)
		os.Stdout.Write(buf)
		buf = buf[0:0]

		select {
		case command := <-commands:
			switch c := command.(type) {
			case rune:
				switch c {
				case 'q':
					break Loop
				}
			}
		}

	}

	buf, cur = cur.Home(buf)
	buf, cur = cur.Clear(buf)
	buf, cur = cur.Show(buf)
	os.Stdout.Write(buf)
	buf = buf[0:0]
	return nil
}

func splash(d *display.Display) {

	bounds := d.Bounds()

	upper, lower := rectangle.SplitHorizontal(bounds)

	bork := upper
	bork.Min.Y = upper.Max.Y - 4
	d.Fill(upper, " ", smog, smog)
	d.Fill(lower, " ", asphalt, asphalt)
	d.Fill(bork, " ", blue, blue)

	borkline := bork
	borkline.Max.Y = borkline.Min.Y + 1

	borkline = borkline.Add(image.Pt(0, 1))
	msg := "B Ø R K"
	msgbox := rectangle.MiddleCenter(text.Bounds(msg), borkline)
	text.Write(d, msgbox, msg, yellow)

	borkline = borkline.Add(image.Pt(0, 2))
	msg = "█ █ █ █"
	msgbox = rectangle.MiddleCenter(text.Bounds(msg), borkline)
	text.Write(d, msgbox, msg, yellow)

	parkingbox := image.Rect(0, 0, 9, 2)

	lower.Min.X--
	lower.Min.Y += 2
	parking := display.New(parkingbox)
	text.Write(parking, parkingbox, "──┬──\n  │  ", yellow)
	for x := lower.Min.X; x < lower.Max.X; x += parkingbox.Max.X {
		for y := lower.Min.Y; y < lower.Max.Y; y += parkingbox.Max.Y {
			at := image.Rect(x, y, lower.Max.X, lower.Max.Y)
			display.Draw(d, at, parking, image.ZP, draw.Over)
			d.Set(x, y+1, car(), asphalt, asphalt)
			d.Set(x+3, y+1, car(), asphalt, asphalt)
		}
	}

	lower.Min.Y += parkingbox.Max.Y
	text.Write(parking, parkingbox, "──┼──\n  │  ", yellow)
	for x := lower.Min.X; x < lower.Max.X; x += parkingbox.Max.X {
		for y := lower.Min.Y; y < lower.Max.Y; y += parkingbox.Max.Y {
			at := image.Rect(x, y, lower.Max.X, lower.Max.Y)
			display.Draw(d, at, parking, image.ZP, draw.Over)
		}
	}
}

func car() string {
	switch rand.Intn(10) {
	case 0, 1:
		return "🚙"
	case 2, 3:
		return "🚗"
	case 4:
		return "🚕"
	default:
		return ""
	}
}
