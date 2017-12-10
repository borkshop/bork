package display

import (
	"image/color"
)

// Model is the interface for a terminal color rendering model.
type Model struct {
	foreground func([]byte, color.RGBA) []byte
	background func([]byte, color.RGBA) []byte
}

// RenderRGBA renders the given colors to their nearest counterparts within the
// available model colors
func (m Model) RenderRGBA(buf []byte, cur Cursor, fg, bg color.RGBA) ([]byte, Cursor) {
	if fg != cur.Foreground {
		buf = m.foreground(buf, fg)
		cur.Foreground = fg
	}
	if bg != cur.Background {
		buf = m.background(buf, bg)
		cur.Background = bg
	}
	return buf, cur
}

var (
	// Model0 is the monochrome color model, which does not print escape
	// sequences for any colors.
	Model0 = Model{renderNoColor, renderNoColor}

	// Model3 supports the first 8 color terminal palette.
	Model3 = Model{renderForegroundColor3, renderBackgroundColor3}

	// Model4 supports the first 16 color terminal palette, the same as Model3
	// but doubled for high intensity variants.
	Model4 = Model{renderForegroundColor4, renderBackgroundColor4}

	// Model8 supports a 256 color terminal palette, comprised of the 16
	// previous colors, a 6x6x6 color cube, and a 24 gray scale.
	Model8 = Model{renderForegroundColor8, renderBackgroundColor8}

	// Model24 supports all 24 bit colors, using palette colors only for exact
	// matches.
	Model24 = Model{renderForegroundColor24, renderBackgroundColor24}
)
