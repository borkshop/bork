package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/cops/terminal"
	"github.com/borkshop/bork/internal/cops/vtio"
	"github.com/pkg/term/termios"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}

func run() (err error) {
	term := terminal.New(os.Stdin.Fd())
	defer func() {
		err = term.Restore()
	}()
	err = term.SetRaw()
	if err != nil {
		return err
	}

	leader, follower, err := termios.Pty()
	if err != nil {
		return err
	}

	bounds, err := term.Bounds()
	if err != nil {
		return err
	}

	if err := terminal.New(follower.Fd()).SetSize(bounds.Max); err != nil {
		return err
	}

	cmd := exec.Command("htop")
	cmd.Stdin = follower
	cmd.Stdout = follower
	cmd.Stderr = follower
	if err := cmd.Start(); err != nil {
		return err
	}

	vtw := vtio.NewDisplayWriter(bounds)
	go func() {
		_, err = io.Copy(vtw, leader)
	}()

	front, back := display.New2(bounds)

	var buf []byte
	cur := display.Start
	buf, cur = cur.Reset(buf)
	buf, cur = cur.Home(buf)
	buf, cur = cur.Clear(buf)
	buf, cur = cur.Hide(buf)

	// Wait for keypress
	r := make(chan struct{}, 0)
	go func() {
		var rbuf [1]byte
		_, err = os.Stdin.Read(rbuf[0:1])
		close(r)
	}()

DrawLoop:
	for {
		select {
		case <-vtw.C():
			vtw.Draw(front, bounds)
			buf, cur = display.RenderOver(buf, cur, front, back, display.Model24)
			front, back = back, front
			// fmt.Printf("%q\r\n", buf)
			_, err = os.Stdout.Write(buf)
			if err != nil {
				break DrawLoop
			}
			buf = buf[0:0]
		case <-r:
			break DrawLoop
		}
	}

	buf, cur = cur.Reset(buf)
	buf, cur = cur.Home(buf)
	buf, cur = cur.Clear(buf)
	buf, cur = cur.Show(buf)
	_, err = os.Stdout.Write(buf)
	buf = buf[0:0]

	return err
}
