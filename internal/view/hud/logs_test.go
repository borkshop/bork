package hud_test

import (
	"fmt"
	"testing"

	"github.com/borkshop/bork/internal/point"
	"github.com/borkshop/bork/internal/view"
	. "github.com/borkshop/bork/internal/view/hud"
	"github.com/stretchr/testify/assert"
)

func TestHUDLogs(t *testing.T) {
	termSize := point.Point{X: 60, Y: 15}
	header := []string{
		">banner",
	}

	var logs Logs
	logs.Init(1000)
	logs.Align = view.AlignLeft | view.AlignTop | view.AlignHFlush

	for _, step := range []struct {
		log      string
		expected []string
	}{
		{
			log: "what we have here is failure to communicate.",
			expected: []string{
				"what we have here is failure to communicate.          banner",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
			},
		},

		{
			log: "how do we fix it?",
			expected: []string{
				"what we have here is failure to communicate.          banner",
				"how do we fix it?                                           ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
			},
		},

		{
			log: "no one's really sure...",
			expected: []string{
				"what we have here is failure to communicate.          banner",
				"how do we fix it?                                           ",
				"no one's really sure...                                     ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
			},
		},

		{
			log: "or at least that's what we think they said",
			expected: []string{
				"what we have here is failure to communicate.          banner",
				"how do we fix it?                                           ",
				"no one's really sure...                                     ",
				"or at least that's what we think they said                  ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
			},
		},

		{
			log: "(at least that's what we live-tweeted)",
			expected: []string{
				"what we have here is failure to communicate.          banner",
				"how do we fix it?                                           ",
				"no one's really sure...                                     ",
				"or at least that's what we think they said                  ",
				"(at least that's what we live-tweeted)                      ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
			},
		},

		{
			log: "maybe we should ask them again",
			expected: []string{
				"what we have here is failure to communicate.          banner",
				"how do we fix it?                                           ",
				"no one's really sure...                                     ",
				"or at least that's what we think they said                  ",
				"(at least that's what we live-tweeted)                      ",
				"maybe we should ask them again                              ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
			},
		},

		{
			log: "only time will tell",
			expected: []string{
				"what we have here is failure to communicate.          banner",
				"how do we fix it?                                           ",
				"no one's really sure...                                     ",
				"or at least that's what we think they said                  ",
				"(at least that's what we live-tweeted)                      ",
				"maybe we should ask them again                              ",
				"only time will tell                                         ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
			},
		},

		{
			log: "now we're just filling space",
			expected: []string{
				"what we have here is failure to communicate.          banner",
				"how do we fix it?                                           ",
				"no one's really sure...                                     ",
				"or at least that's what we think they said                  ",
				"(at least that's what we live-tweeted)                      ",
				"maybe we should ask them again                              ",
				"only time will tell                                         ",
				"now we're just filling space                                ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
			},
		},

		{
			log: "yup",
			expected: []string{
				"what we have here is failure to communicate.          banner",
				"how do we fix it?                                           ",
				"no one's really sure...                                     ",
				"or at least that's what we think they said                  ",
				"(at least that's what we live-tweeted)                      ",
				"maybe we should ask them again                              ",
				"only time will tell                                         ",
				"now we're just filling space                                ",
				"yup                                                         ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
			},
		},

		{
			log: "uh-huh",
			expected: []string{
				"what we have here is failure to communicate.          banner",
				"how do we fix it?                                           ",
				"no one's really sure...                                     ",
				"or at least that's what we think they said                  ",
				"(at least that's what we live-tweeted)                      ",
				"maybe we should ask them again                              ",
				"only time will tell                                         ",
				"now we're just filling space                                ",
				"yup                                                         ",
				"uh-huh                                                      ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
			},
		},

		{
			log: "... and then?",
			expected: []string{
				"how do we fix it?                                     banner",
				"no one's really sure...                                     ",
				"or at least that's what we think they said                  ",
				"(at least that's what we live-tweeted)                      ",
				"maybe we should ask them again                              ",
				"only time will tell                                         ",
				"now we're just filling space                                ",
				"yup                                                         ",
				"uh-huh                                                      ",
				"... and then?                                               ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
				"                                                            ",
			},
		},
	} {
		t.Run(fmt.Sprintf("after logging %q", step.log), func(t *testing.T) {
			termGrid := view.MakeGrid(termSize)

			logs.Log(step.log)

			hud := HUD{
				Logs: logs,
			}
			for _, line := range header {
				hud.HeaderF(line)
			}

			hud.Render(termGrid)
			if !assert.Equal(t, step.expected, termGrid.Lines(' ')) {
				wanted, needed := logs.RenderSize()

				t.Logf("logs render needed=%v", needed)
				g := view.MakeGrid(needed)
				logs.Render(g)
				for i, line := range g.Lines('Ø') {
					t.Logf("min[%v] %q", i, line)
				}

				if wanted != needed {
					t.Logf("logs render wanted=%v", wanted)
					g := view.MakeGrid(needed)
					logs.Render(g)
					for i, line := range g.Lines('Ø') {
						t.Logf("max[%v] %q", i, line)
					}
				}

			}
		})
	}
}
