package point

import (
	"image"
)

// ZFrame represents a frame of reference for computing Z-Curve keys.
type ZFrame struct {
	Bounds image.Rectangle
}

// Key packs a point into a z-curve key; if the point is outside the bounding
// box, then Key(bounds.Max) is returned.
func (zf ZFrame) Key(pt image.Point) (z uint64) {
	if !pt.In(zf.Bounds) {
		pt = zf.Bounds.Max
	}
	pt = pt.Sub(zf.Bounds.Min)
	// TODO: evaluate a table ala
	// https://graphics.stanford.edu/~seander/bithacks.html#InterleaveTableObvious
	x, y := uint32(pt.X), uint32(pt.Y)
	for i := uint(0); i < 32; i++ {
		z |= uint64(x&(1<<i)) << i
		z |= uint64(y&(1<<i)) << (i + 1)
	}
	return z
}

// Point unpacks a z-curve key into a point.
func (zf ZFrame) Point(z uint64) image.Point {
	var x, y uint64
	for i := uint(0); i < 32; i++ {
		j := 2 * i
		x |= (z & (1 << j)) >> i
		y |= (z & (1 << (j + 1))) >> (i + 1)
	}
	return image.Pt(int(x), int(y)).Add(zf.Bounds.Min)
}
