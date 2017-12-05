package view

import (
	"fmt"

	termbox "github.com/nsf/termbox-go"

	"github.com/borkshop/bork/internal/moremath"
	"github.com/borkshop/bork/internal/point"
)

// Layout places Renderables in a Grid, keeping track of used left/right/center
// space to inform future placements.
type Layout struct {
	Grid

	// invariant: avail[i] == Grid.Size.X - lused[i] - rused[i]
	lused []int
	rused []int
	cused []int
	avail []int
}

// Align specifies alignment to Layout placements.
type Align uint8

const (
	// AlignLeft causes left horizontal alignment in a Layout.
	AlignLeft Align = 1 << iota
	// AlignRight causes right horizontal alignment in a Layout.
	AlignRight

	// AlignTop causes top vertical alignment in a Layout.
	AlignTop
	// AlignBottom causes bottom vertical alignment in a Layout.
	AlignBottom

	// AlignHFlush causes horizontal alignment to accept no offset; so it will
	// always get the "next empty row" in the relevant vertical direction.
	AlignHFlush

	// AlignCenter causes center horizontal alignment in a layout.
	AlignCenter = AlignLeft | AlignRight

	// AlignMiddle causes middle vertical alignment in a layout.
	AlignMiddle = AlignTop | AlignBottom
)

func (a Align) String() string {
	parts := make([]string, 0, 3)

	if a&AlignHFlush != 0 {
		parts = append(parts, "flush")
	}
	switch a & AlignCenter {
	case AlignLeft:
		parts = append(parts, "left")
	case AlignRight:
		parts = append(parts, "right")
	case AlignCenter:
		parts = append(parts, "center")
	default:
		parts = append(parts, "default")
	}

	switch a & AlignMiddle {
	case AlignTop:
		parts = append(parts, "top")
	case AlignBottom:
		parts = append(parts, "bottom")
	case AlignMiddle:
		parts = append(parts, "middle")
	default:
		parts = append(parts, "default")
	}

	return fmt.Sprintf("Align%s", parts)
}

// Renderable is an element for Layout to place and maybe render; if its Render
// method is called, it will get a grid of at least the needed RenderSize.
type Renderable interface {
	RenderSize() (wanted, needed point.Point)
	Render(Grid)
}

func (lay *Layout) init() {
	n := lay.Grid.Size.Y
	if cap(lay.avail) < n {
		lay.lused = make([]int, n)
		lay.rused = make([]int, n)
		lay.cused = make([]int, n)
		lay.avail = make([]int, n)
	} else {
		lay.lused = lay.lused[:n]
		lay.rused = lay.rused[:n]
		lay.cused = lay.cused[:n]
		lay.avail = lay.avail[:n]
	}
	n = lay.Grid.Size.X
	for i := range lay.avail {
		lay.avail[i] = n
	}
}

// LayoutPlacement represents a placement made by a Layout for a Renderable.
type LayoutPlacement struct {
	lay *Layout

	ren    Renderable
	align  Align
	wanted point.Point
	needed point.Point
	sep    termbox.Cell // TODO: give user an option

	ok    bool
	start int
	have  point.Point
}

// Place a Renderable into layout, returning false if the placement can't be
// done.
func (lay *Layout) Place(ren Renderable, align Align) LayoutPlacement {
	if len(lay.avail) != lay.Grid.Size.Y {
		lay.init()
	}
	plc := MakeLayoutPlacement(lay, ren)
	plc.Try(align)
	return plc
}

// Render places and renders a Renderable if the placement succeeded.
func (lay *Layout) Render(ren Renderable, align Align) LayoutPlacement {
	plc := lay.Place(ren, align)
	plc.Render()
	return plc
}

// MakeLayoutPlacement makes a new placement for the given layout and
// renderable; it records the wanted/needed render sizes, ready to attempt
// placement.
func MakeLayoutPlacement(lay *Layout, ren Renderable) LayoutPlacement {
	plc := LayoutPlacement{
		lay: lay,
		ren: ren,
	}
	plc.wanted, plc.needed = ren.RenderSize()
	plc.setSep(' ')
	return plc
}

func (plc *LayoutPlacement) setSep(ch rune) {
	plc.sep = termbox.Cell{Ch: ch}
}

