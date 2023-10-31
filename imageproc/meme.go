package imageproc

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"

	"golang.org/x/image/font"
)

func MakeMeme(rawImage []byte, text string) ([]byte, error) {
	_, format, err := image.DecodeConfig(bytes.NewReader(rawImage))
	if err != nil {
		return nil, fmt.Errorf("can't decode image: %w", err)
	}

	if format == "gif" {
		face, err := makeFace(font.HintingFull)
		if err != nil {
			return nil, fmt.Errorf("can't make font face: %w", err)
		}

		img, err := gif.DecodeAll(bytes.NewReader(rawImage))
		if err != nil {
			return nil, fmt.Errorf("can't decode gif: %w", err)
		}

		for _, frame := range img.Image {
			white := frame.Palette.Convert(color.RGBA{R: 255, G: 255, B: 255, A: 255})
			black := frame.Palette.Convert(color.Black)
			whiteImg := image.Uniform{C: white}
			blackImg := image.Uniform{C: black}
			DrawTextGif(frame, text, face, &whiteImg, &blackImg)
		}

		outBuf := bytes.NewBuffer(nil)
		err = gif.EncodeAll(outBuf, img)
		if err != nil {
			return nil, fmt.Errorf("can't encode gif: %w", err)
		}
		return outBuf.Bytes(), nil
	} else {
		face, err := makeFace(font.HintingFull)
		if err != nil {
			return nil, fmt.Errorf("can't make font face: %w", err)
		}

		src, _, err := image.Decode(bytes.NewReader(rawImage))
		if err != nil {
			return nil, fmt.Errorf("can't decode image: %w", err)
		}

		dst := image.NewRGBA(src.Bounds())
		draw.Draw(dst, src.Bounds(), src, image.Point{X: 0, Y: 0}, draw.Over)
		DrawText(src, dst, text, face)

		outBuf := bytes.NewBuffer(nil)
		err = jpeg.Encode(outBuf, dst, nil)
		if err != nil {
			return nil, fmt.Errorf("can't encode jpeg: %w", err)
		}
		return outBuf.Bytes(), nil
	}
}
