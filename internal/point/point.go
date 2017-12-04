package point

// Pt is a convenience constructor for Point.
func Pt(x, y int) Point { return Point{x, y} }

// Point represents a point in <X,Y> 2-space.
type Point struct{ X, Y int }

// Zero is the origin, the zero value of Point.
var Zero = Point{}

// Less returns true if this point's X or Y component is less than the other's.
func (pt Point) Less(other Point) bool {
	return pt.Y < other.Y || pt.X < other.X
}

// Equal returns true if both this point's X and Y components equal another's.
func (pt Point) Equal(other Point) bool {
	return pt.X == other.X && pt.Y == other.Y
}

// Min returns a copy of this point with each component the
// minimum of the two points' components.
func (pt Point) Min(other Point) Point {
	if other.X < pt.X {
		pt.X = other.X
	}
	if other.Y < pt.Y {
		pt.Y = other.Y
	}
	return pt
}

// Max returns a copy of this point with each component the
// maximum of the two points' components.
func (pt Point) Max(other Point) Point {
	if other.X > pt.X {
		pt.X = other.X
	}
	if other.Y > pt.Y {
		pt.Y = other.Y
	}
	return pt
}

// Add adds another point's values to a copy of this point, returning the copy.
func (pt Point) Add(other Point) Point {
	pt.X += other.X
	pt.Y += other.Y
	return pt
}

// Sub subtracts another point's values from a copy of this point, returning
// the copy.
func (pt Point) Sub(other Point) Point {
	pt.X -= other.X
	pt.Y -= other.Y
	return pt
}

// Div divides a copy of this point's values by a constant, returning the copy.
func (pt Point) Div(n int) Point {
	pt.X /= n
	pt.Y /= n
	return pt
}

// Mul multiplies a copy of this point's values by a constant, returning the
// copy.
func (pt Point) Mul(n int) Point {
	pt.X *= n
	pt.Y *= n
	return pt
}

// Abs returns a copy of this point with its values non-negative.
func (pt Point) Abs() Point {
	if pt.X < 0 {
		pt.X = -pt.X
	}
	if pt.Y < 0 {
		pt.Y = -pt.Y
	}
	return pt
}

// Neg negates a copy of this point, returning the copy.
func (pt Point) Neg() Point {
	pt.X = -pt.X
	pt.Y = -pt.Y
	return pt
}

// Sign returns a copy of this point reduced to the values -1, 0, or 1 depending
// on the sign of the original values.
func (pt Point) Sign() Point {
	pt.X = sign(pt.X)
	pt.Y = sign(pt.Y)
	return pt
}

// Dot return the dot product of this point with another.
func (pt Point) Dot(other Point) int {
	return pt.X*other.X + pt.Y*other.Y
}

// SumSQ returns the sum-of-squared components.
func (pt Point) SumSQ() int {
	return pt.X*pt.X + pt.Y*pt.Y
}

func sign(i int) int {
	if i < 0 {
		return -1
	}
	if i > 0 {
		return 1
	}
	return 0
}