// Try attempts to (re)resolve the placement with an other alignment.
func (plc *LayoutPlacement) Try(align Align) bool {
	if plc.wanted.X == 0 || plc.wanted.Y == 0 {
		plc.ok = false
		return false
	}

	// h-flush should default to left-align, not center
	if align&AlignCenter == 0 && align&AlignHFlush != 0 {
		align |= AlignLeft
	}
	plc.align = align

	switch align & AlignMiddle {
	case AlignTop:
		plc.find(0, 1)

	case AlignBottom:
		plc.find(len(plc.lay.avail)-1, -1)

	default: // NOTE: defaults to AlignMiddle:
		mid := len(plc.lay.avail) / 2
		plc.find(mid, 1)
		if !plc.ok {
			plc.find(mid, -1)
		} else {
			alt := *plc
			alt.find(mid, -1)
			if alt.ok {
				if ld, ud := mid-plc.start, alt.start-mid; ud < ld {
					*plc = alt
				}
			}
		}
	}

	return plc.ok
}

func (plc *LayoutPlacement) find(init, dir int) {
	var (
		left   = plc.align&AlignCenter == AlignLeft
		right  = plc.align&AlignCenter == AlignRight
		center = plc.align&AlignCenter == AlignCenter
		lflush = plc.align&AlignHFlush != 0 && left
		rflush = plc.align&AlignHFlush != 0 && right
	)

	plc.ok = false
	plc.start = init
seekStart:
	needed := 0
	plc.have = point.Zero
	for plc.start >= 0 && plc.start < len(plc.lay.avail) {
		needed = plc.needed.X
		if plc.sep.Ch != 0 && ((left && plc.lay.lused[plc.start] > 0) ||
			(right && plc.lay.rused[plc.start] > 0)) {
			needed++
		}
		if plc.lay.avail[plc.start] >= needed &&
			!(center && plc.lay.cused[plc.start] > 0) &&
			!(lflush && plc.lay.lused[plc.start] > 0) &&
			!(rflush && plc.lay.rused[plc.start] > 0) {
			plc.have.X = moremath.MinInt(plc.wanted.X, plc.lay.avail[plc.start])
			goto seekEnd
		}
		plc.start += dir
	}
	return

seekEnd:
	end := plc.start + dir
	plc.have.Y++
	for end >= 0 && end < len(plc.lay.avail) {
		if plc.have.Y >= plc.wanted.Y {
			break
		}
		if plc.lay.avail[end] < needed ||
			(center && plc.lay.cused[end] > 0) ||
			(lflush && plc.lay.lused[end] > 0) ||
			(rflush && plc.lay.rused[end] > 0) {
			if plc.have.Y >= plc.needed.Y {
				break
			}
			plc.start += dir
			goto seekStart
		}
		if plc.lay.avail[end] < plc.have.X {
			plc.have.X = plc.lay.avail[end]
		}
		plc.have.Y++
		end += dir
	}

	if end < plc.start {
		plc.start = end + 1
	}

	plc.ok = plc.have.Y >= plc.needed.Y
}

// Render renders the placement, if it has been resolved successfully.
func (plc *LayoutPlacement) Render() {
	if !plc.ok {
		return
	}

	plc.align &= ^AlignHFlush
	off, used := 0, []int(nil)
	delta := 0

	switch plc.align & AlignCenter {
	case AlignLeft:
		off = moremath.MaxInt(plc.lay.lused[plc.start : plc.start+plc.have.Y]...)
		if off == 0 {
			plc.align |= AlignHFlush
		}
		used = plc.lay.lused
		delta = off

	case AlignRight:
		delta = moremath.MaxInt(plc.lay.rused[plc.start : plc.start+plc.have.Y]...)
		if delta == 0 {
			plc.align |= AlignHFlush
		}
		off = plc.lay.Grid.Size.X - plc.have.X - delta
		used = plc.lay.rused

	default: // NOTE: defaults to AlignCenter:
		lused := moremath.MaxInt(plc.lay.lused[plc.start : plc.start+plc.have.Y]...)
		rused := moremath.MaxInt(plc.lay.rused[plc.start : plc.start+plc.have.Y]...)
		off = lused + (plc.lay.Grid.Size.X-plc.have.X-lused-rused)/2
		used = plc.lay.cused
	}

	grid := MakeGrid(plc.have)
	plc.ren.Render(grid)
	plc.copy(grid, off)
	delta += plc.have.X

	for y, i := plc.start, 0; i < plc.have.Y; y, i = y+1, i+1 {
		used[y] = delta
		plc.lay.avail[y] -= plc.have.X
	}
}

