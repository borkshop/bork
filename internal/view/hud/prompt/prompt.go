package prompt

import (
	"fmt"
	"image"
	"unicode/utf8"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/moremath"
	"github.com/borkshop/bork/internal/view"
)

// Prompt represents a set of actions that the user may select from. Each
// action has an associated key rune, message, and action. Prompts may be
// chained, i.e. the user is being shown a sub-prompt. Prompts may be left or
// right aligned when rendered.
type Prompt struct {
	prior  *Prompt
	mess   string
	align  view.Align
	action []promptAction
}

// Runner represents an action invoked by a Prompt. RunPrompt is called when a
// Prompt action has been invoked by the user. It gets the prior Prompt value,
// and is expected to return a new Prompt value to display to the user. The
// `required` boolean argument indicates whether, semantically, an action was
// taken or whether further input is needed from the user (presumably
// underneath the returned `next` Prompt).
type Runner interface {
	RunPrompt(prior Prompt) (next Prompt, required bool)
}

// Func is a convenient way to implement Runner arounnd a single function.
type Func func(prior Prompt) (next Prompt, required bool)

// RunPrompt calls the aliased function.
func (f Func) RunPrompt(pr Prompt) (Prompt, bool) { return f(pr) }

type promptAction struct {
	ch   rune
	mess string
	run  Runner
}

const (
	headerOverhead = 1
	headerFmt      = "%s:"
	exitLeftMess   = "0) Exit Menu"
	exitRightMess  = "Exit Menu (0"
	actionOverhead = 3
	actionLeftFmt  = "%s) %s"
	actionRightFmt = "%s (%s"
)

func (act promptAction) renderActionLeft() string {
	return fmt.Sprintf(actionLeftFmt, string(act.ch), act.mess)
}

func (act promptAction) renderActionRight() string {
	return fmt.Sprintf(actionRightFmt, act.mess, string(act.ch))
}

// RenderSize calculates how much space the prompt could use and how much it
// needs. TODO: not yet paginated.
func (pr *Prompt) RenderSize() (wanted, needed image.Point) {
	if len(pr.action) == 0 {
		return
	}

	// header
	if n := utf8.RuneCountInString(pr.mess); n > 0 {
		needed.X = moremath.MaxInt(needed.X, n)
		wanted.X = moremath.MaxInt(wanted.X, n+headerOverhead)
		needed.Y++
		wanted.Y++
	}

	// TODO: vary {needed wanted}.Y for pagination
	for _, act := range pr.action {
		n := utf8.RuneCountInString(act.mess)
		needed.X = moremath.MaxInt(needed.X, n+actionOverhead)
		wanted.X = moremath.MaxInt(wanted.X, n+actionOverhead)
		needed.Y++
		wanted.Y++
	}

	// footer
	needed.X = moremath.MaxInt(needed.X, utf8.RuneCountInString(exitLeftMess))
	wanted.X = moremath.MaxInt(wanted.X, utf8.RuneCountInString(exitLeftMess))
	needed.Y++
	wanted.Y++

	return wanted, needed
}

// Render the prompt within the given space.
func (pr *Prompt) Render(d *display.Display) {
	y := d.Rect.Min.Y
	if pr.mess != "" {
		d.WriteString(0, y, nil, nil, headerFmt, pr.mess)
		y++
	}
	for i := 0; y < d.Rect.Max.Y && i < len(pr.action); y, i = y+1, i+1 {
		act := pr.action[i]
		if pr.align&view.AlignCenter == view.AlignRight {
			d.WriteStringRTL(d.Rect.Max.X-1, y, nil, nil, act.renderActionRight())
		} else {
			d.WriteString(0, y, nil, nil, act.renderActionLeft())
		}
	}
	if pr.align&view.AlignCenter == view.AlignRight {
		d.WriteStringRTL(d.Rect.Max.X-1, y, nil, nil, exitRightMess)
	} else {
		d.WriteString(0, y, nil, nil, exitLeftMess)
	}
	// TODO: paginate
}

// Handle a key event, returning: whether the event was handled, if the prompt
// was canceled, and whether more user input is required (to take semantically
// take an action).
func (pr *Prompt) Handle(cmd interface{}) (handled, canceled, required bool) {
	if len(pr.action) == 0 {
		return false, false, false
	}

	switch c := cmd.(type) {
	case rune:
		switch c {
		case '':
			*pr = pr.Unwind()
			return true, true, false
		case '0':
			*pr = pr.Pop()
			return true, true, false
		}
		for i := range pr.action {
			if c == pr.action[i].ch {
				*pr, required = pr.action[i].run.RunPrompt(*pr)
				return true, false, required
			}
		}
	}

	// TODO: pagination support

	return false, false, true
}

