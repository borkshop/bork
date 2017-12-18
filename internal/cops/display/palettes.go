package display

import (
	"image/color"
)

// TerminalPalette is a limited palette of color for legacy terminals.
type TerminalPalette color.Palette

// Render the given colors to their closest palette equivalents.
func (tp TerminalPalette) Render(buf []byte, cur Cursor, fg, bg color.RGBA) ([]byte, Cursor) {
	if fg != cur.Foreground {
		i := color.Palette.Index(color.Palette(tp), fg)
		buf = append(buf, fgColorStrings[i]...)
		cur.Foreground = fg
	}
	if bg != cur.Background {
		i := color.Palette.Index(color.Palette(tp), bg)
		buf = append(buf, bgColorStrings[i]...)
		cur.Background = bg
	}
	return buf, cur
}

var (
	// Palette3 contains the first 8 Colors.
	Palette3 TerminalPalette

	// Palette4 contains the first 16 Colors.
	Palette4 TerminalPalette

	// Palette8 contains all 256 paletted virtual terminal colors.
	Palette8 TerminalPalette

	// colorIndex maps colors back to their palette index, suitable for mapping
	// arbitrary colors back to palette indexes in the 24 bit color model.
	colorIndex map[color.RGBA]int
)

func init() {
	for i := 0; i < 8; i++ {
		Palette3 = append(Palette3, color.Color(Colors[i]))
	}
	Model3 = Palette3.Render

	for i := 0; i < 16; i++ {
		Palette4 = append(Palette4, color.Color(Colors[i]))
	}
	Model4 = Palette4.Render

	for i := 0; i < 256; i++ {
		Palette8 = append(Palette8, color.Color(Colors[i]))
	}
	Model8 = Palette8.Render

	colorIndex = make(map[color.RGBA]int, 256)
	for i := 0; i < 256; i++ {
		colorIndex[Colors[i]] = i
	}
}
