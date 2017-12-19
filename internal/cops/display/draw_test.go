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
	for _, tc := range []struct {
		name     string
		dstSize  image.Point
		srcSize  image.Point
		expected []string
	}{
		{
			name:    "src < dst",
			dstSize: image.Pt(16, 8),
			srcSize: image.Pt(8, 4),
			expected: []string{
				"________________",
				"________________",
				"____xxxxxxxx____",
				"____xxxxxxxx____",
				"____xxxxxxxx____",
				"____xxxxxxxx____",
				"________________",
				"________________",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dst := New(image.Rectangle{Max: tc.dstSize})
			src := New(image.Rectangle{Max: tc.srcSize})
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

			assert.Equal(t, tc.expected, dst.Text.LinesWithFill("0"), "expected textile draw")
			assert.Equal(t, tc.expected, imageTestRepr(dst.Foreground, "0", map[color.RGBA]string{
				Colors[7]: "_",
				Colors[5]: "x",
			}), "expected foreground draw")
			assert.Equal(t, tc.expected, imageTestRepr(dst.Background, "0", map[color.RGBA]string{
				Colors[0]: "_",
				Colors[4]: "x",
			}), "expected background draw")
		})
	}
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
