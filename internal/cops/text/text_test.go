package text_test

import (
	"image"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borkshop/bork/internal/cops/display"
	. "github.com/borkshop/bork/internal/cops/text"
	"github.com/borkshop/bork/internal/rectangle"
)

func TestBounds(t *testing.T) {
	assert.Equal(t, image.Rect(0, 0, 0, 0), Bounds(""))
	assert.Equal(t, image.Rect(0, 0, 1, 1), Bounds("1"))
	assert.Equal(t, image.Rect(0, 0, 3, 2), Bounds("abc\n12"))
	assert.Equal(t, image.Rect(0, 0, 3, 2), Bounds("ab\n123"))
	assert.Equal(t, image.Rect(0, 0, 3, 2), Bounds("abc\n123\n"))
}

func TestRender(t *testing.T) {
	str := "abc\n123\n"
	bounds := Bounds(str)
	front := display.New(bounds)
	back := display.New(bounds)
	front.Fill(front.Bounds(), "", display.Colors[7], display.Colors[0])
	Write(front, bounds, str, display.Colors[7])
	var buf []byte
	cur := display.Reset
	buf, cur = display.RenderOver(buf, cur, front, back, display.Model0)
	assert.Equal(t, "abc\r\n123", string(buf), "renders two line string")
}

func TestOffset(t *testing.T) {
	str := "abc"
	bounds := Bounds(str).Add(image.Pt(2, 1))
	outset := rectangle.Outset(bounds, 2, 1)
	front, back := display.New2(outset)
	front.Fill(outset, ".", display.Colors[7], display.Colors[0])
	assert.Equal(t, ".", front.Text.At(0, 0))
	Write(front, bounds, str, display.Colors[7])
	var buf []byte
	cur := display.Reset
	buf, cur = display.RenderOver(buf, cur, front, back, display.Model0)
	assert.Equal(t, ".......\r\n..abc..\r\n.......", string(buf))
}
