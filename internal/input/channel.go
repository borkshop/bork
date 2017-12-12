package input

import (
	"bufio"
	"image"
	"io"
	"unicode"
)

type Move image.Point
type ShiftMove image.Point

// Channel returns a read channel for commands, distinguishable by type, and a
// a closer to stop channel's writer.
func Channel(reader io.Reader) (<-chan interface{}, func()) {
	ch := make(chan interface{})
	go func() {
		reader := bufio.NewReader(reader)
		for {
			r, _, err := reader.ReadRune()
			if err != nil {
				return
			}
			if pt, ok := parseExtViDir(r); ok {
				ch <- Move(pt)
			} else if pt, ok := parseExtViDir(unicode.ToLower(r)); ok {
				ch <- ShiftMove(pt)
			} else {
				ch <- r
			}
		}
	}()
	return ch, func() {
		// TODO abort goroutine
	}
}
