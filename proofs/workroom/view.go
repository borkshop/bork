package main

import (
	"image"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/views"

	"github.com/borkshop/bork/internal/input"
)

type worldView struct {
	views.WidgetWatchers
	view  views.View
	port  *views.ViewPort
	world *worldT
}

func newView(world *worldT) *worldView {
	v := &worldView{}
	v.world = world
	v.port = views.NewViewPort(nil, 0, 0, 0, 0)
	return v
}

func (v *worldView) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyRune:
			any := false
			move, have := image.ZP, false
			if r := ev.Rune(); r == '.' {
				have = true
			} else {
				move, have = input.ParseMove(r, image.ZP)
			}
			if have {
				for it := v.world.Iter(wcPlayerControl.All()); it.Next(); {
					v.world.moves.AddPendingMove(it.Entity(), move, 1, 4)
					any = true
				}
			}
			if any {
				v.world.postProc()
			}
			return any
		}
	}
	return false
}

func (v *worldView) Size() (int, int) {
	box := v.world.pos.Bounds()
	return box.Dx(), box.Dy()
}

func (v *worldView) SetView(view views.View) {
	v.port.SetView(view)
	v.view = view
	if v.view == nil {
		return
	}
	v.Resize()
	v.PostEventWidgetContent(v) // XXX
}

func (v *worldView) Resize() {
	v.updateSize()
}

func (v *worldView) updateSize() image.Rectangle {
	box := v.world.pos.Bounds()
	px, py := 0, 0
	vw, vh := v.view.Size()
	if n := vw - box.Dx() - 1; n > 0 {
		px = n / 2
		vw -= px
	}
	if n := vh - box.Dy() - 1; n > 0 {
		py = n / 2
		vh -= py
	}
	v.port.Resize(px, py, vw, vh)
	v.port.SetContentSize(box.Dx(), box.Dy(), true)
	return box
}

func (v *worldView) Draw() {
	if v.view == nil {
		return
	}
	box := v.updateSize()
	for it := v.world.Iter(wcPlayerControl.All()); it.Next(); {
		if pt, ok := v.world.pos.Get(it.Entity()); ok {
			v.port.MakeVisible(pt.X, pt.Y)
		}
	}
	v.draw(box)
}

func (v *worldView) draw(box image.Rectangle) {
	ch := world.glyphs[0]
	if ch == 0 {
		ch = ' '
	}
	v.port.Fill(ch, tcell.StyleDefault.Background(world.bg[0]).Foreground(world.fg[0]))

	// TODO this EPS-order draw loop is probably ill-conceived: it design to
	// skip around memory, rather than processing components in memory-order;
	// rather than this whole drawState thing, we should keep max-zval for each
	// cell; flushing to view port after resolving all cells locally.
	ds := drawState{
		port: v.port,
		bg:   tcell.ColorDefault,
		fg:   tcell.ColorDefault,
	}

	// TODO eps range query
	off := box.Min.Mul(-1)
	sbox := image.Rect(v.port.GetVisible())
	for it := v.world.pos.Iter(); it.Next(); {
		if !it.Type().HasAny(wcGlyph | wcBG | wcFG) {
			continue
		}
		pt, _ := v.world.pos.Get(it.Entity())
		if spt := pt.Add(off); spt.In(sbox) {
			ds.advance(spt)
			t, id := it.Type(), it.ID()
			z := v.world.zval[id]
			if t.HasAll(wcGlyph) {
				ds.mergeCH(z, v.world.glyphs[id])
			}
			if t.HasAll(wcBG) {
				ds.mergeBG(z, v.world.bg[id])
			}
			if t.HasAll(wcFG) {
				ds.mergeFG(z, v.world.fg[id])
			}
		}
	}
	ds.flush()
}

type drawState struct {
	port     *views.ViewPort
	pt       image.Point
	ch       rune
	bg, fg   tcell.Color
	zCH      uint8
	zBG, zFG uint8
}

func (ds *drawState) mergeCH(z uint8, ch rune) {
	if z > ds.zCH {
		if ch != 0 {
			ds.ch = ch
		}
		ds.zCH = z
	}
}
func (ds *drawState) mergeBG(z uint8, bg tcell.Color) {
	if z > ds.zBG {
		if bg != tcell.ColorDefault {
			ds.bg = bg
		}
		ds.zBG = z
	}
}
func (ds *drawState) mergeFG(z uint8, fg tcell.Color) {
	if z > ds.zFG {
		if fg != tcell.ColorDefault {
			ds.fg = fg
		}
		ds.zFG = z
	}
}

func (ds *drawState) advance(pt image.Point) {
	if pt.Eq(ds.pt) {
		return
	}
	ds.flush()
	ds.pt = pt
}

func (ds *drawState) flush() {
	style := tcell.StyleDefault
	if ds.bg != tcell.ColorDefault {
		style = style.Background(ds.bg)
	}
	if ds.fg != tcell.ColorDefault {
		style = style.Foreground(ds.fg)
	}
	if ds.ch == 0 && ds.bg != tcell.ColorDefault {
		ds.ch = ' '
	}
	if ds.ch != 0 {
		ds.port.SetContent(ds.pt.X, ds.pt.Y, ds.ch, nil, style)
	}
	ds.ch, ds.bg, ds.fg = 0, tcell.ColorDefault, tcell.ColorDefault
	ds.zCH, ds.zBG, ds.zFG = 0, 0, 0
}
