package utils

import (
	"bufio"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"strconv"

	"golang.org/x/image/webp"
)

// Retrieves the download size of an image via headers
func GetImageSizeFromUrl(imageUrl string) (bytes int64, err error) {
	resp, err := http.Head(imageUrl)
	if err != nil {
		return 0, err
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("Received non 200 response code %d", resp.StatusCode)
	}

	// the Header "Content-Length" will let us know
	// the total file size to download
	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return 0, err
	}

	return int64(size), nil
}

func GetImageWidthHeightFromUrl(imageUrl string, headUrl string, maxSizeBytes int64) (width, height int32, err error) {
	if headUrl == "" {
		headUrl = imageUrl
	}
	// Make sure the size isn't too large
	size, err := GetImageSizeFromUrl(headUrl)
	if err != nil {
		return 0, 0, err
	}
	if size > maxSizeBytes {
		return 0, 0, fmt.Errorf("Image too large")
	}

	//Get the response bytes from the url
	response, err := http.Get(imageUrl)
	if err != nil {
		return 0, 0, err
	}
	defer response.Body.Close()

	// Limit large files
	buffer := bufio.NewReader(response.Body)
	limitReader := io.LimitReader(buffer, maxSizeBytes)

	if response.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("Received non 200 response code %d", response.StatusCode)
	}

	// Parse content type from response
	var im image.Config
	contentType := response.Header.Get("Content-Type")

	switch contentType {
	case "image/jpeg":
		im, err = jpeg.DecodeConfig(limitReader)
		if err != nil {
			return 0, 0, err
		}
	case "image/png":
		im, err = png.DecodeConfig(limitReader)
		if err != nil {
			return 0, 0, err
		}
	case "image/webp":
		im, err = webp.DecodeConfig(limitReader)
		if err != nil {
			return 0, 0, err
		}
	default:
		return 0, 0, fmt.Errorf("Unsupported content type %s", contentType)
	}

	if im.Width > math.MaxInt32 || im.Height > math.MaxInt32 {
		return 0, 0, fmt.Errorf("Image dimensions too large")
	}

	return int32(im.Width), int32(im.Height), nil
}
