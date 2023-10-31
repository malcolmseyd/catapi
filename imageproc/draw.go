package imageproc

import (
	"image"
	"image/color"
	"image/draw"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

func DrawText(src image.Image, dst draw.Image, text string, face font.Face) {
	lines := strings.Split(text, "\n")

	lineHeight := face.Metrics().Height.Round()
	totalHeight := lineHeight * len(lines)

	originX := src.Bounds().Dx() / 2 // horizinally center
	originY := int(float64(src.Bounds().Dy()) * 0.77)
	originY -= totalHeight / 2 // vertically center on original originY

	whiteImg := image.NewUniform(color.RGBA{255, 255, 255, 255})
	blackImg := image.NewUniform(color.Black)

	drawer := &font.Drawer{
		Dst:  dst,
		Face: face,
	}

	for i, line := range lines {
		adv := drawer.MeasureString(line)
		x := originX - (adv.Round() / 2)
		y := originY + lineHeight*i

		drawer.Dot = fixed.P(x+1, y+1)
		drawer.Src = blackImg
		drawer.DrawString(line)

		drawer.Dot = fixed.P(x, y)
		drawer.Src = whiteImg
		drawer.DrawString(line)
	}
}
