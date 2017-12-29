package display

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"unicode/utf8"
)

// Cursor models the known or unknown states of a cursor.
type Cursor struct {
	// Position is the position of the cursor.
	// Negative values indicate that the X or Y position is not known,
	// so the next position change must be relative to the beginning of the
	// same line or possibly the origin.
	Position image.Point

	// Foreground is the foreground color for subsequent text.
	// Transparent indicates that the color is unknown, so the next text must
	// be preceded by an SGR (set graphics) ANSI sequence to set it.
	Foreground color.RGBA

	// Foreground is the foreground color for subsequent text.
	// Transparent indicates that the color is unknown, so the next text must
	// be preceded by an SGR (set graphics) ANSI sequence to set it.
	Background color.RGBA

	// Visibility indicates whether the cursor is visible.
	Visibility Visibility
}

// Visibility represents the visibility of a Cursor.
type Visibility int

const (
	// Hidden represents a hidden cursor.
	Hidden Visibility = iota + 1

	// Visible represents a normal cursor.
	Visible
)

func (v Visibility) String() string {
	switch v {
	case 0:
		return "Unknown"
	case Hidden:
		return "Hidden"
	case Visible:
		return "Visible"
	default:
		return fmt.Sprintf("Invalid<%d>", int(v))
	}
}

var (
	// Lost indicates that the cursor position is unknown.
	Lost = image.Point{-1, -1}

	// Start is a cursor state that makes no assumptions about the cursor's
	// position or colors, necessitating a seek from origin and explicit color
	// settings for the next text.
	Start = Cursor{
		Position:   Lost,
		Foreground: Transparent,
		Background: Transparent,
	}

	// Reset is a cursor state indicating that the cursor is at the origin
	// and that the foreground color is white (7), background black (0).
	// This is the state cur.Reset() returns to, and the state for which
	// cur.Reset() will append nothing to the buffer.
	Reset = Cursor{
		Position:   image.ZP,
		Foreground: Colors[7],
		Background: Colors[0],
	}
)

// Hide hides the cursor.
func (c Cursor) Hide(buf []byte) ([]byte, Cursor) {
	if c.Visibility != Hidden {
		c.Visibility = Hidden
		buf = append(buf, "\033[?25l"...)
	}
	return buf, c
}

// Show reveals the cursor.
func (c Cursor) Show(buf []byte) ([]byte, Cursor) {
	if c.Visibility != Visible {
		c.Visibility = Visible
		buf = append(buf, "\033[?25h"...)
	}
	return buf, c
}

// Clear erases the whole display; implicitly invalidates the cursor position
// since its behavior is inconsistent across terminal implementations.
func (c Cursor) Clear(buf []byte) ([]byte, Cursor) {
	c.Position = Lost
	return append(buf, "\033[2J"...), c
}

// ClearLine erases the current line.
func (c Cursor) ClearLine(buf []byte) ([]byte, Cursor) {
	return append(buf, "\033[2K"...), c
}

// Reset returns the terminal to default white on black colors.
func (c Cursor) Reset(buf []byte) ([]byte, Cursor) {
	if c.Foreground != Colors[7] || c.Background != Colors[0] {
		//lint:ignore SA4005 broken check
		c.Foreground, c.Background = Colors[7], Colors[0]
		buf = append(buf, "\033[m"...)
	}
	return buf, c
}

// Home seeks the cursor to the origin, using display absolute coordinates.
func (c Cursor) Home(buf []byte) ([]byte, Cursor) {
	c.Position = image.ZP
	return append(buf, "\033[H"...), c
}

func (c Cursor) recover(buf []byte, to image.Point) ([]byte, Cursor) {
	if c.Position == Lost {
		// If the cursor position is completely unknown, move relative to
		// screen origin. This mode must be avoided to render relative to
		// cursor position inline with a scrolling log, by setting the cursor
		// position relative to an arbitrary origin before rendering.
		return c.jumpTo(buf, to)
	}

	if c.Position.X == -1 {
		// If only horizontal position is unknown, return to first column and
		// march forward. Rendering a non-ASCII cell of unknown or
		// indeterminate width may invalidate the column number. For example, a
		// skin tone emoji may or may not render as a single column glyph.
		buf = append(buf, "\r"...)
		c.Position.X = 0
		// Continue...
	}

	return buf, c
}

