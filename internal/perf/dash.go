package perf

import (
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/borkshop/bork/internal/point"
	"github.com/borkshop/bork/internal/view"
)

// Dash is a summary widget that can be triggered to show a perf dialog.
type Dash struct {
	*Perf
	notes map[string]string
	parts []string
}

// HandleKey handles key input for the perf dashboard.
func (da Dash) HandleKey(k view.KeyEvent) bool {
	switch k.Ch {
	case '*':
		da.Perf.shouldProfile = !da.Perf.shouldProfile
		return true
	}
	return false
}

// Note adds or updats an optional note in the dashboard.
func (da *Dash) Note(name, mess string, args ...interface{}) {
	if da.notes == nil {
		da.notes = make(map[string]string, 1)
	}
	da.notes[name] = fmt.Sprintf(mess, args...)
}

// RenderSize calculates the wanted/needed size render the dashboard.
func (da *Dash) RenderSize() (wanted, needed point.Point) {
	i := da.lastI()
	lastElapsed := da.Perf.time[i].end.Sub(da.Perf.time[i].start)
	ms := &da.Perf.memStats[i]
	da.Note("heap", "%v/%v", siBytes(ms.HeapAlloc), ms.HeapObjects)

	if len(da.parts) > 0 {
		da.parts = da.parts[:0]
	} else {
		da.parts = make([]string, 0, 1+len(da.notes))
	}

	da.parts = append(da.parts, fmt.Sprintf("t=%d Δt=%v", da.Perf.round, lastElapsed))
	needed.X = 2 + utf8.RuneCountInString(da.parts[0])
	needed.Y = 1
	wanted = needed

	for name, mess := range da.notes {
		part := fmt.Sprintf("%s=%s", name, mess)
		da.parts = append(da.parts, part)
	}
	sort.Strings(da.parts[1:])

	for _, part := range da.parts[1:] {
		wanted.X += utf8.RuneCountInString(part)
	}

	return wanted, needed
}

// Render the dashboard.
func (da *Dash) Render(g view.Grid) {
	x := 0
	g.Set(x, 0, da.status(), 0, 0)
	x++
	for i := 0; i < len(da.parts) && x < g.Size.X-1; i++ {
		x++
		x += g.WriteString(x, 0, da.parts[i])
	}
}

func (da Dash) lastI() int {
	i := da.Perf.i - 1
	if i < 0 {
		i += numSamples
	}
	return i
}

func (da Dash) status() rune {
	if da.Perf.err != nil {
		return '■'
	}
	if da.Perf.profiling {
		return '◉'
	}
	if da.Perf.shouldProfile {
		return '◎'
	}
	return '○'
}

func siBytes(n uint64) string {
	if n < 1024 {
		return fmt.Sprintf("%vB", n)
	}
	if n < 1024*1024 {
		return fmt.Sprintf("%.1fKiB", float64(n)/1024.0)
	}
	if n < 1024*1024*1024 {
		return fmt.Sprintf("%.1fMiB", float64(n)/(1024.0*1024.0))
	}
	return fmt.Sprintf("%.1fGiB", float64(n)/(1024.0*1024.0*1024.0))
}
