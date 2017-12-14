package view

import (
	"fmt"
	"image"
	"unicode/utf8"

	"github.com/borkshop/bork/internal/cops/display"
)

// RenderString constructs a static string Renderable; either the entire string
// is rendered, or not; no truncation is supported.
func RenderString(mess string, args ...interface{}) Renderable {
	return renderStringT(fmt.Sprintf(mess, args...))
}

type renderStringT string

func (rs renderStringT) RenderSize() (wanted, needed image.Point) {
	needed.X = utf8.RuneCountInString(string(rs))
	needed.Y = 1
	return needed, needed
}

func (rs renderStringT) Render(d *display.Display) {
	d.WriteString(0, 0, nil, nil, string(rs))
}
