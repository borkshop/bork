package main

import (
	"fmt"
	"image"
	"image/color"
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

	w, h := 16, 16
	pb := image.Rect(0, 0, w, h)
	ib := braille.Bounds(pb, image.ZP)
	rb := ib
	pb.Max.X += 2
	rb = rb.Add(image.Pt(2, 0))
	page := display.New(pb)
	bmp := bitmap.New(ib)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			b := w*y + x
			if b&0x01 != 0 {
				bmp.Set(x*2, y*4, true)
			}
			if b&0x02 != 0 {
				bmp.Set(x*2+1, y*4, true)
			}
			if b&0x04 != 0 {
				bmp.Set(x*2, y*4+1, true)
			}
			if b&0x08 != 0 {
				bmp.Set(x*2+1, y*4+1, true)
			}

			if b&0x10 != 0 {
				bmp.Set(x*2, y*4+2, true)
			}
			if b&0x20 != 0 {
				bmp.Set(x*2+1, y*4+2, true)
			}
			if b&0x40 != 0 {
				bmp.Set(x*2, y*4+3, true)
			}
			if b&0x80 != 0 {
				bmp.Set(x*2+1, y*4+3, true)
			}
		}
	}

	page.Fill(image.Rect(0, 0, 1, h), string(0x28ff), display.Colors[15], color.Transparent)
	braille.DrawBitmap(page, rb, bmp, image.ZP, image.ZP, display.Colors[7])

	var buf []byte
	cur := display.Reset
	buf, cur = display.Render(buf, cur, page, display.Model8)
	buf = append(buf, "\r\n"...)
	_, err = os.Stdout.Write(buf)

	return err
}
