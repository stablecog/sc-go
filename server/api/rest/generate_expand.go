package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/chai2010/webp"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
)

// POST generate expand (zooom-out)
// Adds generate to queue, if authenticated, returns the ID of the generation
func (c *RestAPI) HandleCreateGenerationZoomOutWebUI(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var generateReq requests.CreateGenerationRequest
	err := json.Unmarshal(reqBody, &generateReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	if generateReq.OutputID == nil {
		responses.ErrBadRequest(w, r, "output_id_required", "")
		return
	}

	if user.BannedAt != nil {
		remainingCredits, _ := c.Repo.GetNonExpiredCreditTotalForUser(user.ID, nil)
		render.Status(r, http.StatusOK)
		render.JSON(w, r, &responses.TaskQueuedResponse{
			ID:               uuid.NewString(),
			UIId:             generateReq.UIId,
			RemainingCredits: remainingCredits,
		})
		return
	}

	// Get output
	output, err := c.Repo.GetGenerationOutputForUser(*generateReq.OutputID, user.ID)
	if err != nil {
		if ent.IsNotFound(err) {
			responses.ErrNotFound(w, r, "output_not_found")
			return
		}
		log.Error("Error getting generation output", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Download image
	imageUrl := utils.GetURLFromImagePath(output.ImagePath)
	//Get the response bytes from the url
	response, err := http.Get(imageUrl)
	if err != nil {
		log.Error("Error downloading image output", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}
	defer response.Body.Close()

	// Extract extension from image
	extension := filepath.Ext(output.ImagePath)
	var image image.Image
	var contentType string
	switch extension {
	case ".jpg":
		image, err = jpeg.Decode(response.Body)
		if err != nil {
			log.Error("Error decoding image", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			return
		}
		contentType = "image/jpeg"
	case ".webp":
		image, err = webp.Decode(response.Body)
		if err != nil {
			log.Error("Error decoding image", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			return
		}
		contentType = "image/webp"
	case ".png":
		image, err = png.Decode(response.Body)
		if err != nil {
			log.Error("Error decoding image", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			return
		}
		contentType = "image/png"
	default:
		log.Error("Unsupported image format", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Create mask and image input
	bg, mask := utils.CreateExpandImageSet(image, 0.5, 0.02)
	var bgBuf bytes.Buffer
	var maskBuf bytes.Buffer
	switch extension {
	case ".jpg":
		jpeg.Encode(&bgBuf, bg, nil)
		jpeg.Encode(&maskBuf, mask, nil)
	case ".webp":
		webp.Encode(&bgBuf, bg, nil)
		webp.Encode(&maskBuf, mask, nil)
	case ".png":
		png.Encode(&bgBuf, bg)
		png.Encode(&maskBuf, mask)
	default:
		log.Error("Unsupported image format", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occured")
		return
	}

	// Upload both images to img2img bucket
	uidHash := utils.Sha256(user.ID.String())

	bgObjKey := fmt.Sprintf("%s/%s%s", uidHash, uuid.New().String(), extension)
	maskObjKey := fmt.Sprintf("%s/%s%s", uidHash, uuid.New().String(), extension)

	// IN parallel
	var wg sync.WaitGroup
	wg.Add(2)

	// Create a channel to receive errors from Goroutines
	errCh := make(chan error, 2)
	defer close(errCh)

	// Use Goroutines to run the PutObject requests concurrently
	go func() {
		defer wg.Done()
		err := func() error {
			_, err = c.S3.PutObject(&s3.PutObjectInput{
				Bucket:      aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
				Key:         aws.String(bgObjKey),
				Body:        bytes.NewReader(bgBuf.Bytes()),
				ContentType: aws.String(contentType),
			})
			return err
		}()
		errCh <- err
	}()

	go func() {
		defer wg.Done()
		err := func() error {
			_, err = c.S3.PutObject(&s3.PutObjectInput{
				Bucket:      aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
				Key:         aws.String(maskObjKey),
				Body:        bytes.NewReader(bgBuf.Bytes()),
				ContentType: aws.String(contentType),
			})
			return err
		}()
		errCh <- err
	}()

	// Wait for both Goroutines to finish
	wg.Wait()

	// Check for errors in the channel
	for err := range errCh {
		if err != nil {
			log.Error("Error uploading object", "err", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occurred")
			return
		}
	}

	// Get signed URLs fro each object
	req, _ := c.S3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
		Key:    aws.String(bgObjKey),
	})
	bgUrlStr, err := req.Presign(5 * time.Minute)
	if err != nil {
		log.Error("Error signing init image URL", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	req, _ = c.S3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
		Key:    aws.String(maskObjKey),
	})
	maskUrlStr, err := req.Presign(5 * time.Minute)
	if err != nil {
		log.Error("Error signing mask image URL", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"bg_url":   bgUrlStr,
		"mask_url": maskUrlStr,
	})
}
