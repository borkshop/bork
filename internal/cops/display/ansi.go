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

func renderNoColor(buf []byte, c color.RGBA) []byte {
	return buf
}

func renderBackgroundColor3(buf []byte, c color.RGBA) []byte {
	return renderBackgroundColor(buf, Palette3, c)
}

func renderForegroundColor3(buf []byte, c color.RGBA) []byte {
	return renderForegroundColor(buf, Palette3, c)
}

func renderBackgroundColor4(buf []byte, c color.RGBA) []byte {
	return renderBackgroundColor(buf, Palette4, c)
}

func renderForegroundColor4(buf []byte, c color.RGBA) []byte {
	return renderForegroundColor(buf, Palette4, c)
}

func renderBackgroundColor8(buf []byte, c color.RGBA) []byte {
	return renderBackgroundColor(buf, Palette8, c)
}

func renderForegroundColor8(buf []byte, c color.RGBA) []byte {
	return renderForegroundColor(buf, Palette8, c)
}

func renderForegroundColor(buf []byte, p color.Palette, c color.RGBA) []byte {
	i := p.Index(c)
	return append(buf, fgColorStrings[i]...)
}

func renderBackgroundColor(buf []byte, p color.Palette, c color.RGBA) []byte {
	i := p.Index(c)
	return append(buf, bgColorStrings[i]...)
}

func renderForegroundColor24(buf []byte, c color.RGBA) []byte {
	if i, ok := colorIndex[c]; ok {
		return append(buf, fgColorStrings[i]...)
	}
	return renderColor24(append(buf, "\033[38;2"...), c)
}

func renderBackgroundColor24(buf []byte, c color.RGBA) []byte {
	if i, ok := colorIndex[c]; ok {
		return append(buf, bgColorStrings[i]...)
	}
	return renderColor24(append(buf, "\033[48;2"...), c)
}

func renderColor24(buf []byte, c color.RGBA) []byte {
	buf = append(buf, byteStrings[c.R]...)
	buf = append(buf, byteStrings[c.G]...)
	buf = append(buf, byteStrings[c.B]...)
	buf = append(buf, "m"...)
	return buf
}
