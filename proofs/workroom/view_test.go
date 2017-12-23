package main

import (
	"bytes"
	"image"
	"testing"

	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/views"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_view_Draw(t *testing.T) {
	for _, tc := range []struct {
		name     string
		width    int
		height   int
		f        func() views.Widget
		expected []string
	}{
		{
			name:   "just player @ origin",
			width:  16,
			height: 8,
			f: func() views.Widget {
				world := &worldT{}
				world.init(func(func()) { panic("fake world") })
				world.addChar("player1", '@', tcell.ColorLightGreen, image.ZP)
				view := newView(world)
				return view
			},
			expected: []string{
				"                ",
				"                ",
				"                ",
				"       @        ",
				"                ",
				"                ",
				"                ",
				"                ",
			},
		},
		{
			name:   "just player @ 1,1",
			width:  16,
			height: 8,
			f: func() views.Widget {
				world := &worldT{}
				world.init(func(func()) { panic("fake world") })
				world.addChar("player1", '@', tcell.ColorLightGreen, image.Pt(1, 1))
				view := newView(world)
				return view
			},
			expected: []string{
				"                ",
				"                ",
				"                ",
				"       @        ",
				"                ",
				"                ",
				"                ",
				"                ",
			},
		},
		{
			name:   "just player @ -1,-1",
			width:  16,
			height: 8,
			f: func() views.Widget {
				world := &worldT{}
				world.init(func(func()) { panic("fake world") })
				world.addChar("player1", '@', tcell.ColorLightGreen, image.Pt(-1, -1))
				view := newView(world)
				return view
			},
			expected: []string{
				"                ",
				"                ",
				"                ",
				"       @        ",
				"                ",
				"                ",
				"                ",
				"                ",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scr := tcell.NewSimulationScreen("")
			require.NoError(t, scr.Init())
			scr.SetSize(tc.width, tc.height)
			wid := tc.f()
			wid.SetView(scr)
			wid.Draw()

			scr.Show()
			cells, width, height := scr.GetContents()
			var buf bytes.Buffer
			lines := make([]string, 0, height)
			for i := 0; i < len(cells); i++ {
				if i > 0 && i%width == 0 {
					lines = append(lines, buf.String())
					buf.Reset()
				}
				buf.Write(cells[i].Bytes)
			}
			lines = append(lines, buf.String())

			assert.Equal(t, tc.expected, lines)
		})
	}
}
