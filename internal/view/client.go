package view

import (
	"errors"
	"io"

	termbox "github.com/nsf/termbox-go"
)

// ErrStop may be returned by a client method to mean "we're done, break run loop".
var ErrStop = errors.New("client stop")

// Client is the interface exposed to the user of View; its various methods are
// called in a loop that provides terminal orchestration.
type Client interface {
	Render(Grid) error
	HandleKey(KeyEvent) error
	Close() error
}

// KeyEvent represents a terminal key event.
type KeyEvent struct {
	Mod termbox.Modifier
	Key termbox.Key
	Ch  rune
}

// JustKeepRunning starts a view, and then running newly minted Runables
// provided by the given factory until an error occurs, or the user quits.
// Useful for implementing main.main.
func JustKeepRunning(factory func(v *View) (Client, error)) error {
	var v View
	return v.runWith(func() error {
		for v.polling {
			client, err := factory(&v)
			if err != nil {
				return err
			}
			err = v.runClient(client)
			if err == ErrStop {
				continue
			}
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Run a Client under this view, returning any error from the run (may be
// caused by the client, or view).
func (v *View) Run(client Client) error {
	return v.runWith(func() error {
		err := v.runClient(client)
		if err == ErrStop {
			return nil
		}
		return err
	})
}
