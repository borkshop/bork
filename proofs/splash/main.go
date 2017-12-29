package main

import (
	"fmt"
	"image"
	"image/draw"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/borkshop/bork/internal/bitmap"
	"github.com/borkshop/bork/internal/borkmark"
	"github.com/borkshop/bork/internal/cops/braille"
	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/cops/text"
	"github.com/borkshop/bork/internal/input"
	"github.com/borkshop/bork/internal/rectangle"
	opensimplex "github.com/ojrac/opensimplex-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func run() (rerr error) {
	term, err := display.NewTerminal(os.Stdout)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := term.Close(); rerr == nil {
			rerr = cerr
		}
	}()

	commands, mute := input.Channel(os.Stdin)
	defer mute()

	sigwinch := make(chan os.Signal, 1)
	signal.Notify(sigwinch, syscall.SIGWINCH)

	ticker := time.NewTicker(time.Second / 60)

	for {
		splash(term.Display, time.Now())

		if err := term.Render(); err != nil {
			return err
		}

		select {
		case <-ticker.C:
		case <-sigwinch:
			if err := term.UpdateSize(); err != nil {
				return err
			}
		case command := <-commands:
			switch c := command.(type) {
			case rune:
				switch c {
				case 'q':
					return nil
				}
			}
		}
	}
}

func splash(d *display.Display, t time.Time) {
	const borkHeight = 4

	bounds := d.Bounds()

	upper, lower := rectangle.SplitHorizontal(bounds)

	bork := upper
	bork.Min.Y = upper.Max.Y - borkHeight
	d.Fill(upper, " ", borkmark.Smog, borkmark.Smog)
	d.Fill(lower, " ", borkmark.Asphalt, borkmark.Asphalt)
	d.Fill(bork, " ", borkmark.Blue, borkmark.Blue)

	skybox := upper
	skybox.Max.Y -= borkHeight
	fillClouds(d, skybox, t)

	borkline := bork
	borkline.Max.Y = borkline.Min.Y + 1

	borkline = borkline.Add(image.Pt(0, 1))
	msg := "B Ø R K"
	msgbox := rectangle.MiddleCenter(text.Bounds(msg), borkline)
	text.Write(d, msgbox, msg, borkmark.Yellow)

	borkline = borkline.Add(image.Pt(0, 2))
	msg = "█ █ █ █"
	msgbox = rectangle.MiddleCenter(text.Bounds(msg), borkline)
	text.Write(d, msgbox, msg, borkmark.Yellow)

	parkingbox := image.Rect(0, 0, 9, 2)

	lower.Min.X--
	lower.Min.Y += 2
	parking := display.New(parkingbox)
	text.Write(parking, parkingbox, "──┬──\n  │  ", borkmark.Yellow)
	for x := lower.Min.X; x < lower.Max.X; x += parkingbox.Max.X {
		for y := lower.Min.Y; y < lower.Max.Y; y += parkingbox.Max.Y {
			at := image.Rect(x, y, lower.Max.X, lower.Max.Y)
			display.Draw(d, at, parking, image.ZP, draw.Over)
			d.Set(x, y+1, car(), borkmark.Asphalt, borkmark.Asphalt)
			d.Set(x+3, y+1, car(), borkmark.Asphalt, borkmark.Asphalt)
		}
	}

	lower.Min.Y += parkingbox.Max.Y
	text.Write(parking, parkingbox, "──┼──\n  │  ", borkmark.Yellow)
	for x := lower.Min.X; x < lower.Max.X; x += parkingbox.Max.X {
		for y := lower.Min.Y; y < lower.Max.Y; y += parkingbox.Max.Y {
			at := image.Rect(x, y, lower.Max.X, lower.Max.Y)
			display.Draw(d, at, parking, image.ZP, draw.Over)
		}
	}
}

func fillClouds(d *display.Display, sky image.Rectangle, now time.Time) {
	t := int(now.UnixNano() * 10 / int64(time.Second))
	r := braille.Bounds(sky, braille.Margin)
	bmp := bitmap.New(r)
	a := opensimplex.NewWithSeed(0)
	b := opensimplex.NewWithSeed(100)
	c := opensimplex.NewWithSeed(200)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			shape := a.Eval2(float64(x+t*2)/40.0, float64(y)/10.0)
			detail := c.Eval2(float64(x), float64(y))
			broad := b.Eval2(float64(x+t/2)/80.0, float64(y)/20.0)
			if shape+detail > 0 && broad+detail > 0 {
				bmp.Set(x, y, true)
			}
		}
	}
	braille.DrawBitmap(d, sky, bmp, image.ZP, braille.Margin, borkmark.Blue)
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
