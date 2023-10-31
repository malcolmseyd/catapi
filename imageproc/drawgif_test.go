package imageproc

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"os"
	"strings"
	"testing"

	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type drawTextFunc = func(b *testing.B, frame *image.Paletted, text string, face font.Face)

func benchmarkGifDrawText(b *testing.B, drawText drawTextFunc) {
	face, err := makeFace(font.HintingFull)
	if err != nil {
		b.Skipf("failed to load font face")
	}
	imgBytes, err := os.ReadFile("../img/kopgb6Rqf64osubp")
	if err != nil {
		b.Skipf("failed to open test gif")
	}
	img, err := gif.DecodeAll(bytes.NewReader(imgBytes))
	if err != nil {
		b.Skipf("failed to decode test gif")
	}

	frame := img.Image[0]
	text := strings.TrimSpace(`
:3 :3 :3 :3 :3 :3 :3 :3 :3 :3 :3
hello darkness my old friend
i've come to talk with you again
:3 :3 :3 :3 :3 :3 :3 :3 :3 :3 :3
`)

	b.ResetTimer()
	drawText(b, frame, text, face)
}

func BenchmarkGifDrawTextReference(b *testing.B) {
	f := func(b *testing.B, frame *image.Paletted, text string, face font.Face) {
		for i := 0; i < b.N; i++ {
			DrawText(frame, frame, text, face)
		}
	}
	benchmarkGifDrawText(b, f)
}

func BenchmarkGifDrawTextOptimized(b *testing.B) {
	f := func(b *testing.B, frame *image.Paletted, text string, face font.Face) {
		white := image.NewUniform(frame.Palette.Convert(color.White))
		black := image.NewUniform(frame.Palette.Convert(color.Black))
		for i := 0; i < b.N; i++ {
			DrawTextGif(frame, text, face, white, black)
		}
	}
	benchmarkGifDrawText(b, f)
}

type drawStringer interface {
	DrawString(s string)
	Init(*image.Paletted, *image.Uniform, font.Face)
	SetDot(fixed.Point26_6)
}

func benchmarkDrawString[T drawStringer](d T, b *testing.B, img *image.Paletted, text string, face font.Face) {
	white := image.NewUniform(img.Palette.Convert(color.White))

	originX := 0                                     // left
	originY := int(float64(img.Bounds().Dy()) * 0.5) // middle
	dot := fixed.P(originX, originY)

	d.Init(img, white, face)
	originalImage := make([]uint8, len(img.Pix))
	copy(originalImage, img.Pix)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.SetDot(dot)
		d.DrawString(text)
		copy(img.Pix, originalImage)
	}
}

type fontDrawerTest struct {
	d *font.Drawer
}

func (t *fontDrawerTest) Init(img *image.Paletted, color *image.Uniform, face font.Face) {
	t.d = &font.Drawer{
		Dst:  img,
		Src:  color,
		Face: face,
	}
}

func (t *fontDrawerTest) SetDot(dot fixed.Point26_6) {
	t.d.Dot = dot
}

func (t *fontDrawerTest) DrawString(s string) {
	t.d.DrawString(s)
}

func BenchmarkGifDrawStringReference(b *testing.B) {
	f := func(b *testing.B, frame *image.Paletted, text string, face font.Face) {
		benchmarkDrawString(&fontDrawerTest{}, b, frame, text, face)
	}
	benchmarkGifDrawText(b, f)
}

type gifDrawerTest struct {
	d *GifDrawer
}

func (t *gifDrawerTest) Init(img *image.Paletted, color *image.Uniform, face font.Face) {
	t.d = &GifDrawer{
		Dst:  img,
		Src:  color,
		Face: face,
	}
}

func (t *gifDrawerTest) SetDot(dot fixed.Point26_6) {
	t.d.Dot = dot
}

func (t *gifDrawerTest) DrawString(s string) {
	t.d.DrawString(s)
}

func BenchmarkGifDrawStringOptimized(b *testing.B) {
	f := func(b *testing.B, frame *image.Paletted, text string, face font.Face) {
		benchmarkDrawString(&gifDrawerTest{}, b, frame, text, face)
	}
	benchmarkGifDrawText(b, f)
}
