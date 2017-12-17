package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"os/signal"
	"syscall"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/hilbert"
	"github.com/borkshop/bork/internal/input"
	"github.com/borkshop/bork/internal/point"
	opensimplex "github.com/ojrac/opensimplex-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}

var (
	white = color.RGBA{192, 198, 187, 255}
	blue  = color.RGBA{2, 50, 145, 255}
)

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

	sigwinch := make(chan os.Signal)
	signal.Notify(sigwinch, syscall.SIGWINCH)

	world := newWorld()
	var at image.Point

Loop:
	for {
		world.Draw(term.Display, at)

		if err := term.Render(); err != nil {
			return err
		}

		select {
		case <-sigwinch:
			if err := term.UpdateSize(); err != nil {
				return err
			}
		case command := <-commands:
			switch c := command.(type) {
			case input.Move:
				at = at.Add(image.Point(c))
			case input.ShiftMove:
				at = at.Add(point.MulRespective(image.Point(c), term.Display.Rect.Size()))
			case rune:
				switch c {
				case 'q':
					break Loop
				}
			}
			if at.X < 0 {
				at.X = 0
			}
			if at.Y < 0 {
				at.Y = 0
			}
		}

	}

	return nil
}

func newWorld() *world {
	noise := opensimplex.NewWithSeed(0)
	return &world{
		noise: noise,
	}
}

type world struct {
	noise *opensimplex.Noise
}

type tileType int

const (
	scale         = 1 << 30
	hstride       = 10
	vstride       = 6
	wallThickness = 1
)

const (
	room tileType = iota
	horizontal
	vertical
	corner
)

func (w world) tileType(x, y int) tileType {
	var t tileType
	if x%hstride < wallThickness {
		t |= horizontal
	}
	if y%vstride < wallThickness {
		t |= vertical
	}
	return t
}

func (w world) tileAt(x, y int) (int, int) {
	return x / hstride, y / vstride
}

func (w world) colorAt(x, y int) color.Color {
	rx, ry := w.tileAt(x, y)
	at := hilbert.Encode(image.Pt(rx, ry), scale)

	switch w.tileType(x, y) {
	case vertical:
		to := hilbert.Encode(image.Pt(rx, ry-1), scale)
		if at+1 != to && at-1 != to {
			return blue
		}
	case horizontal:
		to := hilbert.Encode(image.Pt(rx-1, ry), scale)
		if at+1 != to && at-1 != to {
			return blue
		}
	case corner:
		return blue
	}

	o := uint8(0)
	if (x+y)&1 == 0 {
		o = 5
	}
	n := w.noise.Eval2(float64(x), float64(y))
	return color.Gray{uint8(n*10) + (255 - 15) + o}
}

func (w world) Draw(d *display.Display, about image.Point) {
	// size := d.Bounds().Size()
	// size.X /= 2
	// rect := d.Bounds().Sub(size.Div(2))
	rect := d.Bounds()
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x += 2 {
			c := w.colorAt((x+about.X*2)/2, y+about.Y)
			d.Set(x, y, " ", c, c)
			d.Set(x+1, y, " ", c, c)
		}
	}
}
