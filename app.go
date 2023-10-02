package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"math/rand"
	"os"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

func must[T any](value T, err error) T {
	if err != nil {
		log.Fatalln("fatal error:", err)
	}
	return value
}

const catImagePath = "img"

var impactFont *sfnt.Font = must(opentype.Parse(must(os.ReadFile("impact.ttf"))))
var catImageIds []string = make([]string, 0)

func init() {
	for _, entry := range must(os.ReadDir(catImagePath)) {
		if entry.Type().IsRegular() {
			catImageIds = append(catImageIds, strings.TrimSuffix(entry.Name(), ".jpg"))
		}
	}
}

func main() {
	router := gin.Default()
	router.GET("/cat", func(c *gin.Context) {
		id := c.Query("id")
		if id == "" {
			id = catImageIds[rand.Int()%len(catImageIds)]
		}
		img, err := getCatImage(id)
		if errors.Is(err, os.ErrNotExist) {
			log.Println("cat image 404:", err)
			c.AbortWithStatus(404)
			return
		} else if err != nil {
			c.AbortWithError(500, err)
			return
		}
		if memeText := c.Query("text"); memeText != "" {
			img, err = makeMeme(img, memeText)
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
		}
		c.Data(200, "image/jpeg", img)
	})
	router.Run(":8080")
}

func getCatImage(id string) ([]byte, error) {
	return os.ReadFile(path.Join(catImagePath, id+".jpg"))
}

func makeMeme(rawImage []byte, text string) ([]byte, error) {
	img, err := jpeg.Decode(bytes.NewReader(rawImage))
	if err != nil {
		return nil, fmt.Errorf("can't decode image: %w", err)
	}

	face, err := makeFace()
	if err != nil {
		return nil, fmt.Errorf("can't make font face: %w", err)
	}

	dst := drawText(img, text, face)

	outBuf := bytes.NewBuffer(nil)
	err = jpeg.Encode(outBuf, dst, nil)
	if err != nil {
		return nil, fmt.Errorf("can't encode image: %w", err)
	}
	return outBuf.Bytes(), nil
}

func makeFace() (font.Face, error) {
	return opentype.NewFace(impactFont, &opentype.FaceOptions{
		Size: 30, DPI: 72, Hinting: font.HintingFull,
	})
}

func drawText(src image.Image, text string, face font.Face) image.Image {
	lines := strings.Split(text, "\n")

	lineHeight := face.Metrics().Height.Round()
	totalHeight := lineHeight * len(lines)

	originX := src.Bounds().Dx() / 2
	originY := int(float64(src.Bounds().Dy()) * 0.77)
	originY -= totalHeight / 2

	whiteImg := image.NewUniform(color.RGBA{255, 255, 255, 255})
	blackImg := image.NewUniform(color.Black)

	dst := image.NewRGBA(src.Bounds())
	draw.Draw(dst, src.Bounds(), src, image.Point{X: 0, Y: 0}, draw.Over)

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

	return dst
}
