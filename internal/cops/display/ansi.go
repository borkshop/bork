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

func renderCompatColor24(buf []byte, cur Cursor, fg, bg color.RGBA) ([]byte, Cursor) {
	if fg != cur.Foreground {
		if i, ok := colorIndex[fg]; ok {
			buf = append(buf, fgColorStrings[i]...)
		} else {
			buf = append(buf, "\033[38;2"...)
			buf = renderColor24(buf, fg)
		}
		cur.Foreground = fg
	}
	if bg != cur.Background {
		if i, ok := colorIndex[bg]; ok {
			buf = append(buf, bgColorStrings[i]...)
		} else {
			buf = append(buf, "\033[48;2"...)
			buf = renderColor24(buf, bg)
		}
		cur.Background = bg
	}
	return buf, cur
}

func renderJustColor24(buf []byte, cur Cursor, fg, bg color.RGBA) ([]byte, Cursor) {
	if fg != cur.Foreground {
		buf = append(buf, "\033[38;2"...)
		buf = renderColor24(buf, fg)
		cur.Foreground = fg
	}
	if bg != cur.Background {
		buf = append(buf, "\033[48;2"...)
		buf = renderColor24(buf, bg)
		cur.Background = bg
	}
	return buf, cur
}

func renderColor24(buf []byte, c color.RGBA) []byte {
	buf = append(buf, byteStrings[c.R]...)
	buf = append(buf, byteStrings[c.G]...)
	buf = append(buf, byteStrings[c.B]...)
	buf = append(buf, "m"...)
	return buf
}
