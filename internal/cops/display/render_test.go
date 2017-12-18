package display_test

import (
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

func TestRender_model8(t *testing.T) {
	dis := New(image.Rect(0, 0, 16, 8))
	for y := 0; y < 8; y++ {
		b := Colors[y]
		for x := 0; x < 16; x++ {
			f := Colors[x/2]
			dis.SetRGBA(x, y, "<", f, b)
			x++
			dis.SetRGBA(x, y, ">", f, b)
		}
	}
	// TODO also cur = Start
	buf, _ := Render(nil, Reset, dis, Model3)
	assert.Equal(t, []string{
		"[30m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
		"[30m[41m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
		"[30m[42m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
		"[30m[43m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
		"[30m[44m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
		"[30m[45m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
		"[30m[46m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>",
		"[30m[47m<>[31m<>[32m<>[33m<>[34m<>[35m<>[36m<>[37m<>[m",
	}, strings.Split(string(buf), "\r\n"))
}
