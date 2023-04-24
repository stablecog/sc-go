package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/utils"
)

// Max upload size allowed for img2img
const MAX_UPLOAD_SIZE_MB = 10

// The max number of files a user can have in the bucket under their folder at any time
const MAX_FILES_PER_USER = 100

type Controller struct {
	Repo  *repository.Repository
	Redis *database.RedisWrapper
	S3    *s3.S3
}

// Health check endpoint
func (c *Controller) HandleHealth(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{
		"status": "ok",
	})
	render.Status(r, http.StatusOK)
}

// Handle upload
func (c *Controller) HandleUpload(w http.ResponseWriter, r *http.Request) {
	// See if authenticated
	userIDStr, authenticated := r.Context().Value("user_id").(string)
	// This should always be true because of the auth middleware, but check it anyway
	if !authenticated || userIDStr == "" {
		responses.ErrUnauthorized(w, r)
		return
	}
	// Ensure valid uuid
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		responses.ErrUnauthorized(w, r)
		return
	}

	// See if banned
	banned, err := c.Repo.IsBanned(userID)
	if err != nil {
		log.Error("Error checking if user is banned", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}
	if banned {
		responses.ErrForbidden(w, r)
		return
	}

	// Hash user ID to protect leaking it
	uidHash := utils.Sha256(userID.String())

	// Get total credits
	credits, err := c.Repo.GetNonExpiredCreditTotalForUser(userID, nil)
	if err != nil {
		log.Error("Error getting credits", "err", err)
		responses.ErrInternalServerError(w, r, "Error getting credits")
		return
	}
	if credits <= 0 {
		responses.ErrInsufficientCredits(w, r)
		return
	}

	// Enforce max upload size
	r.Body = http.MaxBytesReader(w, r.Body, MAX_UPLOAD_SIZE_MB*1024*1024)
	if err := r.ParseMultipartForm(MAX_UPLOAD_SIZE_MB * 1024 * 1024); err != nil {
		log.Error("Error parsing multipart form", "err", err)
		responses.ErrBadRequest(w, r, "file_too_large", fmt.Sprintf("Cannot exceed %dMb", MAX_UPLOAD_SIZE_MB))
		return
	}
	defer r.Body.Close()

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Error("Error in FormFile", "err", err)
		responses.ErrBadRequest(w, r, "invalid_file", "Invalid file")
		return
	}

	defer file.Close()

	// Prune files in this users folder, if they have more than MAX_FILES_PER_USER
	out, err := c.S3.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
		Prefix: aws.String(fmt.Sprintf("%s/", uidHash)),
	})

	if err != nil {
		log.Error("Error listing objects", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	// If there are more than MAX_FILES_PER_USER, delete the oldest ones
	if len(out.Contents) > MAX_FILES_PER_USER {
		// Sort by last modified
		sort.Slice(out.Contents, func(i, j int) bool {
			return out.Contents[i].LastModified.After(*out.Contents[j].LastModified)
		})
		// Delete oldest
		for _, content := range out.Contents[MAX_FILES_PER_USER:] {
			_, err := c.S3.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
				Key:    content.Key,
			})
			if err != nil {
				log.Error("Error deleting object", "err", err)
				responses.ErrInternalServerError(w, r, "An unknown error has occurred")
				return
			}
		}
	}

	// Upload the file to S3
	buf, err := io.ReadAll(file)
	if err != nil {
		log.Error("Error reading body", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}
	// Detect content-type
	contentType := http.DetectContentType(buf)
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/webp" {
		responses.ErrBadRequest(w, r, "invalid_content_type", "Content type must be image/jpeg, image/png, or image/webp")
		return
	}
	// Get extension from content-type
	var extension string
	switch contentType {
	case "image/jpeg":
		extension = "jpg"
	case "image/png":
		extension = "png"
	case "image/webp":
		extension = "webp"
	}

	objKey := fmt.Sprintf("%s/%s.%s", uidHash, uuid.New().String(), extension)
	_, err = c.S3.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(os.Getenv("S3_IMG2IMG_BUCKET_NAME")),
		Key:         aws.String(objKey),
		Body:        bytes.NewReader(buf),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		log.Error("Error uploading object", "err", err)
		responses.ErrInternalServerError(w, r, "An unknown error has occurred")
		return
	}

	render.JSON(w, r, map[string]string{
		"object": fmt.Sprintf("s3://%s", objKey),
	})
	render.Status(r, http.StatusOK)
}
