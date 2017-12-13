package display

import "image/color"

// Render appends ANSI escape sequences to a byte slice to overwrite an entire
// terminal window, using the best matching colors in the terminal color model.
func Render(buf []byte, cur Cursor, over *Display, renderColor ColorModel) ([]byte, Cursor) {
	return RenderOver(buf, cur, over, nil, renderColor)
}

// RenderOver appends ANSI escape sequences to a byte slice to update a
// terminal display to look like the front model, skipping cells that are the
// same in the back model, using escape sequences and the nearest matching
// colors in the given color model.
func RenderOver(buf []byte, cur Cursor, over, under *Display, renderColor ColorModel) ([]byte, Cursor) {
	vp := over.Rect
	if under != nil {
		vp = over.Rect.Intersect(under.Rect)
	}
	pt := vp.Min
	i := over.Text.StringsOffset(pt.X, pt.Y)
	j := 0
	if under != nil {
		j = under.Text.StringsOffset(pt.X, pt.Y)
	}
	buf, cur = cur.Go(buf, pt)
	for i < len(over.Text.Strings) {
		var ut string
		var uf, ub color.RGBA
		ot, of, ob := over.rgbaati(i)
		if under != nil {
			ut, uf, ub = under.rgbaati(j)
		}
		if len(ot) == 0 {
			ot = " "
		}
		if len(ut) == 0 {
			ut = " "
		}
		if ot != ut || of != uf || ob != ub {
			if dy := pt.Y - cur.Position.Y; dy > 0 {
				buf, cur = cur.linedown(buf, dy)
			}
			if cur.Position.X < 0 {
				buf = append(buf, "\r"...)
				cur.Position.X = 0
				buf, cur = cur.right(buf, pt.X)
			} else if dx := pt.X - cur.Position.X; dx > 0 {
				buf, cur = cur.right(buf, dx)
			}
			buf, cur = renderColor(buf, cur, of, ob)
			buf, cur = cur.WriteGlyph(buf, ot)
			if under != nil {
				under.setrgbai(j, ot, of, ob)
			}
		}
		pt.X++
		if pt.X >= vp.Max.X {
			pt.X = vp.Min.X
			pt.Y++
		}
		if pt.Y >= vp.Max.Y {
			break
		}
		i++
		j++
	}
	buf, cur = cur.Reset(buf)
	return buf, cur
}
