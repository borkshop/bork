package hilbert

import "image"

// Scale is the width and height of a Hilbert space-filling curve.
// The scale must be a power of two.
type Scale int

// Encode returns the linear index of a point along a Hilbert curve with this
// scale.
func (scale Scale) Encode(pt image.Point) int {
	return Encode(pt, int(scale))
}

// Decode returns the point at the given offset of a Hilbert curve with this
// scale.
func (scale Scale) Decode(hi int) image.Point {
	return Decode(hi, int(scale))
}

// Encode returns the linear index of a point along a Hilbert curve for a given
// scale.
func Encode(pt image.Point, scale int) int {
	var rotation image.Point
	h := 0
	for s := scale >> 1; s > 0; s >>= 1 {
		rotation.X = pt.X & s
		rotation.Y = pt.Y & s
		h += s * ((3 * rotation.X) ^ rotation.Y)
		pt = rotate(pt, s, rotation)
	}
	return h
}

// Decode returns the point corresponding to a given position along a Hilbert
// curve of particular size.
func Decode(h int, scale int) image.Point {
	var pt, rotation image.Point
	for s := 1; s < scale; s <<= 1 {
		rotation.X = 1 & h >> 1
		rotation.Y = 1 & (h ^ rotation.X)
		pt = rotate(pt, scale, rotation)
		rotation = rotation.Mul(s)
		pt = pt.Add(rotation)
		h >>= 2
	}
	return pt
}

func rotate(pt image.Point, scale int, rotation image.Point) image.Point {
	if rotation.Y == 0 {
		if rotation.X != 0 {
			pt.X = scale - 1 - pt.X
			pt.Y = scale - 1 - pt.Y
		}
		pt.X, pt.Y = pt.Y, pt.X
	}
	return pt
}
