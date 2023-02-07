package utils

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"

	"golang.org/x/image/webp"
)

func GetImageWidthHeightFromUrl(imageUrl string) (width, height int32, err error) {
	//Get the response bytes from the url
	response, err := http.Get(imageUrl)
	if err != nil {
		return 0, 0, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return 0, 0, fmt.Errorf("Received non 200 response code %d", response.StatusCode)
	}

	// Parse content type from response
	var im image.Config
	contentType := response.Header.Get("Content-Type")

	switch contentType {
	case "image/jpeg":
		im, err = jpeg.DecodeConfig(response.Body)
		if err != nil {
			return 0, 0, err
		}
	case "image/png":
		im, err = png.DecodeConfig(response.Body)
		if err != nil {
			return 0, 0, err
		}
	case "image/webp":
		im, err = webp.DecodeConfig(response.Body)
		if err != nil {
			return 0, 0, err
		}
	default:
		return 0, 0, fmt.Errorf("Unsupported content type %s", contentType)
	}

	return int32(im.Width), int32(im.Height), nil
}
