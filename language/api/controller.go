package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/language/shared"
	"github.com/stablecog/sc-go/server/responses"
)

type Controller struct {
	LanguageDetector *shared.LanguageDetector
}

// Health check endpoint
func (c *Controller) HandleHealth(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{
		"status": "ok",
	})
	render.Status(r, http.StatusOK)
}

// Handle flores
func (c *Controller) HandleGetTargetFloresCode(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var floresReq FloresRequest
	err := json.Unmarshal(reqBody, &floresReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	if len(floresReq.Inputs) == 0 {
		render.Status(r, http.StatusOK)
		render.JSON(w, r, FloresResponse{
			Outputs: []string{},
		})
		return
	}

	// Get target flores code
	outputCodes := c.LanguageDetector.GetFloresCodes(floresReq.Inputs)

	render.Status(r, http.StatusOK)
	render.JSON(w, r, FloresResponse{
		Outputs: outputCodes,
	})
}

type FloresRequest struct {
	Inputs []string `json:"inputs"`
}

type FloresResponse struct {
	Outputs []string `json:"outputs"`
}
