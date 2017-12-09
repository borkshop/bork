package display

import (
	"image/color"
	"strconv"
)

var (
	byteStrings    [256]string
	fgColorStrings [256]string
	bgColorStrings [256]string
)

func init() {
	for i := 0; i < len(byteStrings); i++ {
		byteStrings[i] = ";" + strconv.Itoa(i)
	}

	i := 0
	for ; i < 8; i++ {
		fgColorStrings[i] = "\033[" + strconv.Itoa(30+i) + "m"
		bgColorStrings[i] = "\033[" + strconv.Itoa(40+i) + "m"
	}
	for ; i < 16; i++ {
		fgColorStrings[i] = "\033[" + strconv.Itoa(90-8+i) + "m"
		bgColorStrings[i] = "\033[" + strconv.Itoa(100-8+i) + "m"
	}
	for ; i < 256; i++ {
		fgColorStrings[i] = "\033[38;5;" + strconv.Itoa(i) + "m"
		bgColorStrings[i] = "\033[48;5;" + strconv.Itoa(i) + "m"
	}
}

func renderNoColor(buf []byte, cur Cursor, _, _ color.RGBA) ([]byte, Cursor) { return buf, cur }

var (
	fgColorCache = make(map[color.RGBA]string, 1024)
	bgColorCache = make(map[color.RGBA]string, 1024)
)

func renderCompatColor24(buf []byte, cur Cursor, fg, bg color.RGBA) ([]byte, Cursor) {
	if fg != cur.Foreground {
		if i, ok := colorIndex[fg]; ok {
			buf = append(buf, fgColorStrings[i]...)
		} else {
			s, ok := fgColorCache[fg]
			if !ok {
				s = "\033[38;2" + byteStrings[fg.R] + byteStrings[fg.G] + byteStrings[fg.B] + "m"
				fgColorCache[fg] = s
			}
			buf = append(buf, s...)
		}
		cur.Foreground = fg
	}
	if bg != cur.Background {
		if i, ok := colorIndex[bg]; ok {
			buf = append(buf, bgColorStrings[i]...)
		} else {
			s, ok := bgColorCache[bg]
			if !ok {
				s = "\033[48;2" + byteStrings[bg.R] + byteStrings[bg.G] + byteStrings[bg.B] + "m"
				bgColorCache[bg] = s
			}
			buf = append(buf, s...)
		}
		cur.Background = bg
	}
	return buf, cur
}

func renderJustColor24(buf []byte, cur Cursor, fg, bg color.RGBA) ([]byte, Cursor) {
	if fg != cur.Foreground {
		s, ok := fgColorCache[fg]
		if !ok {
			s = "\033[38;2" + byteStrings[fg.R] + byteStrings[fg.G] + byteStrings[fg.B] + "m"
			fgColorCache[fg] = s
		}
		buf = append(buf, s...)
		cur.Foreground = fg
	}
	if bg != cur.Background {
		s, ok := bgColorCache[bg]
		if !ok {
			s = "\033[48;2" + byteStrings[bg.R] + byteStrings[bg.G] + byteStrings[bg.B] + "m"
			bgColorCache[bg] = s
		}
		buf = append(buf, s...)
		cur.Background = bg
	}
	return buf, cur
}
