package main

import (
	"fmt"
	"image/color"
	"os"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/cops/terminal"
	"github.com/borkshop/bork/internal/rectangle"
	opensimplex "github.com/ojrac/opensimplex-go"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func run() error {

	term := terminal.New(os.Stdout.Fd())
	defer term.Restore()
	term.SetRaw()

	bounds, err := term.Bounds()
	if err != nil {
		return err
	}

	front, back := display.New2(bounds)

	w := 3
	h := 4

	white := color.RGBA{192, 198, 187, 255}
	blue := color.RGBA{2, 50, 145, 255}

	var buf []byte
	cur := display.Start
	buf, cur = cur.Home(buf)
	buf, cur = cur.Clear(buf)
	buf, cur = cur.Hide(buf)

Loop:
	for {
		front.Fill(front.Bounds(), " ", blue, blue)
		walls := rectangle.Inset(bounds, 6, 3)
		front.Fill(walls, " ", white, white)
		floor := rectangle.Inset(walls, 3, 1)

		// generate floor tile noise
		noise := opensimplex.NewWithSeed(0)
		for y := walls.Min.Y; y < walls.Max.Y; y++ {
			for x := walls.Min.X; x < walls.Max.X; x++ {
				n := noise.Eval2(float64((x+1)/4), float64(y/2))
				o := uint8(0)
				if ((x+1)/4+(y/2))&1 == 0 {
					o = 5
				}
				c := color.Gray{uint8(n*10) + (255 - 15) + o}
				front.Foreground.Set(x, y, c)
				front.Background.Set(x, y, c)
			}
		}

		for y := floor.Min.Y; y < floor.Min.Y+h; y++ {
			for x := floor.Min.X; x < floor.Min.X+w*2; x += 2 {
				front.Text.Set(x, y, "👨‍👩‍👧‍👧")
			}
		}

		buf, cur = display.RenderOver(buf, cur, front, back, display.Model24)
		front, back = back, front
		os.Stdout.Write(buf)
		buf = buf[0:0]

		var rbuf [1]byte
		os.Stdin.Read(rbuf[0:1])

		switch rbuf[0] {
		case ' ':
			back.Clear(back.Bounds())
		case 'q':
			break Loop
		case 'j':
			h++
		case 'k':
			h--
		case 'h':
			w--
		case 'l':
			w++
		case 'c':
			buf, cur = cur.Clear(buf)
			back.Clear(back.Bounds())
		}

	}

	buf, cur = cur.Home(buf)
	buf, cur = cur.Clear(buf)
	buf, cur = cur.Show(buf)
	os.Stdout.Write(buf)
	buf = buf[0:0]
	return nil
}
