package display

import (
	"image/color"
)

// ColorModel renders colors to a partiular terminal color rendering protocol.
type ColorModel func(buf []byte, cur Cursor, fg, bg color.RGBA) ([]byte, Cursor)

var (
	// Model0 is the monochrome color model, which does not print escape
	// sequences for any colors.
	Model0 ColorModel = renderNoColor

	// Model3 supports the first 8 color terminal palette.
	Model3 ColorModel = Palette3.Render

	// Model4 supports the first 16 color terminal palette, the same as Model3
	// but doubled for high intensity variants.
	Model4 ColorModel = Palette4.Render

	// Model8 supports a 256 color terminal palette, comprised of the 16
	// previous colors, a 6x6x6 color cube, and a 24 gray scale.
	Model8 ColorModel = Palette8.Render

	// Model24 supports all 24 bit colors, and renders only to 24-bit terminal
	// sequences.
	Model24 ColorModel = renderJustColor24

	// ModelCompat24 supports all 24 bit colors, using palette colors only for exact
	// matches.
	ModelCompat24 ColorModel = renderCompatColor24
)
