package utils

import (
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/disintegration/imaging"
)

// Create mask and background image for zoom-out/expansion
func CreateExpandImageSet(img image.Image, scaleDownBy float64, blurRadiusFraction float64) (image.Image, image.Image) {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	newWidth := int(float64(width) * scaleDownBy)
	newHeight := int(float64(height) * scaleDownBy)

	blurRadiusPixels := int(blurRadiusFraction * math.Min(float64(width), float64(height)))

	imgScaled := imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)

	x0 := (width - newWidth) / 2
	y0 := (height - newHeight) / 2

	bg := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(bg, bg.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	draw.Draw(bg, imgScaled.Bounds().Add(image.Pt(x0, y0)), imgScaled, image.Point{}, draw.Src)

	maskWidth := newWidth - 2*blurRadiusPixels
	maskHeight := newHeight - 2*blurRadiusPixels

	mask := image.NewRGBA(image.Rect(0, 0, maskWidth, maskHeight))
	draw.Draw(mask, mask.Bounds(), image.NewUniform(color.Black), image.Point{}, draw.Src)

	maskBg := imaging.New(width, height, color.NRGBA{0, 0, 0, 0})
	draw.Draw(maskBg, maskBg.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)

	draw.Draw(maskBg, mask.Bounds().Add(image.Pt(x0+blurRadiusPixels, y0+blurRadiusPixels)), mask, image.Point{}, draw.Src)
	maskBg = imaging.Blur(maskBg, float64(blurRadiusPixels))

	return bg, maskBg
}
