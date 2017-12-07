package moremath

import "math"

// Round ala https://www.cockroachlabs.com/blog/rounding-implementations-in-go/
//
// NOTE: obsolete in Go 1.10 when this finally comes into the "math" package.
func Round(x float64) float64 {
	const (
		mask  = 0x7FF
		shift = 64 - 11 - 1
		bias  = 1023

		signMask = 1 << 63
		fracMask = (1 << shift) - 1
		halfMask = 1 << (shift - 1)
		one      = bias << shift
	)

	bits := math.Float64bits(x)
	e := uint(bits>>shift) & mask
	switch {
	case e < bias:
		// Round abs(x)<1 including denormals.
		bits &= signMask // +-0
		if e == bias-1 {
			bits |= one // +-1
		}
	case e < bias+shift:
		// Round any abs(x)>=1 containing a fractional component [0,1).
		e -= bias
		bits += halfMask >> e
		bits &^= fracMask >> e
	}
	return math.Float64frombits(bits)
}
