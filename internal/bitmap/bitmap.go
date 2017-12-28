package bitmap

import "image"

// Bitmap is a compact bitmap image with a two-color palette.
type Bitmap struct {
	Bytes  []byte
	Stride int
	Rect   image.Rectangle
}

// New returns a bitmap with the given rectangle and two-color palette.
func New(r image.Rectangle) *Bitmap {
	w, h := r.Dx(), r.Dy()
	stride := (w + 7) / 8
	count := stride * h
	return &Bitmap{
		Bytes:  make([]byte, count),
		Stride: stride,
		Rect:   r,
	}
}

// Bounds returns the bounds of the bitmap
func (b *Bitmap) Bounds() image.Rectangle {
	return b.Rect
}

// maskIndex returns the bit mask and byte index for the bit at a given point.
func (b *Bitmap) maskIndex(x, y int) (byte, int) {
	index := y * b.Stride
	byteOffset := index + x>>3
	bitOffset := (index + x) & 07
	return 1 << uint(bitOffset), byteOffset
}

// At returns whether the bit is set at a point.
func (b *Bitmap) At(x, y int) bool {
	if !image.Pt(x, y).In(b.Rect) {
		return false
	}

	mask, index := b.maskIndex(x, y)
	bits := b.Bytes[index]
	return bits&mask != 0
}

// Set sets or resets the bit at a point.
func (b *Bitmap) Set(x, y int, bit bool) {
	if !image.Pt(x, y).In(b.Rect) {
		return
	}

	mask, index := b.maskIndex(x, y)
	if bit {
		b.Bytes[index] |= mask
	} else {
		b.Bytes[index] &^= mask
	}
}
