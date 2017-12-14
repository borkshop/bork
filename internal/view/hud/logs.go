package hud

import (
	"fmt"
	"image"
	"unicode/utf8"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/moremath"
	"github.com/borkshop/bork/internal/view"
)

// Logs represents a renderable buffer of log messages.
type Logs struct {
	Buffer   []string
	Align    view.Align
	Min, Max int
}

// Init initializes the log buffer and metadata, allocating the given capacity.
func (logs *Logs) Init(logCap int) {
	logs.Align = view.AlignTop | view.AlignCenter
	logs.Min = 5
	logs.Max = 10
	logs.Buffer = make([]string, 0, logCap)
}

// RenderSize returns the desired and necessary sizes for rendering.
func (logs Logs) RenderSize() (wanted, needed image.Point) {
	needed.X = 1
	needed.Y = moremath.MinInt(len(logs.Buffer), logs.Min)
	wanted.X = 1
	wanted.Y = moremath.MinInt(len(logs.Buffer), logs.Max)
	for i := range logs.Buffer {
		if n := utf8.RuneCountInString(logs.Buffer[i]); n > wanted.X {
			wanted.X = n
		}
	}
	if needed.Y > wanted.Y {
		needed.Y = wanted.Y
	}
	return wanted, needed
}

// Render renders the log buffer.
func (logs Logs) Render(d *display.Display) {
	off := len(logs.Buffer) - d.Rect.Max.Y
	if off < 0 {
		off = 0
	}
	for i, y := off, 0; i < len(logs.Buffer); i, y = i+1, y+1 {
		d.WriteString(0, y, nil, nil, logs.Buffer[i])
	}
}

// Log formats and appends a log message to the buffer, discarding the oldest
// message if full.
func (logs *Logs) Log(mess string, args ...interface{}) {
	mess = fmt.Sprintf(mess, args...)
	if len(logs.Buffer) < cap(logs.Buffer) {
		logs.Buffer = append(logs.Buffer, mess)
	} else {
		copy(logs.Buffer, logs.Buffer[1:])
		logs.Buffer[len(logs.Buffer)-1] = mess
	}
}
