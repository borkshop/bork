package main

import (
	"fmt"
	"image"
	"image/color"
	"os"

	"github.com/borkshop/bork/internal/cops/bitmap"
	"github.com/borkshop/bork/internal/cops/braille"
	"github.com/borkshop/bork/internal/cops/display"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func run() error {
	w, h := 32, 16
	pb := image.Rect(0, 0, w, h)
	bb := braille.Bounds(pb)
	front := display.New(pb)
	img := bitmap.New(bb, color.Black, color.White)

	for y := 0; y < h*6; y++ {
		for x := 0; x < w*3; x++ {
			if x == y || x+y*2/3 == 50 {
				img.Set(x, y, color.White)
			}
		}
	}

	braille.Draw(front, pb, img, image.ZP, color.White, display.Colors[8])

	var buf []byte
	cur := display.Reset
	buf, cur = display.Render(buf, cur, front, display.Model0)
	buf = append(buf, "\r\n"...)
	os.Stdout.Write(buf)

	return nil
}
