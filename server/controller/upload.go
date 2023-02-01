package controller

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/shared"
	"github.com/stablecog/go-apps/utils"
	"golang.org/x/exp/slices"
	"k8s.io/klog/v2"
)

// HTTP PUT for uploading files to S3 bucket
// Invoked by the cog
func (c *HttpController) HandleUploadFile(w http.ResponseWriter, r *http.Request) {
	// Parse request
	paths := strings.Split(r.URL.Path, "/")
	if len(paths) != 5 {
		render.Status(r, http.StatusBadRequest)
		return
	}
	imageFullName := paths[4]
	imageNameParts := strings.Split(imageFullName, ".")
	if len(imageNameParts) != 2 {
		render.Status(r, http.StatusBadRequest)
		return
	}
	imageExtension := shared.ImageExtension(imageNameParts[1])
	if imageExtension == shared.JPG {
		imageExtension = shared.JPEG
	}
	if !slices.Contains(shared.ALLOWS_IMAGE_EXTENSIONS_UPLOAD, imageExtension) {
		render.Status(r, http.StatusBadRequest)
		return
	}

	// Get S3 data
	s3Data := utils.GetS3Data()

	id := uuid.New()
	imageKey := fmt.Sprintf("%s/%s.%s", s3Data.BucketPrivateOutputQueueFolder, id, imageExtension)
	log.Printf("S3 bucket private presign image key: %v", imageKey)
	contentType := fmt.Sprintf("image/%s", imageExtension)
	presignResult, err := c.S3PresignClient.PresignPutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      aws.String(s3Data.BucketPrivate),
		Key:         aws.String(imageKey),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		klog.Errorf("S3 bucket private presign error: %v", err)
		render.Status(r, http.StatusInternalServerError)
		return
	}
	klog.Infof("S3 bucket private presign result: %v", presignResult)
	http.Redirect(w, r, presignResult.URL, http.StatusTemporaryRedirect)
}
