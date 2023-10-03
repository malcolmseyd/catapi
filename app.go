package main

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/gif"
	"image/jpeg"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"path"
	"strings"

	"github.com/gabriel-vasile/mimetype"
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
			catImageIds = append(catImageIds, entry.Name())
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
		log.Println("using image with id", id)
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
		c.Data(200, mimetype.Detect(img).String(), img)
	})
	router.Run(":8080")
}

func getCatImage(id string) ([]byte, error) {
	return os.ReadFile(path.Join(catImagePath, id))
}

func makeMeme(rawImage []byte, text string) ([]byte, error) {
	face, err := makeFace()
	if err != nil {
		return nil, fmt.Errorf("can't make font face: %w", err)
	}

	_, format, err := image.DecodeConfig(bytes.NewReader(rawImage))
	if err != nil {
		return nil, fmt.Errorf("can't decode image: %w", err)
	}

	if format == "gif" {
		img, err := gif.DecodeAll(bytes.NewReader(rawImage))
		if err != nil {
			return nil, fmt.Errorf("can't decode gif: %w", err)
		}

		newFrames := make([]*image.Paletted, 0, len(img.Image))
		for _, frame := range img.Image {
			// less likely to break palette when we just copy over
			dst := new(image.Paletted)
			*dst = *frame
			dst.Pix = make([]uint8, len(frame.Pix))
			copy(dst.Pix, frame.Pix)

			drawText(frame, dst, text, face)
			newFrames = append(newFrames, dst)
		}
		img.Image = newFrames

		outBuf := bytes.NewBuffer(nil)
		err = gif.EncodeAll(outBuf, img)
		if err != nil {
			return nil, fmt.Errorf("can't encode gif: %w", err)
		}
		return outBuf.Bytes(), nil
	} else {
		src, _, err := image.Decode(bytes.NewReader(rawImage))
		if err != nil {
			return nil, fmt.Errorf("can't decode image: %w", err)
		}

		dst := image.NewRGBA(src.Bounds())
		draw.Draw(dst, src.Bounds(), src, image.Point{X: 0, Y: 0}, draw.Over)
		drawText(src, dst, text, face)

		outBuf := bytes.NewBuffer(nil)
		err = jpeg.Encode(outBuf, dst, nil)
		if err != nil {
			return nil, fmt.Errorf("can't encode jpeg: %w", err)
		}
		return outBuf.Bytes(), nil
	}
}

func makeFace() (font.Face, error) {
	return opentype.NewFace(impactFont, &opentype.FaceOptions{
		Size: 30, DPI: 72, Hinting: font.HintingFull,
	})
}

func drawText(src image.Image, dst draw.Image, text string, face font.Face) image.Image {
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

	return dst
}
