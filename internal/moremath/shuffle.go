package moremath

import "math"

// Shuffle interleaves two 32 bit integers into a single 64-bit one; can be
// used to map 2d points to 1d on a z-order curve.
func Shuffle(x, y uint32) (z uint64) {
	// TODO: evaluate a table ala
	// https://graphics.stanford.edu/~seander/bithacks.html
	for i := uint(0); i < 32; i++ {
		z |= (uint64(x)&(1<<i))<<i | (uint64(y)&(1<<i))<<(i+1)
	}
	return z
}

// ClampInt32 maps an int into a uint32 ranging from min-int32 to max-int32;
// anything outside that range is clamped to the extreme it surpassed.
func ClampInt32(n int) uint32 {
	if n < math.MinInt32 {
		return 0
	}
	if n > math.MaxInt32 {
		return math.MaxUint32
	}
	return uint32(n - math.MinInt32)
}
