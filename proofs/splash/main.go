package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/borkshop/bork/internal/cops/bitmap"
	"github.com/borkshop/bork/internal/cops/braille"
	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/cops/terminal"
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

	commands, mute := input.Channel(os.Stdin)
	defer mute()

	cur := display.Start
	var buf []byte

	ticker := time.NewTicker(time.Second / 60)

	sigwinch := make(chan os.Signal)
	signal.Notify(sigwinch, syscall.SIGWINCH)

Loop:
	for {

		cur = display.Start
		buf, cur = cur.Home(buf)
		buf, cur = cur.Clear(buf)
		buf, cur = cur.Hide(buf)

		bounds, err := term.Bounds()
		if err != nil {
			return err
		}

		front := display.New(bounds)

	Animation:
		for {
			splash(front, int(time.Now().UnixNano()/100000000))

			buf, cur = display.Render(buf, cur, front, display.Model24)
			buf, cur = cur.Reset(buf)
			os.Stdout.Write(buf)
			buf = buf[0:0]

			select {
			case <-ticker.C:
				continue Animation
			case <-sigwinch:
				continue Loop
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
	}

	buf, cur = cur.Home(buf)
	buf, cur = cur.Clear(buf)
	buf, cur = cur.Show(buf)
	os.Stdout.Write(buf)
	buf = buf[0:0]
	return nil
}

func splash(d *display.Display, t int) {
	const borkHeight = 4

	bounds := d.Bounds()

	upper, lower := rectangle.SplitHorizontal(bounds)

	bork := upper
	bork.Min.Y = upper.Max.Y - borkHeight
	d.Fill(upper, " ", smog, smog)
	d.Fill(lower, " ", asphalt, asphalt)
	d.Fill(bork, " ", blue, blue)

	skybox := upper
	skybox.Max.Y -= borkHeight
	fillClouds(d, skybox, t)

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

func fillClouds(d *display.Display, sky image.Rectangle, t int) {
	r := braille.Bounds(sky)
	img := bitmap.New(r, white, blue)
	a := opensimplex.NewWithSeed(0)
	b := opensimplex.NewWithSeed(100)
	c := opensimplex.NewWithSeed(200)
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			if a.Eval2(float64(x+t*2)/40.0, float64(y)/10.0)+c.Eval2(float64(x), float64(y)) > 0 &&
				b.Eval2(float64(x+t/2)/80.0, float64(y)/20.0)+c.Eval2(float64(x), float64(y)) > 0 {
				img.SetBit(x, y, true)
			}
		}
	}
	braille.DrawBitmap(d, sky, img, image.ZP, blue)
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
