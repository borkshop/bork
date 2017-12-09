package display

import (
	"image/color"
	"io"
)

const (
	bufferSize     = 1024
	flushThreshold = 512
)

func NewRenderer(writer io.Writer, mod Model) *Renderer {
	return &Renderer{
		writer:         writer,
		flushThreshold: flushThreshold,
		mod:            mod,
	}
}

type Renderer struct {
	writer         io.Writer
	flushThreshold int
	mod            Model
}

func (r *Renderer) RenderOver(buf []byte, cur Cursor, over, under *Display) ([]byte, Cursor) {
	// TODO choose: A.) make model fully concrete and public B.) restore model
	// interface C.) switch on model type for conditional optimization, like
	// draw package.
	mod := r.mod.(model)

	rect := over.Rect.Intersect(under.Rect)
	pt := rect.Min
	i := over.Text.StringsOffset(pt.X, pt.Y)
	j := 0
	if under != nil {
		j = under.Text.StringsOffset(pt.X, pt.Y)
	}
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
			buf, cur = cur.Go(buf, pt)
			buf, cur = mod.RenderRGBA(buf, cur, of, ob)
			buf, cur = cur.WriteGlyph(buf, ot)
			if under != nil {
				under.setrgbai(j, ot, of, ob)
			}
		}
		pt.X++
		if pt.X >= rect.Max.X {
			pt.X = rect.Min.X
			pt.Y++
		}
		if pt.Y >= rect.Max.Y {
			break
		}
		i++
		j++

		r.writer.Write(buf)
		buf = buf[0:0]
	}
	buf, cur = cur.Reset(buf)

	r.writer.Write(buf)
	buf = buf[0:0]
	return buf, cur
}
