package input

import (
	"bytes"
	"strconv"
)

const (
	EventKey EventType = iota
	EventMouse
)

const (
	ModAlt Modifier = 1 << iota
	ModMotion
)

type Event struct {
	Type   EventType // one of Event* constants
	Mod    Modifier  // one of Mod* constants or 0
	Key    Key       // one of Key* constants, invalid if 'Ch' is not 0
	Ch     rune      // a unicode character
	MouseX int       // x coord of mouse
	MouseY int       // y coord of mouse
}

const (
	KeyF1 Key = 0xFFFF - iota
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
	KeyInsert
	KeyDelete
	KeyHome
	KeyEnd
	KeyPgup
	KeyPgdn
	KeyArrowUp
	KeyArrowDown
	KeyArrowLeft
	KeyArrowRight
	keyMin // see terminfo
	MouseLeft
	MouseMiddle
	MouseRight
	MouseRelease
	MouseWheelUp
	MouseWheelDown
)

func parseMouseEvent(buf []byte) (Event, int, bool) {
	if len(buf) >= 6 && bytes.HasPrefix(buf, "\033[M") {
		// X10 mouse encoding, the simplest one
		// \033 [ M Cb Cx Cy
		b := buf[3] - 32
		switch b & 3 {
		case 0:
			if b&64 != 0 {
				event.Key = MouseWheelUp
			} else {
				event.Key = MouseLeft
			}
		case 1:
			if b&64 != 0 {
				event.Key = MouseWheelDown
			} else {
				event.Key = MouseMiddle
			}
		case 2:
			event.Key = MouseRight
		case 3:
			event.Key = MouseRelease
		default:
			return 6, false
		}
		event.Type = EventMouse // KeyEvent by default
		if b&32 != 0 {
			event.Mod |= ModMotion
		}

		// the coord is 1,1 for upper left
		event.MouseX = int(buf[4]) - 1 - 32
		event.MouseY = int(buf[5]) - 1 - 32
		return 6, true
	}

	if bytes.HasPrefix(buf, "\033[<") ||
		bytes.HasPrefix(buf, "\033[") {
		// xterm 1006 extended mode or urxvt 1015 extended mode
		// xterm: \033 [ < Cb ; Cx ; Cy (M or m)
		// urxvt: \033 [ Cb ; Cx ; Cy M

		// find the first M or m, that's where we stop
		mi := bytes.IndexAny(buf, "Mm")
		if mi == -1 {
			return 0, false
		}

		// whether it's a capital M or not
		isM := buf[mi] == 'M'

		// whether it's urxvt or not
		isU := false

		// buf[2] is safe here, because having M or m found means we have at
		// least 3 bytes in a string
		if buf[2] == '<' {
			buf = buf[3:mi]
		} else {
			isU = true
			buf = buf[2:mi]
		}

		s1 := bytes.Index(buf, ";")
		s2 := bytes.LastIndex(buf, ";")
		// not found or only one ';'
		if s1 == -1 || s2 == -1 || s1 == s2 {
			return 0, false
		}

		n1, err := strconv.ParseInt(buf[0:s1], 10, 64)
		if err != nil {
			return 0, false
		}
		n2, err := strconv.ParseInt(buf[s1+1:s2], 10, 64)
		if err != nil {
			return 0, false
		}
		n3, err := strconv.ParseInt(buf[s2+1:], 10, 64)
		if err != nil {
			return 0, false
		}

		// on urxvt, first number is encoded exactly as in X10, but we need to
		// make it zero-based, on xterm it is zero-based already
		if isU {
			n1 -= 32
		}
		switch n1 & 3 {
		case 0:
			if n1&64 != 0 {
				event.Key = MouseWheelUp
			} else {
				event.Key = MouseLeft
			}
		case 1:
			if n1&64 != 0 {
				event.Key = MouseWheelDown
			} else {
				event.Key = MouseMiddle
			}
		case 2:
			event.Key = MouseRight
		case 3:
			event.Key = MouseRelease
		default:
			return mi + 1, false
		}
		if !isM {
			// on xterm mouse release is signaled by lowercase m
			event.Key = MouseRelease
		}

		event.Type = EventMouse // KeyEvent by default
		if n1&32 != 0 {
			event.Mod |= ModMotion
		}

		event.MouseX = int(n2) - 1
		event.MouseY = int(n3) - 1
		return mi + 1, true
	}

	return 0, false
}
