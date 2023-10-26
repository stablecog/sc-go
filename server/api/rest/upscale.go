package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

func (c *RestAPI) HandleUpscale(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var upscaleReq requests.CreateUpscaleRequest
	err := json.Unmarshal(reqBody, &upscaleReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	if user.BannedAt != nil {
		remainingCredits, _ := c.Repo.GetNonExpiredCreditTotalForUser(user.ID, nil)
		render.Status(r, http.StatusOK)
		render.JSON(w, r, &responses.TaskQueuedResponse{
			ID:               uuid.NewString(),
			UIId:             upscaleReq.UIId,
			RemainingCredits: remainingCredits,
		})
		return
	}

	voiceover, initSettings, workerErr := c.SCWorker.CreateUpscale(
		enttypes.SourceTypeWebUI,
		r,
		user,
		nil,
		upscaleReq,
	)

	if workerErr != nil {
		errResp := responses.ApiFailedResponse{
			Error: workerErr.Err.Error(),
		}
		if initSettings != nil {
			errResp.Settings = initSettings
		}
		render.Status(r, workerErr.StatusCode)
		render.JSON(w, r, errResp)
		return
	}

	// Return response
	render.Status(r, http.StatusOK)
	render.JSON(w, r, voiceover.QueuedResponse)
}

// POST upscale endpoint
// Handles creating a upscale with API token
func (c *RestAPI) HandleCreateUpscaleToken(w http.ResponseWriter, r *http.Request) {
	var user *ent.User
	if user = c.GetUserIfAuthenticated(w, r); user == nil {
		return
	}
	var apiToken *ent.ApiToken
	if apiToken = c.GetApiToken(w, r); apiToken == nil {
		return
	}
	var upscaleReq *requests.CreateUpscaleRequest

	// See if multipart request
	if strings.Contains(r.Header.Get("Content-Type"), "multipart/form-data") {
		// Image key in S3
		var imageKey string
		// Enforce max upload size
		r.Body = http.MaxBytesReader(w, r.Body, shared.MAX_UPLOAD_SIZE_MB*1024*1024)

		mr, err := r.MultipartReader()
		if err != nil {
			responses.ErrBadRequest(w, r, "parse_error", err.Error())
			return
		}

		for {
			part, err := mr.NextPart()

			// Done reading
			if err == io.EOF {
				break
			}

			if err != nil {
				responses.ErrBadRequest(w, r, "parse_error", err.Error())
				return
			}

			// Image part
			if part.FormName() == "file" {
				buf, err := io.ReadAll(part)
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

				imageKey = fmt.Sprintf("%s/%s.%s", utils.Sha256(user.ID.String()), uuid.New().String(), extension)
				_, err = c.S3.PutObject(&s3.PutObjectInput{
					Bucket:      aws.String(utils.GetEnv().S3Img2ImgBucketName),
					Key:         aws.String(imageKey),
					Body:        bytes.NewReader(buf),
					ContentType: aws.String(contentType),
				})
				if err != nil {
					log.Error("Error uploading object", "err", err)
					responses.ErrInternalServerError(w, r, "An unknown error has occurred")
					return
				}
			} else if part.FormName() == "data" {
				// Parse request body
				var upscaleReqB requests.CreateUpscaleRequest
				reqBody, _ := io.ReadAll(part)
				err := json.Unmarshal(reqBody, &upscaleReqB)
				if err != nil {
					responses.ErrUnableToParseJson(w, r)
					return
				}
				upscaleReq = utils.ToPtr(upscaleReqB)
			}
		}

		if upscaleReq == nil {
			upscaleReq = &requests.CreateUpscaleRequest{
				Type:  utils.ToPtr(requests.UpscaleRequestTypeImage),
				Input: fmt.Sprintf("s3://%s", imageKey),
			}
		} else {
			upscaleReq.Input = fmt.Sprintf("s3://%s", imageKey)
			upscaleReq.Type = utils.ToPtr(requests.UpscaleRequestTypeImage)
		}

	} else {
		// Parse request body
		var upscaleReqB requests.CreateUpscaleRequest
		reqBody, _ := io.ReadAll(r.Body)
		err := json.Unmarshal(reqBody, &upscaleReqB)
		if err != nil {
			responses.ErrUnableToParseJson(w, r)
			return
		}
		upscaleReq = utils.ToPtr(upscaleReqB)
	}

	// Create upscale
	upscale, initSettings, workerErr := c.SCWorker.CreateUpscale(
		enttypes.SourceTypeAPI,
		r,
		user,
		&apiToken.ID,
		*upscaleReq,
	)

	if workerErr != nil {
		errResp := responses.ApiFailedResponse{
			Error: workerErr.Err.Error(),
		}
		if initSettings != nil {
			errResp.Settings = initSettings
		}
		render.Status(r, workerErr.StatusCode)
		render.JSON(w, r, errResp)
		return
	}

	err := c.Repo.UpdateLastSeenAt(user.ID)
	if err != nil {
		log.Warn("Error updating last seen at", "err", err, "user", user.ID.String())
	}

	// Return response
	render.Status(r, http.StatusOK)
	render.JSON(w, r, upscale)
}
