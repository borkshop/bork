package point_test

import (
	"image"
	"testing"

	. "github.com/borkshop/bork/internal/point"
	"github.com/stretchr/testify/assert"
)

func TestPoint_compat(t *testing.T) {
	pt := Pt(2, 5)
	ipt := image.Pt(2, 5)
	assert.Equal(t, ipt, image.Point(pt))
	assert.Equal(t, pt, Point(ipt))
}

// func TestBox_cast(t *testing.T) {
// 	box := Bx(3, 5, 8, 13)
// 	rect := image.Rect(3, 5, 8, 13)
// 	assert.Equal(t, box, image.Rectangle(rect))
// 	assert.Equal(t, rect, Box(rect))
// }
