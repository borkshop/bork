package display_test

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"testing"

	. "github.com/borkshop/bork/internal/cops/display"
	"github.com/stretchr/testify/assert"
)

func TestDraw_centered(t *testing.T) {
	dst := New(image.Rect(0, 0, 16, 8))
	src := New(image.Rect(0, 0, 8, 4))
	dst.Fill(dst.Rect, "_", Colors[7], Colors[0])
	src.Fill(src.Rect, "x", Colors[5], Colors[4])

	bound, off := dst.Rect, image.ZP
	if n := bound.Dx() - src.Rect.Dx(); n > 0 {
		bound.Min.X += n / 2
	} else if n < 0 {
		off.X += n / 2
	}
	if n := bound.Dy() - src.Rect.Dy(); n > 0 {
		bound.Min.Y += n / 2
	} else if n < 0 {
		off.Y += n / 2
	}
	Draw(dst, bound, src, off, draw.Over)

	assert.Equal(t, []string{
		"________________",
		"________________",
		"____xxxxxxxx____",
		"____xxxxxxxx____",
		"____xxxxxxxx____",
		"____xxxxxxxx____",
		"________________",
		"________________",
	}, dst.Text.Lines("0"))

	assert.Equal(t, []string{
		"................",
		"................",
		"....!!!!!!!!....",
		"....!!!!!!!!....",
		"....!!!!!!!!....",
		"....!!!!!!!!....",
		"................",
		"................",
	}, imageTestRepr(dst.Foreground, "0", map[color.RGBA]string{
		Colors[7]: ".",
		Colors[5]: "!",
	}))

	assert.Equal(t, []string{
		"----------------",
		"----------------",
		"----########----",
		"----########----",
		"----########----",
		"----########----",
		"----------------",
		"----------------",
	}, imageTestRepr(dst.Background, "0", map[color.RGBA]string{
		Colors[0]: "-",
		Colors[4]: "#",
	}))
}

func imageTestRepr(img *image.RGBA, dflt string, c2s map[color.RGBA]string) (r []string) {
	r = make([]string, img.Rect.Dy())
	nx := img.Rect.Dx()
	img.Bounds()
	var buf bytes.Buffer
	for y := 0; y < len(r); y++ {
		buf.Reset()
		for x := 0; x < nx; x++ {
			c := img.RGBAAt(x, y)
			s := c2s[c]
			if s == "" {
				s = dflt
			}
			_, _ = buf.WriteString(s)
		}
		r[y] = buf.String()
	}
	return r
}
