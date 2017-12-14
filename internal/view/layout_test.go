package view_test

import (
	"image"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/borkshop/bork/internal/cops/display"
	. "github.com/borkshop/bork/internal/view"
)

func makeDis(x, y int) func() *display.Display {
	return func() *display.Display {
		return display.New(image.Rect(0, 0, x, y))
	}
}

func TestLayout(t *testing.T) {
	type sa struct {
		x interface{}
		a Align
	}
	for _, tc := range []struct {
		name     string
		init     func() *display.Display
		sas      []sa
		expected []string
	}{
		{
			name: "basic top left",
			init: makeDis(25, 10),
			sas: []sa{
				{"left1", AlignTop | AlignLeft},
				{"left2", AlignTop | AlignLeft},
				{"left3", AlignTop | AlignLeft | AlignHFlush},
			},
			expected: []string{
				"left1 left2              ",
				"left3                    ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
			},
		},

		{
			name: "basic top right",
			init: makeDis(25, 10),
			sas: []sa{
				{"right1", AlignTop | AlignRight},
				{"rrright4", AlignTop | AlignRight},
				{"right2", AlignTop | AlignRight},
				{"right3", AlignTop | AlignRight | AlignHFlush},
			},
			expected: []string{
				"   right2 rrright4 right1",
				"                   right3",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
			},
		},

		{
			name: "basic top left&right",
			init: makeDis(25, 10),
			sas: []sa{
				{"left1", AlignTop | AlignLeft},
				{"left2", AlignTop | AlignLeft},
				{"left3", AlignTop | AlignLeft | AlignHFlush},
				{"right1", AlignTop | AlignRight},
				{"rrright4", AlignTop | AlignRight},
				{"right2", AlignTop | AlignRight},
				{"right3", AlignTop | AlignRight | AlignHFlush},
			},
			expected: []string{
				"left1 left2 right2 right1",
				"left3            rrright4",
				"                   right3",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
			},
		},

		{
			name: "basic top left&right&center",
			init: makeDis(25, 10),
			sas: []sa{
				{"left1", AlignTop | AlignLeft},
				{"left2", AlignTop | AlignLeft},
				{"left3", AlignTop | AlignLeft | AlignHFlush},
				{"right1", AlignTop | AlignRight},
				{"rrright4", AlignTop | AlignRight},
				{"right2", AlignTop | AlignRight},
				{"right3", AlignTop | AlignRight | AlignHFlush},
				{"center1", AlignTop | AlignCenter},
			},
			expected: []string{
				"left1 left2 right2 right1",
				"left3  center1   rrright4",
				"                   right3",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
			},
		},

		{
			name: "basic bottom left",
			init: makeDis(25, 10),
			sas: []sa{
				{"left4", AlignBottom | AlignLeft},
				{"left5", AlignBottom | AlignLeft},
				{"left6", AlignBottom | AlignLeft | AlignHFlush},
			},
			expected: []string{
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"left6                    ",
				"left4 left5              ",
			},
		},

		{
			name: "basic bottom right",
			init: makeDis(25, 10),
			sas: []sa{
				{"right4", AlignBottom | AlignRight},
				{"right5", AlignBottom | AlignRight},
				{"right6", AlignBottom | AlignRight | AlignHFlush},
			},
			expected: []string{
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                   right6",
				"            right5 right4",
			},
		},

		{
			name: "basic bottom left&right",
			init: makeDis(25, 10),
			sas: []sa{
				{"left4", AlignBottom | AlignLeft},
				{"left5", AlignBottom | AlignLeft},
				{"left6", AlignBottom | AlignLeft | AlignHFlush},
				{"right4", AlignBottom | AlignRight},
				{"right5", AlignBottom | AlignRight},
				{"right6", AlignBottom | AlignRight | AlignHFlush},
			},
			expected: []string{
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"left6              right6",
				"left4 left5 right5 right4",
			},
		},

		{
			name: "basic bottom left&right&center",
			init: makeDis(25, 10),
			sas: []sa{
				{"left4", AlignBottom | AlignLeft},
				{"left5", AlignBottom | AlignLeft},
				{"left6", AlignBottom | AlignLeft | AlignHFlush},
				{"right4", AlignBottom | AlignRight},
				{"right5", AlignBottom | AlignRight},
				{"right6", AlignBottom | AlignRight | AlignHFlush},
				{"center2", AlignBottom | AlignCenter},
			},
			expected: []string{
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"left6   center2    right6",
				"left4 left5 right5 right4",
			},
		},

		// TODO: not exactly happy with these middle-ing outcomes

		{
			name: "basic middle left",
			init: makeDis(25, 10),
			sas: []sa{
				{"left7", AlignMiddle | AlignLeft},
				{"left8", AlignMiddle | AlignLeft},
				{"left9", AlignMiddle | AlignLeft | AlignHFlush},
			},
			expected: []string{
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"left7 left8              ",
				"left9                    ",
				"                         ",
				"                         ",
				"                         ",
			},
		},

		{
			name: "basic middle right",
			init: makeDis(25, 10),
			sas: []sa{
				{"right7", AlignMiddle | AlignRight},
				{"right8", AlignMiddle | AlignRight},
				{"right9", AlignMiddle | AlignRight | AlignHFlush},
			},
			expected: []string{
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"            right8 right7",
				"                   right9",
				"                         ",
				"                         ",
				"                         ",
			},
		},

		{
			name: "basic middle left&right",
			init: makeDis(25, 10),
			sas: []sa{
				{"left7", AlignMiddle | AlignLeft},
				{"left8", AlignMiddle | AlignLeft},
				{"left9", AlignMiddle | AlignLeft | AlignHFlush},
				{"right7", AlignMiddle | AlignRight},
				{"right8", AlignMiddle | AlignRight},
				{"right9", AlignMiddle | AlignRight | AlignHFlush},
			},
			expected: []string{
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"left7 left8 right8 right7",
				"left9              right9",
				"                         ",
				"                         ",
				"                         ",
			},
		},

		{
			name: "basic middle left&right&center",
			init: makeDis(25, 10),
			sas: []sa{
				{"left7", AlignMiddle | AlignLeft},
				{"left8", AlignMiddle | AlignLeft},
				{"left9", AlignMiddle | AlignLeft | AlignHFlush},
				{"right7", AlignMiddle | AlignRight},
				{"right8", AlignMiddle | AlignRight},
				{"right9", AlignMiddle | AlignRight | AlignHFlush},
				{"center3", AlignMiddle | AlignCenter},
			},
			expected: []string{
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"                         ",
				"left7 left8 right8 right7",
				"left9   center3    right9",
				"                         ",
				"                         ",
				"                         ",
			},
		},

		{
			name: "basic full-up",
			init: makeDis(25, 10),
			sas: []sa{
				{"left1", AlignTop | AlignLeft},
				{"left2", AlignTop | AlignLeft},
				{"left3", AlignTop | AlignLeft | AlignHFlush},
				{"right1", AlignTop | AlignRight},
				{"rrright4", AlignTop | AlignRight},
				{"right2", AlignTop | AlignRight},
				{"right3", AlignTop | AlignRight | AlignHFlush},
				{"center1", AlignTop | AlignCenter},
				{"left4", AlignBottom | AlignLeft},
				{"left5", AlignBottom | AlignLeft},
				{"left6", AlignBottom | AlignLeft | AlignHFlush},
				{"right4", AlignBottom | AlignRight},
				{"right5", AlignBottom | AlignRight},
				{"right6", AlignBottom | AlignRight | AlignHFlush},
				{"center2", AlignBottom | AlignCenter},
				{"left7", AlignMiddle | AlignLeft},
				{"left8", AlignMiddle | AlignLeft},
				{"left9", AlignMiddle | AlignLeft | AlignHFlush},
				{"right7", AlignMiddle | AlignRight},
				{"right8", AlignMiddle | AlignRight},
				{"right9", AlignMiddle | AlignRight | AlignHFlush},
				{"center3", AlignMiddle | AlignCenter},
			},
			expected: []string{
				"left1 left2 right2 right1",
				"left3  center1   rrright4",
				"                   right3",
				"                         ",
				"                         ",
				"left7 left8 right8 right7",
				"left9   center3    right9",
				"                         ",
				"left6   center2    right6",
				"left4 left5 right5 right4",
			},
		},

		{
			name: "multi-line top left",
			init: makeDis(35, 15),
			sas: []sa{
				{lshaped('a', 1, 2, 3), AlignTop | AlignLeft},
			},
			expected: []string{
				"a                                  ",
				"aa                                 ",
				"aaa                                ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
			},
		},

		{
			name: "multi-line top right",
			init: makeDis(35, 15),
			sas: []sa{
				{rshaped('b', 2, 3, 1), AlignTop | AlignRight},
			},
			expected: []string{
				"                                 bb",
				"                                bbb",
				"                                  b",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
			},
		},

		{
			name: "multi-line bottom left",
			init: makeDis(35, 15),
			sas: []sa{
				{lshaped('c', 3, 2, 1), AlignBottom | AlignLeft},
			},
			expected: []string{
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"ccc                                ",
				"cc                                 ",
				"c                                  ",
			},
		},

		{
			name: "multi-line bottom right",
			init: makeDis(35, 15),
			sas: []sa{
				{rshaped('d', 2, 1, 3), AlignBottom | AlignRight},
			},
			expected: []string{
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                 dd",
				"                                  d",
				"                                ddd",
			},
		},

		{
			name: "multi-line middle center",
			init: makeDis(35, 15),
			sas: []sa{
				{lshaped('e', 3, 5, 8, 5, 3), AlignMiddle | AlignCenter},
			},
			expected: []string{
				"                                   ",
				"                                   ",
				"                                   ",

				"             eee                   ",
				"             eeeee                 ",
				"             eeeeeeee              ",
				"             eeeee                 ",
				"             eee                   ",

				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
			},
		},

		{
			name: "multi-line full-up",
			init: makeDis(35, 15),
			sas: []sa{
				{lshaped('a', 1, 2, 3), AlignTop | AlignLeft},
				{rshaped('b', 2, 3, 1), AlignTop | AlignRight},
				{lshaped('c', 3, 2, 1), AlignBottom | AlignLeft},
				{rshaped('d', 2, 1, 3), AlignBottom | AlignRight},
				{lshaped('e', 3, 5, 8, 5, 3), AlignMiddle | AlignCenter},
			},
			expected: []string{
				"a                                bb",
				"aa                              bbb",
				"aaa                               b",
				"             eee                   ",
				"             eeeee                 ",
				"             eeeeeeee              ",
				"             eeeee                 ",
				"             eee                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"                                   ",
				"ccc                              dd",
				"cc                                d",
				"c                               ddd",
			},
		},

		// TODO: center first, then big left occludes the prior center

		{
			name: "single over wanted",
			init: makeDis(16, 6),
			sas: []sa{
				{overWant(lshaped('a', 3, 2, 1), 2, 0), AlignTop | AlignLeft},
			},
			expected: []string{
				"aaa             ",
				"aa              ",
				"a               ",
				"                ",
				"                ",
				"                ",
			},
		},

		{
			name: "single over needed",
			init: makeDis(16, 6),
			sas: []sa{
				{overNeed(lshaped('a', 3, 2, 1), 2, 0), AlignTop | AlignLeft},
			},
			expected: []string{
				"aaa             ",
				"aa              ",
				"a               ",
				"                ",
				"                ",
				"                ",
			},
		},

		{
			name: "over wanted&needed w/company",
			init: makeDis(16, 6),
			sas: []sa{
				{lshaped('a', 3, 2, 1), AlignTop | AlignLeft},
				{overWant(lshaped('b', 3, 2, 1), 2, 0), AlignTop | AlignLeft},
				{overNeed(lshaped('c', 3, 2, 1), 2, 0), AlignTop | AlignLeft},
			},
			expected: []string{
				"aaa bbb ccc     ",
				"aa  bb  cc      ",
				"a   b   c       ",
				"                ",
				"                ",
				"                ",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lay := Layout{}
			lay.Display = tc.init()
			for _, sa := range tc.sas {
				switch v := sa.x.(type) {
				case Renderable:
					lay.Render(v, sa.a)
				case string:
					lay.Render(RenderString(v), sa.a)
				}
			}
			assert.Equal(t, tc.expected, lay.Display.Lines(" "))
		})
	}
}

