package hud

import (
	"image"
	"image/draw"
	"unicode/utf8"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/view"
)

// HUD provides an opinionated view system with a Header, Footer, and Logs on
// top of a base display (e.g world map).
type HUD struct {
	World *display.Display
	Logs  Logs

	parts []view.Renderable
	align []view.Align
}

// Render the context into the given display buffer
func (hud HUD) Render(d *display.Display) {
	// NOTE: intentionally not a layout item so that the UI elements overlay
	// the world display.

	// TODO factor out DrawCentered
	bound, off := d.Rect, image.ZP
	if n := bound.Dx() - hud.World.Rect.Dx(); n > 0 {
		bound.Min.X += n / 2
	} else if n < 0 {
		off.X -= n     // align Mins (expected by draw clipping logic)
		off.X += n / 2 // center the source window
	}
	if n := bound.Dy() - hud.World.Rect.Dy(); n > 0 {
		bound.Min.Y += n / 2
	} else if n < 0 {
		off.Y -= n     // align Mins (expected by draw clipping logic)
		off.Y += n / 2 // center the source window
	}
	display.Draw(d, bound, hud.World, off, draw.Over)

	if len(hud.Logs.Buffer) > 0 {
		// TODO: scrolling
		if hud.Logs.Align == 0 {
			hud.AddRenderable(hud.Logs, view.AlignTop|view.AlignCenter)
		} else {
			hud.AddRenderable(hud.Logs, hud.Logs.Align)
		}
	}

	lay := view.Layout{Display: d}
	for i := range hud.parts {
		lay.Render(hud.parts[i], hud.align[i])
	}
}

// HeaderF adds a static string part to the header; the mess string may begin
// with layout markers such as "<^>" to cause left, center, right alignment;
// mess may also start with "." to cause an alignment flush (otherwise the
// layout tries to pack as many parts onto one line as possible).
func (hud *HUD) HeaderF(mess string, args ...interface{}) {
	align, n := readLayoutOpts(mess)
	hud.AddRenderable(view.RenderString(mess[n:], args...), align|view.AlignTop)
}

// FooterF adds a static string to the header; the same alignment marks are
// available as to AddHeader.
func (hud *HUD) FooterF(mess string, args ...interface{}) {
	align, n := readLayoutOpts(mess)
	hud.AddRenderable(view.RenderString(mess[n:], args...), align|view.AlignBottom)
}

// AddRenderable adds an aligned Renderable to the hud.
func (hud *HUD) AddRenderable(ren view.Renderable, align view.Align) {
	hud.parts = append(hud.parts, ren)
	hud.align = append(hud.align, align)
}

func readLayoutOpts(s string) (opts view.Align, n int) {
	for len(s) > 0 {
		switch r, m := utf8.DecodeRuneInString(s[n:]); r {
		case '.':
			opts |= view.AlignHFlush
			n += m
			continue
		case '<':
			opts |= view.AlignLeft
			n += m
		case '>':
			opts |= view.AlignRight
			n += m
		case '^':
			opts |= view.AlignCenter
			n += m
		}
		break
	}
	return opts, n
}
