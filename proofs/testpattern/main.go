package main

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/borkshop/bork/internal/cops/display"
	"github.com/borkshop/bork/internal/input"
)

func colorGB(x, y int) (f, b color.RGBA) {
	f.A = 0xff
	b.A = 0xff
	f.G += uint8(y)
	b.B += uint8(x)
	return f, b
}

func colorEights(x, y int) (f, b color.RGBA) {
	return display.Colors[y%8], display.Colors[x%8]
}

type xyColorFunc func(x, y int) (f, b color.RGBA)

func fgOnly(cf xyColorFunc) xyColorFunc {
	return func(x, y int) (f, b color.RGBA) {
		f, _ = cf(x, y)
		b = display.Colors[0]
		return f, b
	}
}

func bgOnly(cf xyColorFunc) xyColorFunc {
	return func(x, y int) (f, b color.RGBA) {
		_, b = cf(x, y)
		f = display.Colors[7]
		return f, b
	}
}

func run() (rerr error) {
	f, err := os.Create("debug.log")
	if err != nil {
		return err
	}
	log.SetOutput(f)

	term, err := display.NewTerminal(os.Stdout)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := term.Close(); rerr == nil {
			rerr = cerr
		}
	}()

	commands, mute := input.Channel(os.Stdin)
	defer mute()

	sigwinch := make(chan os.Signal)
	signal.Notify(sigwinch, syscall.SIGWINCH)

	var colorFunc xyColorFunc = colorGB

	for {
		sz := term.Display.Rect.Size()
		for y := 0; y < sz.Y; y++ {
			for x := 0; x < sz.X; x++ {
				t := strconv.Itoa(y % 10)
				f, b := colorFunc(x, y)
				term.Display.SetRGBA(x, y, t, f, b)
				// term.Display.Set(x, y, t, nil, nil)
			}
		}

		if err := term.Render(); err != nil {
			return err
		}

		select {
		case <-sigwinch:
			if err := term.UpdateSize(); err != nil {
				return err
			}
			log.Printf("display resized to %v", term.Bounds())

		case command := <-commands:
			switch c := command.(type) {
			case rune:
				switch c {
				case 'q':
					colorFunc = colorEights
					log.Printf("using colorFunc = colorEights")
				case 'w':
					colorFunc = colorGB
					log.Printf("using colorFunc = colorGB")
				case 'a':
					colorFunc = fgOnly(colorEights)
					log.Printf("using colorFunc = fgOnly(colorEights)")
				case 's':
					colorFunc = fgOnly(colorGB)
					log.Printf("using colorFunc = fgOnly(colorGB)")
				case 'z':
					colorFunc = bgOnly(colorEights)
					log.Printf("using colorFunc = bgOnly(colorEights)")
				case 'x':
					colorFunc = bgOnly(colorGB)
					log.Printf("using colorFunc = bgOnly(colorGB)")

				case '1':
					term.ColorModel = display.Model0
					log.Printf("using ColorModel = Model0")
				case '2':
					term.ColorModel = display.Model3
					log.Printf("using ColorModel = Model3")
				case '3':
					term.ColorModel = display.Model4
					log.Printf("using ColorModel = Model4")
				case '4':
					term.ColorModel = display.Model8
					log.Printf("using ColorModel = Model8")
				case '5':
					term.ColorModel = display.Model24
					log.Printf("using ColorModel = Model24")
				case '6':
					term.ColorModel = display.ModelCompat24
					log.Printf("using ColorModel = ModelCompat24")

				case '', '':
					return nil
				}
			}
		}

	}
}

func main() {
	if err := run(); err != nil {
		fmt.Printf("%v\n", err)
	}
}