func (c Cursor) jumpTo(buf []byte, to image.Point) ([]byte, Cursor) {
	buf = append(buf, "\033["...)
	buf = strconv.AppendInt(buf, int64(to.Y+1), 10)
	buf = append(buf, ";"...)
	buf = strconv.AppendInt(buf, int64(to.X+1), 10)
	buf = append(buf, "H"...)
	c.Position = to
	return buf, c
}

func (c Cursor) linedown(buf []byte, n int) ([]byte, Cursor) {
	// Use \r\n to advance cursor Y on the chance it will advance the
	// display bounds.
	buf = append(buf, "\r\n"...)
	for m := n - 1; m > 0; m-- {
		buf = append(buf, "\n"...)
	}
	c.Position.X = 0
	c.Position.Y += n
	return buf, c
}

func (c Cursor) up(buf []byte, n int) ([]byte, Cursor) {
	buf = append(buf, "\033["...)
	buf = strconv.AppendInt(buf, int64(n), 10)
	buf = append(buf, "A"...)
	c.Position.Y -= n
	return buf, c
}

func (c Cursor) down(buf []byte, n int) ([]byte, Cursor) {
	buf = append(buf, "\033["...)
	buf = strconv.AppendInt(buf, int64(n), 10)
	buf = append(buf, "B"...)
	c.Position.Y += n
	return buf, c
}

func (c Cursor) left(buf []byte, n int) ([]byte, Cursor) {
	buf = append(buf, "\033["...)
	buf = strconv.AppendInt(buf, int64(n), 10)
	buf = append(buf, "D"...)
	c.Position.X -= n
	return buf, c
}

func (c Cursor) right(buf []byte, n int) ([]byte, Cursor) {
	buf = append(buf, "\033["...)
	buf = strconv.AppendInt(buf, int64(n), 10)
	buf = append(buf, "C"...)
	c.Position.X += n
	return buf, c
}

// Go moves the cursor to another position, preferring to use relative motion,
// using line relative if the column is unknown, using display origin relative
// only if the line is also unknown. If the column is unknown, use "\r" to seek
// to column 0 of the same line.
func (c Cursor) Go(buf []byte, to image.Point) ([]byte, Cursor) {
	buf, c = c.recover(buf, to)

	if to.X == 0 && to.Y == c.Position.Y+1 {
		buf, c = c.Reset(buf)
		buf = append(buf, "\r\n"...)
		c.Position.X = 0
		c.Position.Y++
	} else if to.X == 0 && c.Position.X != 0 {
		buf, c = c.Reset(buf)
		buf = append(buf, "\r"...)
		c.Position.X = 0

		// In addition to scrolling back to the first column generally, this
		// has the effect of resetting the column if writing a multi-byte
		// string invalidates the cursor's horizontal position. For example, a
		// skin tone emoji may or may not render as a single column glyph.
	}

	if n := to.Y - c.Position.Y; n > 0 {
		buf, c = c.linedown(buf, n)
	} else if n < 0 {
		buf, c = c.up(buf, -n)
	}

	if n := to.X - c.Position.X; n > 0 {
		buf, c = c.right(buf, n)
	} else if n < 0 {
		buf, c = c.left(buf, -n)
	}

	return buf, c
}

// TODO: func (c Cursor) Write(buf, p []byte) ([]byte, Cursor)

// WriteGlyph appends the given string's UTF8 bytes into the given
// buffer, invalidating the cursor if the string COULD HAVE rendered
// to more than one glyph; otherwise the cursor's X is advanced by 1.
func (c Cursor) WriteGlyph(buf []byte, s string) ([]byte, Cursor) {
	buf = append(buf, s...)
	if n := utf8.RuneCountInString(s); n == 1 {
		c.Position.X++
	} else {
		// Invalidate cursor column to force position reset
		// before next draw, if the string drawn might be longer
		// than one cell wide or simply empty.
		c.Position.X = -1
	}
	return buf, c
}
