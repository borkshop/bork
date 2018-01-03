package main

import (
	"fmt"
	"image"
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
	"github.com/borkshop/bork/internal/parking"
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

	// Animation interval, 60Hz
	ticker := time.NewTicker(time.Second / 60)

	bork := newBork(term.Display.Bounds())

	for {
		now := time.Now()

		bork.splash(term.Display, now)

		if err := term.Render(); err != nil {
			return err
		}

		select {
		case <-ticker.C:
		case <-sigwinch:
			bork = newBork(term.Display.Bounds())
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

type bork struct {
	now time.Time
	lot *parking.Lot
}

func newBork(bounds image.Rectangle) *bork {
	_, lower := rectangle.SplitHorizontal(bounds)
	lot := parking.NewLotForBounds(lower)
	return &bork{
		lot: lot,
		now: time.Now(),
	}
}

func (b *bork) splash(d *display.Display, t time.Time) {
	for b.now.Before(t) {
		b.lot.Tick()
		b.now = b.now.Add(500 * time.Millisecond)
	}

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
	msg := "B  Ø  R  K"
	msgbox := rectangle.MiddleCenter(text.Bounds(msg), borkline)
	text.Write(d, msgbox, msg, borkmark.Yellow)

	borkline = borkline.Add(image.Pt(0, 2))
	msg = "█  █  █  █"
	msgbox = rectangle.MiddleCenter(text.Bounds(msg), borkline)
	text.Write(d, msgbox, msg, borkmark.Yellow)

	b.lot.Draw(d, lower)
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
