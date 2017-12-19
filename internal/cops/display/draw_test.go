package display_test

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"strconv"
	"strings"
	"testing"

	. "github.com/borkshop/bork/internal/cops/display"
	"github.com/stretchr/testify/assert"
)

func TestDraw_centered(t *testing.T) {
	for _, tc := range []struct {
		name                            string
		dstSize                         image.Point
		srcSize                         image.Point
		expectedT, expectedF, expectedB []string
	}{
		{
			name:    "src < dst",
			dstSize: image.Pt(16, 8),
			srcSize: image.Pt(8, 4),
			expectedT: []string{
				"________________",
				"________________",
				"____0x1x2x3x____",
				"____4x5x6x7x____",
				"____8x9xAxBx____",
				"____CxDxExFx____",
				"________________",
				"________________",
			},
			expectedF: []string{
				"________________",
				"________________",
				"____1x2x3x4x____",
				"____1x2x3x4x____",
				"____1x2x3x4x____",
				"____1x2x3x4x____",
				"________________",
				"________________",
			},
			expectedB: []string{
				"________________",
				"________________",
				"____1x1x1x1x____",
				"____2x2x2x2x____",
				"____3x3x3x3x____",
				"____4x4x4x4x____",
				"________________",
				"________________",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dst := New(image.Rectangle{Max: tc.dstSize})
			src := New(image.Rectangle{Max: tc.srcSize})
			dst.Fill(dst.Rect, "_", Colors[0], Colors[1])
			src.Fill(src.Rect, "x", Colors[2], Colors[3])
			for y, dy := 0, tc.srcSize.Y/4; y < tc.srcSize.Y; y += dy {
				yi := y / dy
				for x, dx := 0, tc.srcSize.X/4; x < tc.srcSize.X; x += dx {
					xi := x / dx
					t := strings.ToUpper(strconv.FormatInt(int64(4*yi+xi), 16))
					src.Set(x, y, t, Colors[4+xi], Colors[4+yi])
				}
			}

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

			assert.Equal(t, tc.expectedT, dst.Text.LinesWithFill("0"), "expected textile draw")
			assert.Equal(t, tc.expectedF, imageTestRepr(dst.Foreground, "0", map[color.RGBA]string{
				Colors[0]: "_",
				Colors[2]: "x",
				Colors[4]: "1",
				Colors[5]: "2",
				Colors[6]: "3",
				Colors[7]: "4",
			}), "expected foreground draw")
			assert.Equal(t, tc.expectedB, imageTestRepr(dst.Background, "0", map[color.RGBA]string{
				Colors[1]: "_",
				Colors[3]: "x",
				Colors[4]: "1",
				Colors[5]: "2",
				Colors[6]: "3",
				Colors[7]: "4",
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
