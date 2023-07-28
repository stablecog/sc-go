package utils

import (
	"image"
	"image/color"
	"image/draw"
)

// Create mask and background image for zoom-out/expansion
func CreateExpandImageSet(img image.Image, scaleDownBy float64, blurRadiusFraction float64) (image.Image, image.Image) {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	// Create the mask
	mask := image.NewRGBA(image.Rect(0, 0, width, height))
	whiteRect := image.Rect(0, 0, width, height/2)
	blackRect := image.Rect(0, height/2, width, height)

	draw.Draw(mask, whiteRect, image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(mask, blackRect, image.NewUniform(color.Black), image.Point{}, draw.Src)

	return img, mask
}
