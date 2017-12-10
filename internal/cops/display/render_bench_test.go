package display_test

import (
	"image"
	"image/color"
	"strconv"
	"testing"

	"github.com/borkshop/bork/internal/cops/display"
)

type pcg struct{ state, inc uint64 }

func (pcg *pcg) rand() uint32 {
	const M = 6364136223846793005
	oldstate := pcg.state
	pcg.state = oldstate*M + (pcg.inc | 1)
	xorshifted := ((oldstate >> 18) ^ oldstate) >> 27
	rot := oldstate >> 59
	return uint32(xorshifted>>rot) | uint32(xorshifted<<((-rot)&31))
}

type benchSim struct {
	pcg
	*display.Display
}

func Benchmark_displayDemo(b *testing.B) {
	for _, sz := range []int{4, 8, 16, 32, 64, 128, 256} {
		b.Run(strconv.Itoa(sz), func(b *testing.B) {
			var sim benchSim
			front := display.New(image.Rect(0, 0, sz, sz))
			sim.Display = front
			sim.generate()

			var (
				// 64 chosen for a gross overestimate: never want to
				// re-allocate in the inner loop
				buf = make([]byte, 0, sz*sz*64)
				cur = display.Reset
			)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				sim.Display = front
				sim.iterate()
				buf = buf[:0]
				buf, cur = display.Render(buf, cur, front, display.Model24)
			}
		})
	}
}

var letters [26]string

func init() {
	for i := range letters {
		letters[i] = string('a' + i)
	}
}

func (sim *benchSim) generate() {
	for y := sim.Rect.Min.Y; y < sim.Rect.Max.Y; y++ {
		for x := sim.Rect.Min.X; x < sim.Rect.Max.X; x++ {
			t := letters[sim.rand()%26]
			r := sim.rand()
			f := color.RGBA{R: uint8(r), G: uint8(r >> 8), B: uint8(r >> 16)}
			b := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
			sim.SetRGBA(x, y, t, f, b)
		}
	}
}

func (sim *benchSim) iterate() {
	for y := sim.Rect.Min.Y; y < sim.Rect.Max.Y; y++ {
		for x := sim.Rect.Min.X; x < sim.Rect.Max.X; x++ {
			if n := byte(sim.rand() % 6); n > 2 {
				t, f, b := sim.RGBAAt(x, y)
				t = letters[((t[0]-'a')+(n-3))%26]
				sim.SetRGBA(x, y, t, f, b)
			}
		}
	}
}
