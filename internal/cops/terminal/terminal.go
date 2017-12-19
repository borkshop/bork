// Package terminal provides an idiomatic Go interface for reading, writing,
// and restoring terminal capabilities.
package terminal

import (
	"image"
	"syscall"
	"unsafe"

	"github.com/pkg/term/termios"
)

// Terminal models a virtual terminal's current and former capabilities, so
// they can be easily altered and restored.
type Terminal struct {
	fd       uintptr
	old, now syscall.Termios
}

// New returns a Terminal for the given file descriptor, capable of restoring
// that terminal to its current state.
//
// FIXME New() *T; value can be Make() T or T.Init()
func New(fd uintptr) Terminal {
	t := Terminal{fd: fd}
	termios.Tcgetattr(fd, &t.old)
	t.now = t.old
	return t
}

func (t Terminal) set() error {
	return termios.Tcsetattr(t.fd, termios.TCSANOW, &t.now)
}

// Restore resets the terminal capabilities to their original values,
// at time of construction.
func (t *Terminal) Restore() error {
	return termios.Tcsetattr(t.fd, termios.TCSANOW, &t.old)
}

// SetNoEcho suppresses input to output echoing, so printable characters typed
// into the terminal are not implicitly written back out.
func (t Terminal) SetNoEcho() error {
	t.now.Lflag &^= syscall.ECHO
	return t.set()
}

// SetRaw makes a terminal suitable for full-screen terminal user interfaces,
// eliminating keyboard shortcuts for job control, echo, line buffering, and
// escape key debouncing.
func (t Terminal) SetRaw() error {
	termios.Cfmakeraw(&t.now)
	return t.set()
}

// Bounds returns the terminal dimensions as an "image".Rectangle, suitable for
// constructing a virtual display.
func (t Terminal) Bounds() (image.Rectangle, error) {
	return bounds(t.fd)
}

// Size returns the width and height of the terminal as an "image".Point.
func (t Terminal) Size() (image.Point, error) {
	return size(t.fd)
}

// SetSize alters the dimensions of the virtual terminal.
func (t Terminal) SetSize(size image.Point) error {
	return setSize(t.fd, size)
}

func bounds(fd uintptr) (image.Rectangle, error) {
	size, err := size(fd)
	if err != nil {
		return image.Rectangle{}, err
	}
	return image.Rect(0, 0, size.X, size.Y), nil
}

type dimensions struct {
	rows    uint16
	cols    uint16
	xpixels uint16
	ypixels uint16
}

func size(fd uintptr) (size image.Point, err error) {
	var dim dimensions
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		fd, syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&dim)))
	if errno != 0 {
		return image.Point{}, errno
	}
	return image.Pt(int(dim.cols), int(dim.rows)), nil
}

func setSize(fd uintptr, size image.Point) error {
	dim := dimensions{uint16(size.Y), uint16(size.X), 0, 0}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		fd, syscall.TIOCSWINSZ, uintptr(unsafe.Pointer(&dim)))
	if errno != 0 {
		return errno
	}
	return nil
}
