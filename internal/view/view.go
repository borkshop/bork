package view

import (
	"errors"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/input"
)

const keyBufferSize = 1100

// View implements a terminal user interaction, based around a grid, header,
// and footer. Additionally a log is provided, whose tail is displayed beneath
// the header.
type View struct {
	term *display.Terminal

	in        *os.File
	input     <-chan interface{}
	stopInput func()

	sigwinch chan os.Signal

	renderPending bool
	renderTimer   *time.Timer

	done chan struct{}
}

func (v *View) runWith(f func() error) (rerr error) {
	// TODO termbox opens /dev/tty, why don't we?
	v.done = make(chan struct{})
	err := v.setup(os.Stdin, os.Stdout)
	defer func() {
		if cerr := v.Close(); rerr == nil {
			rerr = cerr
		}
	}()
	if ferr := f(); err == nil {
		err = ferr
	}
	return err
}

var errViewState = errors.New("invalid view state")

func (v *View) setup(in, out *os.File) error {
	if v.input != nil {
		return errViewState
	}

	term, err := display.NewTerminal(out)
	if err != nil {
		return err
	}

	v.term = term

	v.in = in
	v.input, v.stopInput = input.Channel(in,
		input.RecognizeViKeys,
		input.RecognizeShiftedViKeys,
	)
	// TODO other signals like SIGINT, SIGTERM, etc
	v.sigwinch = make(chan os.Signal, 1)
	signal.Notify(v.sigwinch, syscall.SIGWINCH)

	return nil
}

// Close shuts down the view, restoring the terminal to its former self.
func (v *View) Close() error {
	signal.Stop(v.sigwinch)
	if v.stopInput != nil {
		v.stopInput()
		v.stopInput = nil
	}
	v.input = nil

	return v.term.Close()
}

func (v *View) runClient(client Client) (rerr error) {
	defer func() {
		if cerr := client.Close(); rerr == nil || rerr == ErrStop {
			rerr = cerr
		}
	}()

	// TODO observability / introspection / other Nice To Haves?

	v.requestRender()

	// TODO support a ticker to pump animations; needs to be client-afforded,
	// since not every animation tick should be a simulation tick

	for {
		select {
		case <-v.sigwinch:
			if err := v.term.UpdateSize(); err != nil {
				return err
			}
			v.requestRender()

		case <-v.done:
			return io.EOF

		case cmd := <-v.input:
			if err := v.handleInput(cmd, client); err != nil {
				return err
			}

		case <-v.renderTimer.C:
			if err := v.render(client); err != nil {
				return err
			}
		}
	}
}

func (v *View) handleInput(cmd interface{}, client Client) error {
	// TODO handle base commands like interrupt, redraw, etc
	switch c := cmd.(type) {
	case rune:
		switch c {
		case '':
			// NOTE immediate, not v.requestRender()
			return v.render(client)
		case 'q', 'Q', '':
			close(v.done)
			return nil
		}
	}
	err := client.HandleInput(cmd)
	if err == nil {
		v.requestRender()
	}
	return err
}

func (v *View) requestRender() {
	if !v.renderPending {
		v.renderPending = true
		if v.renderTimer == nil {
			v.renderTimer = time.NewTimer(10 * time.Millisecond)
		} else {
			v.renderTimer.Reset(10 * time.Millisecond)
		}
	}
}

func (v *View) render(client Client) error {
	if v.renderTimer.Stop() {
		<-v.renderTimer.C
	}
	v.renderPending = false
	v.term.Display.Clear(v.term.Display.Rect)
	if err := client.Render(v.term.Display); err != nil {
		return err
	}
	return v.term.Render()
}