func (plc *LayoutPlacement) copy(g Grid, off int) {
	var (
		left   = plc.align&AlignCenter == AlignLeft
		right  = plc.align&AlignCenter == AlignRight
		center = plc.align&AlignCenter == AlignCenter
		lflush = plc.align&AlignHFlush != 0 && left
		rflush = plc.align&AlignHFlush != 0 && right
		pad    = plc.sep
		paded  point.Point
	)

	bound := trim(g)

	if dx := moremath.MaxInt(0, g.Size.X-bound.BottomRight.X) + bound.TopLeft.X; dx > 0 {
		if right {
			off += dx
		} else if center {
			off += dx / 2
		}
	}

	// pad left
	if pad.Ch != 0 {
		if left && !lflush {
			for ly, gy := plc.start, bound.TopLeft.Y; gy < bound.BottomRight.Y; ly, gy = ly+1, gy+1 {
				li := ly*plc.lay.Grid.Size.X + off
				plc.lay.Grid.Data[li] = pad
			}
			off++
			paded.X++
			pad.Ch = 0
		} else if right && !rflush {
			off--
		} else {
			pad.Ch = 0
		}
	}

	// actual copy
	for ly, gy := plc.start, bound.TopLeft.Y; gy < bound.BottomRight.Y; ly, gy = ly+1, gy+1 {
		li := ly*plc.lay.Grid.Size.X + off
		gi := gy*g.Size.X + bound.TopLeft.X
		for gx := bound.TopLeft.X; gx < bound.BottomRight.X; gx++ {
			plc.lay.Grid.Data[li] = g.Data[gi]
			li++
			gi++
		}
	}

	// pad right
	if pad.Ch != 0 {
		off += bound.BottomRight.X - bound.TopLeft.X
		for ly, gy := plc.start, bound.TopLeft.Y; gy < bound.BottomRight.Y; ly, gy = ly+1, gy+1 {
			li := ly*plc.lay.Grid.Size.X + off
			plc.lay.Grid.Data[li] = pad
		}
		paded.X++
	}

	plc.have = bound.Size().Add(paded)
}

func trim(g Grid) (bound point.Box) {
	anyCol, anyRow := usedColumns(g)
	bound.BottomRight = g.Size

	// trim top
	for y := 0; y < bound.BottomRight.Y; y++ {
		if anyRow[y] {
			break
		}
		bound.TopLeft.Y++
	}

	// trim left
	for x := 0; x < bound.BottomRight.X; x++ {
		if anyCol[x] {
			break
		}
		bound.TopLeft.X++
	}

	// trim right
	for x := bound.BottomRight.X - 1; x >= bound.TopLeft.X; x-- {
		if anyCol[x] {
			break
		}
		bound.BottomRight.X--
	}

	// trim top
	for y := bound.BottomRight.Y - 1; y >= bound.BottomRight.Y; y-- {
		if anyRow[y] {
			break
		}
		bound.BottomRight.Y--
	}

	return bound
}

func usedColumns(g Grid) (anyCol, anyRow []bool) {
	anyCol = make([]bool, g.Size.X)
	anyRow = make([]bool, g.Size.Y)
	for y, i := 0, 0; i < len(g.Data); y++ {
		for x := 0; x < g.Size.X; x++ {
			if ch := g.Data[i].Ch; ch != 0 {
				anyCol[x] = true
				anyRow[y] = true
			}
			if fg := g.Data[i].Fg; fg != 0 {
				anyCol[x] = true
				anyRow[y] = true
			}
			if bg := g.Data[i].Bg; bg != 0 {
				anyCol[x] = true
				anyRow[y] = true
			}
			i++
		}
	}
	return anyCol, anyRow
}