type shape struct {
	r  rune
	ns []int
	a  Align
}

func lshaped(r rune, ns ...int) Renderable { return shape{r, ns, AlignLeft} }
func rshaped(r rune, ns ...int) Renderable { return shape{r, ns, AlignRight} }

func (sh shape) RenderSize() (wanted, needed image.Point) {
	needed.Y = len(sh.ns)
	for _, n := range sh.ns {
		if n > needed.X {
			needed.X = n
		}
	}
	return needed, needed
}

func (sh shape) Render(d *display.Display) {
	for y, n := range sh.ns {
		switch s := strings.Repeat(string(sh.r), n); sh.a {
		case AlignRight:
			d.WriteStringRTL(d.Rect.Max.X-1, y, nil, nil, s)

		default: // AlignLeft:
			d.WriteString(0, y, nil, nil, s)
		}
	}
}

func overWant(ren Renderable, x, y int) Renderable { return overSize{ren, image.ZP, image.Pt(x, y)} }
func overNeed(ren Renderable, x, y int) Renderable { return overSize{ren, image.Pt(x, y), image.ZP} }
func overWantNeed(ren Renderable, wanted, needed image.Point) Renderable {
	return overSize{ren, wanted, needed}
}

type overSize struct {
	Renderable
	wanted, needed image.Point
}

func (over overSize) RenderSize() (wanted, needed image.Point) {
	wanted, needed = over.Renderable.RenderSize()
	wanted = wanted.Add(over.wanted)
	needed = needed.Add(over.needed)
	if wanted.X < needed.X {
		wanted.X = needed.X
	}
	if wanted.Y < needed.Y {
		wanted.Y = needed.Y
	}
	return wanted, needed
}