// Run runs the i( >= 0 && <= Len())-th action, returning its next and required
// return values with handled=true if there is an ith-action; the current
// prompt, required=false and handled=false are retured if i is invalid.
func (pr Prompt) Run(i int) (next Prompt, required, handled bool) {
	if i < 0 || i >= len(pr.action) {
		return pr, false, false
	}
	next, required = pr.action[i].run.RunPrompt(pr)
	return next, required, true
}

// SetMess sets the header message.
func (pr *Prompt) SetMess(mess string, args ...interface{}) {
	if len(args) > 0 {
		pr.mess = fmt.Sprintf(mess, args...)
	} else if len(mess) > 0 {
		pr.mess = mess
	} else {
		pr.mess = ""
	}
}

// SetAlign ment for this prompt; only horizontal left/right bits matter.
func (pr *Prompt) SetAlign(align view.Align) {
	pr.align = align
}

// Sub returns a new sub-prompt of the current one with the given header message.
func (pr Prompt) Sub(mess string, args ...interface{}) Prompt {
	sub := Prompt{
		prior: &pr,
		align: pr.align,
	}
	sub.SetMess(mess, args...)
	return sub
}

// Pop returns the parent prompt, if any, or this prompt if it has no parent.
func (pr Prompt) Pop() Prompt {
	if pr.prior != nil {
		return *pr.prior
	}
	return pr
}

// Unwind the prompt, returning the root prompt (which may be the current
// prompt if not a sub-prompt).
func (pr Prompt) Unwind() Prompt {
	for pr.prior != nil {
		pr = *pr.prior
	}
	return pr
}

// Clear prompt state, by unwinding the prompt, clearing its mesage, and
// truncating its actions.
func (pr *Prompt) Clear() {
	*pr = pr.Unwind()
	pr.mess = ""
	pr.action = pr.action[:0]
}

// Len returns how many actions are in this prompt.
func (pr Prompt) Len() int { return len(pr.action) }

// IsRoot returns true only if this prompt is not a sub-prompt.
func (pr Prompt) IsRoot() bool { return pr.prior == nil }

// AddAction adds a new action to the prompt with the given activation rune,
// display message, and action to run; if the rune conflicts with an already
// added action, then the addition fails and false is returned; otherwise true
// is returned.
func (pr *Prompt) AddAction(ch rune, run Runner, mess string, args ...interface{}) bool {
	for i := range pr.action {
		if pr.action[i].ch == ch {
			return false
		}
	}
	if len(args) > 0 {
		mess = fmt.Sprintf(mess, args...)
	}
	pr.action = append(pr.action, promptAction{ch, mess, run})
	return true
}

// RemoveAction removes an action matching the given rune, runner, or message
// (in that order of precedence); zero values will not match. Returns true if
// an action was removed, false otherwise.
func (pr *Prompt) RemoveAction(ch rune, run Runner, mess string) bool {
	for i := range pr.action {
		if (ch != 0 && pr.action[i].ch == ch) ||
			(run != nil && pr.action[i].run == run) ||
			(mess != "" && pr.action[i].mess == mess) {
			pr.action = append(pr.action[:i], pr.action[i+1:]...)
			return true
		}
	}
	return false
}

// SetActionMess updates the message on an existing action, matched by run or
// runner; it returns true only if an action was updated.
func (pr *Prompt) SetActionMess(ch rune, run Runner, mess string, args ...interface{}) bool {
	if len(args) > 0 {
		mess = fmt.Sprintf(mess, args...)
	}
	for i := range pr.action {
		if (ch != 0 && pr.action[i].ch == ch) ||
			(run != nil && pr.action[i].run == run) {
			pr.action[i].mess = mess
			return true
		}
	}
	return false
}

// RunPrompt runs the prompt as a sub-prompt of another; causes Prompt to
// implement Runner, allowing prompts to be added as actions to other prompts.
func (pr Prompt) RunPrompt(prior Prompt) (Prompt, bool) {
	return Prompt{
		prior:  &prior,
		align:  prior.align,
		mess:   pr.mess,
		action: pr.action,
	}, true
}
