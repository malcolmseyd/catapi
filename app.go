package main

import (
	"bytes"
	"encoding/json"
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
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

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

		isGithubBot := strings.Contains(c.Request.UserAgent(), "github-camo")
		if isGithubBot {
			c.Header("Cache-Control", "no-cache")
		}

		c.Data(200, mimetype.Detect(img).String(), img)

		if isGithubBot {
			time.Sleep(time.Millisecond * 200)
			purgeSelf()
		}
	})
	router.Run(":8080")
}

func getCatImage(id string) ([]byte, error) {
	sanitizedId := path.Join("/", id)
	return os.ReadFile(path.Join(catImagePath, sanitizedId))
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

		for _, frame := range img.Image {
			white := frame.Palette.Convert(color.RGBA{R: 255, G: 255, B: 255, A: 255})
			black := frame.Palette.Convert(color.Black)
			// optimizePaletted(frame, white, black)
			whiteImg := image.Uniform{C: white}
			blackImg := image.Uniform{C: black}
			drawGif(frame, text, face, &whiteImg, &blackImg)
		}

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

// ASSUMPTION: `black` and `white` are always in the pallete if this is called
// ASSUMPTION: palette is non-empty
func optimizePaletted(img *image.Paletted, white color.Color, black color.Color) {
	// TODO: fix weird white fringing
	// id=Fk4koFgPTqBIf2hE
	// id=kopgb6Rqf64osubp
	wr, wg, wb, wa := white.RGBA()
	br, bg, bb, ba := black.RGBA()
	dst := uint8(len(img.Palette) - 1)
	// translations := [256]uint8{}
	translations := make([]uint8, len(img.Palette))
	// shuffle colors to the right
	for src := dst; src >= 2; src-- {
		r, g, b, a := img.Palette[src].RGBA()
		if r == wr && b == wb && g == wg && a == wa {
			translations[src] = 0
			continue
		}
		if r == br && b == bb && g == bg && a == ba {
			translations[src] = 1
			continue
		}
		img.Palette[dst] = img.Palette[src]
		translations[src] = dst
		dst--
	}
	img.Palette[0] = white
	img.Palette[1] = black
	for i := range img.Pix {
		img.Pix[i] = translations[img.Pix[i]]
	}
}

func drawGif(img *image.Paletted, text string, face font.Face, white *image.Uniform, black *image.Uniform) {
	lines := strings.Split(text, "\n")

	lineHeight := face.Metrics().Height.Round()
	totalHeight := lineHeight * len(lines)

	originX := img.Bounds().Dx() / 2 // horizinally center
	originY := int(float64(img.Bounds().Dy()) * 0.77)
	originY -= totalHeight / 2 // vertically center on original originY

	drawer := &font.Drawer{
		Dst:  img,
		Face: face,
	}

	for i, line := range lines {
		adv := drawer.MeasureString(line)
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

func drawText(src image.Image, dst draw.Image, text string, face font.Face) {
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

func purgeSelf() {
	client := http.Client{Timeout: 10 * time.Second}

	selfURL, err := getSelfURL(&client)
	if err != nil {
		log.Println("failed to get self url:", err)
		return
	}

	purgeReq, err := http.NewRequest("PURGE", selfURL, nil)
	if err != nil {
		log.Println("bad url in purge request:", err)
		return
	}

	_, err = client.Do(purgeReq)
	if err != nil {
		log.Println("failed to purge self:", err)
		return
	}
	log.Println("successfully purged!")
}

var selfURLPattern = regexp.MustCompile(`<img[^>]+alt="cat"[^>]+ src="(https:\/\/camo[^"]*)"[^>]*>`)

func getSelfURL(client *http.Client) (string, error) {
	req, _ := http.NewRequest("GET", "https://github.com/malcolmseyd/malcolmseyd/blob/main/README.md", nil)
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request readme: %w", err)
	}
	defer resp.Body.Close()

	var body struct {
		Payload struct {
			Blob struct {
				RichText string `json:"richText"`
			} `json:"blob"`
		} `json:"payload"`
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return "", fmt.Errorf("failed to decode json body: %w", err)
	}

	matches := selfURLPattern.FindStringSubmatch(body.Payload.Blob.RichText)
	if len(matches) < 2 {
		return "", fmt.Errorf("no match in readme")
	}
	return matches[1], nil
}
