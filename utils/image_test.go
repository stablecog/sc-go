package utils

import (
	"image"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateExpandImageSet(t *testing.T) {
	// Load test image "image_testcreateexpandimageset.jpg"
	testImagePath := path.Join(RootDir(), "utils", "image_testcreateexpandimageset.jpg")

	f, err := os.Open(testImagePath)
	assert.NoError(t, err)
	defer f.Close()
	image, _, err := image.Decode(f)
	assert.NoError(t, err)

	assert.Equal(t, 4096, image.Bounds().Dx())
	assert.Equal(t, 4096, image.Bounds().Dy())

	bg, mask := CreateExpandImageSet(image, 0.5, 0.02)
	assert.Equal(t, 4096, bg.Bounds().Dx())
	assert.Equal(t, 4096, bg.Bounds().Dy())
	assert.Equal(t, 4096, mask.Bounds().Dx())
	assert.Equal(t, 4096, mask.Bounds().Dy())

	// Write bg and mask to file system
	// 	bgPath := path.Join(RootDir(), "utils", "image_testcreateexpandimageset_bg.jpg")
	// 	maskPath := path.Join(RootDir(), "utils", "image_testcreateexpandimageset_mask.jpg")

	// 	bGF, err := os.Create(bgPath)
	// 	assert.NoError(t, err)
	// 	defer bGF.Close()
	// 	err = jpeg.Encode(bGF, bg, nil)
	// 	assert.NoError(t, err)
	// 	maskF, err := os.Create(maskPath)
	// 	assert.NoError(t, err)
	// 	defer maskF.Close()
	// 	err = jpeg.Encode(maskF, mask, nil)
	// 	assert.NoError(t, err)
}
