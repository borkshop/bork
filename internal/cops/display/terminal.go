package display

import (
	"io"
	"os"

	"github.com/borkshop/bork/internal/cops/terminal"
)

// Terminal manages rendering a display buffer rendered to a terminal.
type Terminal struct {
	*Display

	out   *os.File
	term  terminal.Terminal
	model ColorModel
	buf   []byte
	cur   Cursor
}

// NewTerminal takes control of a terminal, readying it for rendering by
// putting it in raw mode, clearing it, and hiding the cursor.
func NewTerminal(out *os.File) (*Terminal, error) {
	term := &Terminal{
		out:   out,
		term:  terminal.New(out.Fd()),
		model: Model24, // TODO #choices
		buf:   make([]byte, 0, 65536),
		cur:   Start,
	}
	return term, term.open()
}

func (term *Terminal) open() error {
	if err := term.UpdateSize(); err != nil {
		return err
	}
	if err := term.term.SetRaw(); err != nil {
		return err
	}
	term.curse(
		Cursor.Home,
		Cursor.Clear,
		Cursor.Hide,
	)
	return nil
}

// Close the terminal, clearing it and restoring state.
func (term *Terminal) Close() error {
	term.curse(
		Cursor.Home,
		Cursor.Clear,
		Cursor.Show,
	)
	err := term.flush()
	if rerr := term.term.Restore(); err == nil {
		err = rerr
	}
	return err
}

func (term *Terminal) curse(words ...func(Cursor, []byte) ([]byte, Cursor)) {
	buf, cur := term.buf, term.cur
	for _, word := range words {
		buf, cur = word(cur, buf)
	}
	term.buf, term.cur = buf, cur
}

func (term *Terminal) flush() error {
	if len(term.buf) == 0 {
		return nil
	}
	attempts := 5 // TODO sanity check: is this even worthwhile?
	n, err := term.out.Write(term.buf)
	for attempts > 1 && err == io.ErrShortWrite {
		attempts--
		term.buf = term.buf[:copy(term.buf, term.buf[n:])]
		n, err = term.out.Write(term.buf)
	}
	term.buf = term.buf[:0]
	return err
}

// UpdateSize updates the terminal buffer to match the
// terminal's current size.
func (term *Terminal) UpdateSize() error {
	bounds, err := term.term.Bounds()
	if err == nil {
		term.Display = New(bounds)
		// TODO re-use underlying capacity by drilling down to a sub-display
		// when decreasing, and conversely when increasing
	}
	return err
}

func (term Terminal) fullRender(cur Cursor, buf []byte) ([]byte, Cursor) {
	return Render(buf, cur, term.Display, term.model)
}

// Render the display buffer to the terminal.
func (term *Terminal) Render() error {
	// TODO support differential render
	term.curse(
		term.fullRender,
		Cursor.Reset,
	)
	return term.flush()
}
