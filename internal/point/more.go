package point

import "image"

// MulRespective returns a point scaled by another point, multiplying the
// values of their respective axes.
func MulRespective(a, b image.Point) image.Point {
	return image.Pt(a.X*b.X, a.Y*b.Y)
}
