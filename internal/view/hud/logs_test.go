package hud_test

import (
	"fmt"
	"image"
	"testing"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/view"
	. "github.com/borkshop/bork/internal/view/hud"
	"github.com/stretchr/testify/assert"
)

func TestHUDLogs(t *testing.T) {
	termSize := image.Point{X: 60, Y: 15}
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
			d := display.New(image.Rectangle{Max: termSize})

			logs.Log(step.log)

			hud := HUD{
				Logs:  logs,
				World: display.New(image.Rect(0, 0, 3, 3)),
			}
			for _, line := range header {
				hud.HeaderF(line)
			}

			hud.Render(d)
			if !assert.Equal(t, step.expected, d.Lines(" ")) {
				wanted, needed := logs.RenderSize()

				t.Logf("logs render needed=%v", needed)
				d := display.New(image.Rectangle{Max: needed})
				logs.Render(d)
				for i, line := range d.Lines("Ø") {
					t.Logf("min[%v] %q", i, line)
				}

				if wanted != needed {
					t.Logf("logs render wanted=%v", wanted)
					d := display.New(image.Rectangle{Max: needed})
					logs.Render(d)
					for i, line := range d.Lines("Ø") {
						t.Logf("max[%v] %q", i, line)
					}
				}

			}
		})
	}
}
