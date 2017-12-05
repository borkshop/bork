package view

import (
	"fmt"
	"io"
	"sync"
	"time"

	termbox "github.com/nsf/termbox-go"

	"github.com/borkshop/bork/internal/point"
)

const keyBufferSize = 1100

// View implements a terminal user interaction, based around a grid, header,
// and footer. Additionally a log is provided, whose tail is displayed beneath
// the header.
type View struct {
	polling bool
	pollErr error
	keys    chan KeyEvent
	redraw  chan struct{}
	done    chan struct{}

	sizeLock sync.Mutex
	size     point.Point
	termGrid Grid
}

func (v *View) runWith(f func() error) (rerr error) {
	if v.polling {
		panic("invalid view state")
	}

	v.polling = true

	if err := termbox.Init(); err != nil {
		return err
	}

	priorInputMode := termbox.SetInputMode(termbox.InputCurrent)
	defer termbox.SetInputMode(priorInputMode)
	termbox.SetInputMode(termbox.InputEsc)

	priorOutputMode := termbox.SetOutputMode(termbox.OutputCurrent)
	defer termbox.SetOutputMode(priorOutputMode)
	termbox.SetOutputMode(termbox.Output256)

	v.pollErr = nil
	v.redraw = make(chan struct{}, 1)
	v.keys = make(chan KeyEvent, keyBufferSize)
	v.done = make(chan struct{})
	v.size = termboxSize()

	go v.pollEvents()
	defer func() {
		go termbox.Interrupt()
		v.polling = false
		if v.done != nil {
			<-v.done
		}
		if rerr == nil {
			rerr = v.pollErr
		}
	}()

	return f()
}

func (v *View) runClient(client Client) (rerr error) {
	defer func() {
		if cerr := client.Close(); rerr == nil || rerr == ErrStop {
			rerr = cerr
		}
	}()

	raise(v.redraw)

	// TODO: observability / introspection / other Nice To Haves?

	for {
		select {

		case <-v.done:
			return io.EOF

		case <-v.redraw:

		case k := <-v.keys:
			if err := client.HandleKey(k); err != nil {
				return err
			}

		}

		if err := v.render(client); err != nil {
			return err
		}
	}
}

func (v *View) render(client Client) error {
	v.sizeLock.Lock()
	defer v.sizeLock.Unlock()

	if !point.Zero.Less(v.size) {
		v.size = termboxSize()
	}
	if !point.Zero.Less(v.size) {
		return fmt.Errorf("bogus terminal size %v", v.size)
	}

	if !v.termGrid.Size.Equal(v.size) {
		v.termGrid.Resize(v.size)
	}
	for i := range v.termGrid.Data {
		v.termGrid.Data[i] = termbox.Cell{}
	}

	if err := client.Render(v.termGrid); err != nil {
		return err
	}

	if err := termbox.Clear(termbox.ColorDefault, termbox.ColorDefault); err != nil {
		return fmt.Errorf("termbox.Clear failed: %v", err)
	}
	copy(termbox.CellBuffer(), v.termGrid.Data)
	if err := termbox.Flush(); err != nil {
		return fmt.Errorf("termbox.Flush failed: %v", err)
	}
	return nil
}

func (v *View) pollEvents() {
	defer termbox.Close()
	defer close(v.done)

	v.pollErr = func() error {
		for v.polling {
			switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key {
				case termbox.KeyCtrlC:
					return nil
				case termbox.KeyCtrlL:
					raise(v.redraw)
					continue
				}
				switch ev.Ch {
				case 'q', 'Q':
					return nil
				}
				select {
				case v.keys <- KeyEvent{ev.Mod, ev.Key, ev.Ch}:
				case <-time.After(10 * time.Millisecond):
				}

			case termbox.EventResize:
				// TODO: would rather defer this into the client running code
				// to coalesce resize events; that seems to be the intent of
				// termbox.Clear, but we've already built our grid by the time
				// we clear that... basically a simpler/lower layer than
				// termbox would be really nice...
				v.sizeLock.Lock()
				v.size.X = ev.Width
				v.size.Y = ev.Height
				v.sizeLock.Unlock()
				raise(v.redraw)

			case termbox.EventError:
				return ev.Err
			}
		}
		return nil
	}()
}

func termboxSize() point.Point {
	w, h := termbox.Size()
	return point.Point{X: w, Y: h}
}

func raise(ch chan<- struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}
