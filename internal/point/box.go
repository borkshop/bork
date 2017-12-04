package point

// Bx is a convenience constructor for Box.
func Bx(tlx, tly int, brx, bry int) Box {
	return Box{Point{tlx, tly}, Point{brx, bry}}
}

// Box represents a bounding box defined by a top-left and bottom-right point.
type Box struct {
	TopLeft     Point
	BottomRight Point
}

// Size returns the width and height of the box as a point.
func (b Box) Size() Point {
	return b.BottomRight.Sub(b.TopLeft).Abs()
}

// ExpandTo expands a copy of the box to include the given point, returning the
// copy.
func (b Box) ExpandTo(pt Point) Box {
	if pt.X < b.TopLeft.X {
		b.TopLeft.X = pt.X
	}
	if pt.Y < b.TopLeft.Y {
		b.TopLeft.Y = pt.Y
	}
	if pt.X >= b.BottomRight.X {
		b.BottomRight.X = pt.X + 1
	}
	if pt.Y >= b.BottomRight.Y {
		b.BottomRight.Y = pt.Y + 1
	}
	return b
}

// ExpandBy symmetrically expands a copy of the box by a given x/y
// displacement, returning the copy.
func (b Box) ExpandBy(d Point) Box {
	b.TopLeft = b.TopLeft.Sub(d)
	b.BottomRight = b.BottomRight.Add(d)
	return b
}

// Contains returns true if a given point is inside the box.
func (b Box) Contains(pt Point) bool {
	return !(pt.Less(b.TopLeft) || b.BottomRight.Less(pt))
}

// DistanceTo returns a signed distance to the given point
// from the nearest box corner.
func (b Box) DistanceTo(pt Point) Point {
	if pt.Less(b.TopLeft) {
		return pt.Sub(b.TopLeft)
	}
	if b.BottomRight.Less(pt) {
		return pt.Sub(b.BottomRight)
	}
	return Zero
}

// Min returns a copy of the box each corner min'd to the given point.
func (b Box) Min(pt Point) Box {
	b.TopLeft = b.TopLeft.Min(pt)
	b.BottomRight = b.BottomRight.Min(pt)
	return b
}

// Max returns a copy of the box each corner max'd to the given point.
func (b Box) Max(pt Point) Box {
	b.TopLeft = b.TopLeft.Max(pt)
	b.BottomRight = b.BottomRight.Max(pt)
	return b
}

// Add returns a copy of the box with the given point added to the corners.
func (b Box) Add(pt Point) Box {
	b.TopLeft = b.TopLeft.Add(pt)
	b.BottomRight = b.BottomRight.Add(pt)
	return b
}

// Sub returns a copy of the box with the given point subtracted from the
// corners.
func (b Box) Sub(pt Point) Box {
	b.TopLeft = b.TopLeft.Sub(pt)
	b.BottomRight = b.BottomRight.Sub(pt)
	return b
}
