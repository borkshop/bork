package display_test

import (
	"fmt"
	"image"
	"image/color"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/borkshop/bork/internal/cops/display"
)

func TestRender_multiRuneCell(t *testing.T) {
	whiteHand := "👍🏻"
	d := New(image.Rect(0, 0, 2, 1))
	d.Set(0, 0, whiteHand, color.White, color.Transparent)
	d.Set(1, 0, whiteHand, color.White, color.Transparent)
	cur := Reset
	var buf []byte
	buf, cur = Render(buf, cur, d, Model0)
	assert.Equal(t, []byte(whiteHand+"\r\033[1C"+whiteHand), buf)
}

func TestRender_blankAndMultiRuneCell(t *testing.T) {
	whiteHand := "👍🏻"
	d := New(image.Rect(0, 0, 3, 1))
	d.Set(0, 0, "", color.White, color.Transparent)
	d.Set(1, 0, whiteHand, color.White, color.Transparent)
	d.Set(2, 0, "", color.White, color.Transparent)
	cur := Reset
	var buf []byte
	buf, cur = Render(buf, cur, d, Model0)
	assert.Equal(t, []byte(" "+whiteHand+"\r\033[2C "), buf)
}

func TestRender_blankAndMultiRuneCellOver(t *testing.T) {
	whiteHand := "👍🏻"
	front, back := New2(image.Rect(0, 0, 3, 1))
	front.Set(0, 0, "", color.White, color.Transparent)
	front.Set(1, 0, whiteHand, color.White, color.Transparent)
	front.Set(2, 0, "", color.White, color.Transparent)
	cur := Reset
	var buf []byte
	buf, cur = RenderOver(buf, cur, front, back, Model0)
	assert.Equal(t, []byte(" "+whiteHand+"\r\033[2C "), buf)
}

func TestRender_integration(t *testing.T) {
	for _, tc := range []struct {
		name     string
		size     image.Point
		model    ColorModel
		setup    func(x, y int) (t string, f, b color.RGBA)
		expected []string
	}{
		{
			name:  "model3 coverage",
			size:  image.Pt(16, 8),
			model: Model3,
			setup: func(x, y int) (t string, f, b color.RGBA) {
				if x%2 == 0 {
					t = "<"
				} else {
					t = ">"
				}
				return t, Colors[(x/2)%8], Colors[y%8]
			},
			expected: []string{
				"[30m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
				"[30m[41m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
				"[30m[42m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
				"[30m[43m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
				"[30m[44m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
				"[30m[45m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
				"[30m[46m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
				"[30m[47m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>[m",
			},
		},
	} {
		// TODO also cur = Start

		t.Run(fmt.Sprintf("%s Set", tc.name), func(t *testing.T) {
			dis := New(image.Rect(0, 0, tc.size.X, tc.size.Y))
			for y := 0; y < 8; y++ {
				for x := 0; x < 16; x++ {
					t, f, b := tc.setup(x, y)
					dis.Set(x, y, t, f, b)
				}
			}
			buf, _ := Render(nil, Reset, dis, tc.model)
			assert.Equal(t, tc.expected, strings.Split(string(buf), "\r\n"))
		})

		t.Run(fmt.Sprintf("%s SetRGBA", tc.name), func(t *testing.T) {
			dis := New(image.Rect(0, 0, tc.size.X, tc.size.Y))
			for y := 0; y < 8; y++ {
				for x := 0; x < 16; x++ {
					t, f, b := tc.setup(x, y)
					dis.SetRGBA(x, y, t, f, b)
				}
			}
			buf, _ := Render(nil, Reset, dis, tc.model)
			assert.Equal(t, tc.expected, strings.Split(string(buf), "\r\n"))
		})

	}

}
