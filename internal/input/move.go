package input

import (
	"image"
	"unicode"
)

// ParseMove parses an X/Y move from the given rune using the
// classic extended-vi roguelike keybindings of h/j/k/l and
// y/u/b/n. If the extra point is non-zero, then capitalized
// moves are parsed as a componentwise-multiple of it.
//
// Returns the parsed point and true if the rune was
// recognized, zero point and false otherwise.
func ParseMove(ch rune, extra image.Point) (image.Point, bool) {
	if pt, ok := parseExtViDir(ch); ok {
		return pt, true
	}
	if !extra.Eq(image.ZP) {
		if pt, ok := parseExtViDir(unicode.ToLower(ch)); ok {
			pt = image.Pt(extra.X*pt.X, extra.Y*pt.Y)
			return pt, true
		}
	}
	return image.ZP, false
}

func parseExtViDir(ch rune) (image.Point, bool) {
	switch ch {
	case 'h':
		return image.Pt(-1, 0), true
	case 'l':
		return image.Pt(1, 0), true
	case 'k':
		return image.Pt(0, -1), true
	case 'j':
		return image.Pt(0, 1), true
	case 'y':
		return image.Pt(-1, -1), true
	case 'u':
		return image.Pt(1, -1), true
	case 'b':
		return image.Pt(-1, 1), true
	case 'n':
		return image.Pt(1, 1), true
	}
	return image.ZP, false
}
