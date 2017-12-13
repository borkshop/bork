package moremath

// MinInt returns the smallest int from its arguments; panics if called with no
// args.
func MinInt(ints ...int) int {
	min := ints[0]
	for i := 1; i < len(ints); i++ {
		if n := ints[i]; n < min {
			min = n
		}
	}
	return min
}

// MaxInt returns the largest int from its arguments; panics if called with no
// args.
func MaxInt(ints ...int) int {
	max := ints[0]
	for i := 1; i < len(ints); i++ {
		if n := ints[i]; n > max {
			max = n
		}
	}
	return max
}

// IntSign returns -1, 1, or 0 if n is less than, greater
// than, or equal to 0 respectively.
func IntSign(n int) int {
	if n < 0 {
		return -1
	}
	if n > 0 {
		return 1
	}
	return 0
}
