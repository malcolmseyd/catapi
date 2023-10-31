package main

import (
	"image"
	"image/color"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Point struct {
	X int
	Y int
}

type GifDrawer struct {
	// image to draw on in place
	Dst  *image.Paletted
	Src  *image.Uniform
	Face font.Face
	// draw to the up and left of
	Dot fixed.Point26_6
}

func (d *GifDrawer) DrawString(s string) {
	colorIdx := d.Dst.Palette.Index(d.Src.C)
	prevC := rune(-1)
	for _, c := range s {
		if prevC >= 0 {
			d.Dot.X += d.Face.Kern(prevC, c)
		}
		dr, mask, _, advance, _ := d.Face.Glyph(d.Dot, c)

		b := mask.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				c := mask.At(x, y)
				a := c.(color.Alpha).A
				if a > 128 {
					d.Dst.SetColorIndex(dr.Min.X+x, dr.Min.Y+y, uint8(colorIdx))
				}
			}
		}

		d.Dot.X += advance
		prevC = c
	}
}
