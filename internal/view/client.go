package view

import (
	"errors"
	"io"

	"github.com/borkshop/bork/internal/cops/display"
)

// ErrStop may be returned by a client method to mean "we're done, break run loop".
var ErrStop = errors.New("client stop")

// Client is the interface exposed to the user of View; its various methods are
// called in a loop that provides terminal orchestration.
type Client interface {
	Render(*display.Display) error
	HandleInput(interface{}) error
	Close() error
}

// JustKeepRunning starts a view, and then running newly minted Runables
// provided by the given factory until an error occurs, or the user quits.
// Useful for implementing main.main.
func JustKeepRunning(factory func(v *View) (Client, error)) error {
	var v View
	return v.runWith(func() error {
		for {
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
