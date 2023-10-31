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

var impactFont *sfnt.Font

func init() {
	impactFilename := os.Getenv("IMPACT_FILENAME")
	if impactFilename == "" {
		impactFilename = "impact.ttf"
	}
	impactFont = must(opentype.Parse(must(os.ReadFile(impactFilename))))
}

func makeFace(hinting font.Hinting) (font.Face, error) {
	return opentype.NewFace(impactFont, &opentype.FaceOptions{
		Size: 30, DPI: 72, Hinting: hinting,
	})
}
