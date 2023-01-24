package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/render"
	"github.com/stablecog/go-apps/server/models"
	"github.com/stablecog/go-apps/server/models/constants"
	"k8s.io/klog/v2"
)

// POST generate endpoint
// Adds generate to queue, if authenticated, returns the ID of the generation
func (c *HttpController) PostGenerate(w http.ResponseWriter, r *http.Request) {
	// See if authenticated
	userID, authenticated := r.Context().Value("user_id").(string)
	// This should always be true because of the auth middleware, but check it anyway
	if !authenticated || userID == "" {
		models.ErrUnauthorized(w, r)
		return
	}

	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var generateReq models.GenerateRequestBody
	err := json.Unmarshal(reqBody, &generateReq)
	if err != nil {
		models.ErrUnableToParseJson(w, r)
		return
	}

	// Validate request body
	if generateReq.Height > constants.MaxGenerateHeight {
		models.ErrBadRequest(w, r, fmt.Sprintf("Height is too large, max is: %d", constants.MaxGenerateHeight))
		return
	}
	if generateReq.Width > constants.MaxGenerateWidth {
		models.ErrBadRequest(w, r, fmt.Sprintf("Width is too large, max is: %d", constants.MaxGenerateWidth))
		return
	}
	if generateReq.Width*generateReq.Height*generateReq.NumInferenceSteps >= constants.MaxProPixelSteps {
		klog.Infof(
			"Pick fewer inference steps or smaller dimensions: %d - %d - %d",
			generateReq.Width,
			generateReq.Height,
			generateReq.NumInferenceSteps,
		)
		models.ErrBadRequest(w, r, "Pick fewer inference steps or smaller dimensions")
		return
	}
	if generateReq.Seed < 0 {
		rand.Seed(time.Now().Unix())
		generateReq.Seed = rand.Intn(math.MaxInt32)
	}

	// Get country code
	//countryCode := utils.GetCountryCode(r)

	isProUser, err := c.Repo.IsProUser(userID)
	if err != nil {
		klog.Errorf("Error checking if user is pro: %v", err)
		models.ErrInternalServerError(w, r, "Error retrieving user")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{
		"status": "ok",
	})
}
