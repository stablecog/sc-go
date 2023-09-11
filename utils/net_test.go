package utils

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stablecog/sc-go/shared"
	"github.com/stretchr/testify/assert"
)

// A 1x1 JPEG image
const TestJPEG = "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEAYABgAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAABAAEDASIAAhEBAxEB/8QAHwAAAQUBAQEBAQEAAAAAAAAAAAECAwQFBgcICQoL/8QAtRAAAgEDAwIEAwUFBAQAAAF9AQIDAAQRBRIhMUEGE1FhByJxFDKBkaEII0KxwRVS0fAkM2JyggkKFhcYGRolJicoKSo0NTY3ODk6Q0RFRkdISUpTVFVWV1hZWmNkZWZnaGlqc3R1dnd4eXqDhIWGh4iJipKTlJWWl5iZmqKjpKWmp6ipqrKztLW2t7i5usLDxMXGx8jJytLT1NXW19jZ2uHi4+Tl5ufo6erx8vP09fb3+Pn6/8QAHwEAAwEBAQEBAQEBAQAAAAAAAAECAwQFBgcICQoL/8QAtREAAgECBAQDBAcFBAQAAQJ3AAECAxEEBSExBhJBUQdhcRMiMoEIFEKRobHBCSMzUvAVYnLRChYkNOEl8RcYGRomJygpKjU2Nzg5OkNERUZHSElKU1RVVldYWVpjZGVmZ2hpanN0dXZ3eHl6goOEhYaHiImKkpOUlZaXmJmaoqOkpaanqKmqsrO0tba3uLm6wsPExcbHyMnK0tPU1dbX2Nna4uPk5ebn6Onq8vP09fb3+Pn6/9oADAMBAAIRAxEAPwD3+iiigD//2Q=="

// A 1x1 PNG image
const TestPNG = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABAQMAAAAl21bKAAAAA1BMVEUAAACnej3aAAAAAXRSTlMAQObYZgAAAApJREFUCNdjYAAAAAIAAeIhvDMAAAAASUVORK5CYII="

// A 1x1 WebP image
const TestWebP = "data:image/webp;base64,UklGRkAAAABXRUJQVlA4WAoAAAAQAAAAAAAAAAAAQUxQSAIAAAAAAFZQOCAYAAAAMAEAnQEqAQABAAIANCWkAANwAP77/VAA"

func TestGetImageSizeFromUrl(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("HEAD", "http://localhost:123456/image.jpeg",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "OK")
			resp.Header.Add("Content-Length", "43")
			return resp, nil
		},
	)

	bytes, err := GetImageSizeFromUrl("http://localhost:123456/image.jpeg")
	assert.Nil(t, err)
	assert.Equal(t, int64(43), bytes)
}

func TestGetImageWidthHeightFromUrlFailsIfTooLarge(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("HEAD", "http://localhost:123456/image.jpeg",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "OK")
			resp.Header.Add("Content-Length", strconv.Itoa(shared.MAX_UPSCALE_IMAGE_SIZE+1))
			return resp, nil
		},
	)
	httpmock.RegisterResponder("GET", "http://localhost:123456/image.jpeg",
		func(req *http.Request) (*http.Response, error) {
			i := strings.Index(TestJPEG, ",")
			decoded, err := base64.StdEncoding.DecodeString(TestJPEG[i+1:])
			if err != nil {
				return nil, err
			}

			resp := httpmock.NewBytesResponse(200, decoded)
			resp.Header.Add("Content-Type", "image/jpeg")
			return resp, nil
		},
	)

	_, _, err := GetImageWidthHeightFromUrl("http://localhost:123456/image.jpeg", "", shared.MAX_UPSCALE_IMAGE_SIZE)
	assert.NotNil(t, err)
	assert.Equal(t, "Image too large", err.Error())
}

