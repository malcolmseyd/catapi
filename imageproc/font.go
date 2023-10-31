package imageproc

import (
	"log"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
)

func must[T any](value T, err error) T {
	if err != nil {
		log.Fatalln("fatal error:", err)
	}
	return value
}

var impactFont *sfnt.Font = must(opentype.Parse(must(os.ReadFile("impact.ttf"))))

func makeFace(hinting font.Hinting) (font.Face, error) {
	return opentype.NewFace(impactFont, &opentype.FaceOptions{
		Size: 30, DPI: 72, Hinting: hinting,
	})
}
