package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"time"

	"github.com/borkshop/bork/internal/cops/braille"
	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/cops/terminal"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func run() (err error) {
	term := terminal.New(os.Stdout.Fd())
	defer func() {
		err = term.Restore()
	}()
	err = term.SetRaw()
	if err != nil {
		return err
	}

	bounds, err := term.Bounds()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(16 * time.Millisecond)

	stopper := make(chan struct{}, 0)
	go func() {
		var rbuf [1]byte
		_, err = os.Stdin.Read(rbuf[0:1])
		close(stopper)
	}()

	img := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	var buf []byte
	cur := display.Start
	buf, cur = cur.Hide(buf)
	buf, cur = cur.Home(buf)
	buf, cur = cur.Clear(buf)

	front, back := display.New2(bounds)
	bb := braille.Bounds(bounds)

Loop:
	for {
		t := int(time.Now().UnixNano() / 10000000)

		for x := bb.Min.X; x < bb.Max.X; x++ {
			z := int(float64(bb.Dy()) * (0.5 + 0.30*math.Sin(float64(t+x)*math.Pi*2.0*2.0/float64(bb.Dx()))))
			for y := bb.Min.Y; y < bb.Max.Y; y++ {
				if y < z+2 && y > z-2 {
					img.Set(x, y, color.White)
				} else {
					img.Set(x, y, color.Black)
				}
			}
		}

		// Size that image down and write it in braille to the display.
		front.Fill(bounds, " ", color.Black, color.Transparent)
		braille.Draw(front, bounds, img, image.ZP, color.RGBA{191, 191, 127, 255}, color.Black)

		buf, cur = display.RenderOver(buf, cur, front, back, display.Model24)
		front, back = back, front
		_, err = os.Stdout.Write(buf)
		if err != nil {
			return err
		}
		buf = buf[0:0]

		select {
		case <-ticker.C:
		case <-stopper:
			break Loop
		}
	}

	ticker.Stop()

	buf, cur = cur.Home(buf)
	buf, cur = cur.Clear(buf)
	buf, cur = cur.Show(buf)
	_, err = os.Stdout.Write(buf)
	buf = buf[0:0]

	return err
}
