package imageproc

import (
	"image"
	"image/color"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func DrawTextGif(img *image.Paletted, text string, face font.Face, white *image.Uniform, black *image.Uniform) {
	lines := strings.Split(text, "\n")

	lineHeight := face.Metrics().Height.Round()
	totalHeight := lineHeight * len(lines)

	originX := img.Bounds().Dx() / 2 // horizinally center
	originY := int(float64(img.Bounds().Dy()) * 0.77)
	originY -= totalHeight / 2 // vertically center on original originY

	drawer := &GifDrawer{
		Dst:  img,
		Face: face,
	}

	for i, line := range lines {
		adv := font.MeasureString(face, line)
		x := originX - (adv.Round() / 2)
		y := originY + lineHeight*i

		drawer.Dot = fixed.P(x+1, y+1)
		drawer.Src = black
		drawer.DrawString(line)

		drawer.Dot = fixed.P(x, y)
		drawer.Src = white
		drawer.DrawString(line)
	}
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