// Test when somebody does something nasty with content-length header
func TestGetImageWidthHeightFromUrlJPEGSpoofContentLength(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("HEAD", "http://localhost:123456/image.jpeg",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "OK")
			// Content-Length header is 1 byte
			resp.Header.Add("Content-Length", "1")
			return resp, nil
		},
	)
	httpmock.RegisterResponder("GET", "http://localhost:123456/image.jpeg",
		func(req *http.Request) (*http.Response, error) {
			i := strings.Index(TestJPEG, ",")
			decoded, err := base64.StdEncoding.DecodeString(TestJPEG[i+1:])
			if err != nil {
				return nil, err
			}

			resp := httpmock.NewBytesResponse(200, decoded)
			resp.Header.Add("Content-Type", "image/jpeg")
			return resp, nil
		},
	)

	// Content-Length Tells us 1 byte
	// We don't want more than 2 bytes (that should be ok)
	// This should still fail since the actual body is more than 2 byte
	_, _, err := GetImageWidthHeightFromUrl("http://localhost:123456/image.jpeg", "", 2)
	assert.NotNil(t, err)
	assert.Equal(t, "unexpected EOF", err.Error())
}

func TestGetImageWidthHeightFromUrlJPEG(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("HEAD", "http://localhost:123456/image.jpeg",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "OK")
			resp.Header.Add("Content-Length", "43")
			return resp, nil
		},
	)
	httpmock.RegisterResponder("GET", "http://localhost:123456/image.jpeg",
		func(req *http.Request) (*http.Response, error) {
			i := strings.Index(TestJPEG, ",")
			decoded, err := base64.StdEncoding.DecodeString(TestJPEG[i+1:])
			if err != nil {
				return nil, err
			}

			resp := httpmock.NewBytesResponse(200, decoded)
			resp.Header.Add("Content-Type", "image/jpeg")
			return resp, nil
		},
	)

	width, height, err := GetImageWidthHeightFromUrl("http://localhost:123456/image.jpeg", "", shared.MAX_UPSCALE_IMAGE_SIZE)
	assert.Nil(t, err)
	assert.Equal(t, int32(1), width)
	assert.Equal(t, int32(1), height)
}

func TestGetImageWidthHeightFromUrlPNG(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("HEAD", "http://localhost:123456/image.png",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "OK")
			resp.Header.Add("Content-Length", "43")
			return resp, nil
		},
	)
	httpmock.RegisterResponder("GET", "http://localhost:123456/image.png",
		func(req *http.Request) (*http.Response, error) {
			i := strings.Index(TestPNG, ",")
			decoded, err := base64.StdEncoding.DecodeString(TestPNG[i+1:])
			if err != nil {
				return nil, err
			}

			resp := httpmock.NewBytesResponse(200, decoded)
			resp.Header.Add("Content-Type", "image/png")
			return resp, nil
		},
	)

	width, height, err := GetImageWidthHeightFromUrl("http://localhost:123456/image.png", "", shared.MAX_UPSCALE_IMAGE_SIZE)
	assert.Nil(t, err)
	assert.Equal(t, int32(1), width)
	assert.Equal(t, int32(1), height)
}

func TestGetImageWidthHeightFromUrlWEBP(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("HEAD", "http://localhost:123456/image.webp",
		func(req *http.Request) (*http.Response, error) {
			resp := httpmock.NewStringResponse(200, "OK")
			resp.Header.Add("Content-Length", "43")
			return resp, nil
		},
	)
	httpmock.RegisterResponder("GET", "http://localhost:123456/image.webp",
		func(req *http.Request) (*http.Response, error) {
			i := strings.Index(TestWebP, ",")
			decoded, err := base64.StdEncoding.DecodeString(TestWebP[i+1:])
			if err != nil {
				return nil, err
			}

			resp := httpmock.NewBytesResponse(200, decoded)
			resp.Header.Add("Content-Type", "image/webp")
			return resp, nil
		},
	)

	width, height, err := GetImageWidthHeightFromUrl("http://localhost:123456/image.webp", "", shared.MAX_UPSCALE_IMAGE_SIZE)
	assert.Nil(t, err)
	assert.Equal(t, int32(1), width)
	assert.Equal(t, int32(1), height)
}
