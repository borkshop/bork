package display_test

import (
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"

	. "github.com/borkshop/bork/internal/cops/display"
)

func TestRenderMultiRuneCell(t *testing.T) {
	whiteHand := "👍🏻"
	d := New(image.Rect(0, 0, 2, 1))
	d.Set(0, 0, whiteHand, color.White, color.Transparent)
	d.Set(1, 0, whiteHand, color.White, color.Transparent)
	cur := Reset
	var buf []byte
	buf, cur = Render(buf, cur, d, Model0)
	assert.Equal(t, []byte(whiteHand+"\r\033[1C"+whiteHand+"\033[m"), buf)
}

func TestRenderBlankAndMultiRuneCell(t *testing.T) {
	whiteHand := "👍🏻"
	d := New(image.Rect(0, 0, 3, 1))
	d.Set(0, 0, "", color.White, color.Transparent)
	d.Set(1, 0, whiteHand, color.White, color.Transparent)
	d.Set(2, 0, "", color.White, color.Transparent)
	cur := Reset
	var buf []byte
	buf, cur = Render(buf, cur, d, Model0)
	assert.Equal(t, []byte(" "+whiteHand+"\r\033[2C \033[m"), buf)
}

func TestRenderBlankAndMultiRuneCellOver(t *testing.T) {
	whiteHand := "👍🏻"
	front, back := New2(image.Rect(0, 0, 3, 1))
	front.Set(0, 0, "", color.White, color.Transparent)
	front.Set(1, 0, whiteHand, color.White, color.Transparent)
	front.Set(2, 0, "", color.White, color.Transparent)
	cur := Reset
	var buf []byte
	buf, cur = RenderOver(buf, cur, front, back, Model0)
	assert.Equal(t, []byte(" "+whiteHand+"\r\033[2C \033[m"), buf)
}
