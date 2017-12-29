// Package input provides an adapter that changes a stream of terminal input
// into a channel of commands.
// To use the command channel, read command by command and switch on the
// command type.
package input

import (
	"bufio"
	"image"
	"io"
	"unicode"
)

// Move captures a motion command. The data are a unit vector.
type Move image.Point

// ShiftMove captures a motion command with the shift key pressed. The data are a unit vector.
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
