package main

import (
	"fmt"
	"image"
	"os"

	"github.com/borkshop/bork/internal/bitmap"
	"github.com/borkshop/bork/internal/cops/braille"
	"github.com/borkshop/bork/internal/cops/display"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func run() (err error) {
	w, h := 32, 16
	pb := image.Rect(0, 0, w, h)
	bb := braille.Bounds(pb, braille.Margin)
	front := display.New(pb)
	bmp := bitmap.New(bb)

	for y := 0; y < h*6; y++ {
		for x := 0; x < w*3; x++ {
			if x == y || x+y*2/3 == 50 {
				bmp.Set(x, y, true)
			}
		}
	}

	braille.DrawBitmap(front, pb, bmp, image.ZP, braille.Margin, display.Colors[7])

	var buf []byte
	cur := display.Reset
	buf, cur = display.Render(buf, cur, front, display.Model0)
	buf = append(buf, "\r\n"...)
	_, err = os.Stdout.Write(buf)

	return err
}
