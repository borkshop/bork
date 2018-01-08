package point

import (
	"image"
	"math"
)

// ZKey computes the z-curve key for the given point.
func ZKey(pt image.Point) (z uint64) {
	// TODO: evaluate a table ala
	// https://graphics.stanford.edu/~seander/bithacks.html#InterleaveTableObvious
	x, y := truncInt32(pt.X), truncInt32(pt.Y)
	for i := uint(0); i < 32; i++ {
		z |= (x&(1<<i))<<i | (y&(1<<i))<<(i+1)
	}
	return z
}

func truncInt32(n int) uint64 {
	if n < math.MinInt32 {
		return math.MinInt32
	}
	if n > math.MaxInt32 {
		return math.MaxUint32
	}
	return uint64(uint32(n - math.MinInt32))
}
