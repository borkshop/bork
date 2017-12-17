package input

import (
	"bufio"
	"io"
)

// Recognizer is an input rune recognizer. It gets passed a single read rune,
// and may scan zero or more additional runes.
type Recognizer func(rune, io.RuneScanner) (interface{}, error)

// Channel returns a channel fed by scanning runes through a delegate list of
// recognizers. If an IO or recognizer error occurs, it is put on the channel,
// and the channel is closed.
func Channel(r io.Reader, recs ...Recognizer) (<-chan interface{}, func()) {
	ch := make(chan interface{}, 1)
	go func(rs io.RuneScanner) {
		for {
			if val, err := scanOne(rs, recs); err != nil {
				ch <- err
				close(ch)
				return
			} else if val != nil {
				ch <- val
			}
		}
	}(bufio.NewReader(r))
	return ch, func() {
		// TODO (kris) abort goroutine
		// TODO (josh) we'll probably need to do some non-blocking sigio
		// shenanigans like termbox does (at least on unix).
	}
}

func scanOne(rs io.RuneScanner, recs []Recognizer) (interface{}, error) {
	r, _, err := rs.ReadRune()
	if err != nil {
		return nil, err
	}
	for _, rec := range recs {
		if val, err := rec(r, rs); val != nil || err != nil {
			return val, err
		}
	}
	return r, nil
}
