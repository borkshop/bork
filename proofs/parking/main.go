package main

import (
	"fmt"
	"image"
	"image/draw"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/input"
	"github.com/borkshop/bork/internal/parking"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func run() (rerr error) {
	term, err := display.NewTerminal(os.Stdout)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := term.Close(); rerr == nil {
			rerr = cerr
		}
	}()

	rand.Seed(time.Now().UnixNano())

	lot := parking.NewLotForBounds(term.Display.Bounds())
	dis := display.New(lot.Bounds())

	commands, mute := input.Channel(os.Stdin)
	defer mute()

	sigwinch := make(chan os.Signal, 1024)
	signal.Notify(sigwinch, syscall.SIGWINCH)

	var buf []byte
	cur := display.Reset
	buf, cur = cur.Hide(buf)
	buf, cur = cur.Clear(buf)

	ticker := time.NewTicker(16 * time.Millisecond)
	last := time.Now()

Loop:
	for {
		lot.Draw(dis, term.Display.Bounds())
		term.Display.Draw(term.Display.Bounds(), dis, image.ZP, draw.Src)

		if err := term.Render(); err != nil {
			return err
		}

		now := time.Now()
		for last.Before(now) {
			lot.Tick()
			last = last.Add(200 * time.Millisecond)
		}

		_, err := os.Stdout.Write(buf)
		buf = buf[0:0]
		if err != nil {
			return err
		}
		buf, cur = cur.Home(buf)

		select {
		case <-sigwinch:
			time.Sleep(10 * time.Millisecond)
			lot = parking.NewLotForBounds(term.Display.Bounds())
			if err := term.UpdateSize(); err != nil {
				return err
			}
		case command := <-commands:
			switch c := command.(type) {
			case rune:
				switch c {
				case 'q':
					break Loop
				}
			}
		case <-ticker.C:
		}
	}

	return nil
}
