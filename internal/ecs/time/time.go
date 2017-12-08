package time

import (
	"fmt"
	"math"
)

// Time represents Proccess() time in an ecs.System; it counts number of
// processing ticks.
type Time uint64

// Duration represents ecs.System time spans; it counts a number of Process()
// ticks.
type Duration int64

// String the time, either as "tNNN" or "EoT if maxed out.
func (t Time) String() string {
	if t == math.MaxUint64 {
		return "EoT"
	}
	return fmt.Sprintf("t%d", uint64(t))
}

func (d Duration) String() string { return fmt.Sprintf("%d ticks", int64(d)) }

// Sub tract another Time, returning the difference as a Duration.
func (t Time) Sub(ot Time) Duration {
	// TODO probably wrong for large times (over 2^63)
	return Duration(t - ot)
}

// Add a Duration, returning a Time; result is clamped to Time(0) and
// Time(math.MaxUint64).
func (t Time) Add(d Duration) Time {
	if d < 0 && Time(-d) > t {
		return 0
	}
	if d > 0 && math.MaxUint64-Time(d) < t {
		return math.MaxUint64
	}
	return t + Time(d)
}
