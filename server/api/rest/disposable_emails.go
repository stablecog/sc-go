package rest

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
)

type EmailDomainRequest struct {
	Email string `json:"email"`
}

// POST verify email or email domain
func (c *RestAPI) HandleVerifyEmailDomain(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	reqBody, _ := io.ReadAll(r.Body)
	var emailReq EmailDomainRequest
	err := json.Unmarshal(reqBody, &emailReq)
	if err != nil {
		responses.ErrUnableToParseJson(w, r)
		return
	}

	valid := true
	if emailReq.Email == "" || shared.GetCache().IsDisposableEmail(emailReq.Email) {
		valid = false
	}

	if valid {
		email, exists, err := c.Repo.CheckIfEmailExists(emailReq.Email)
		if err != nil {
			log.Errorf("Error checking if email exists: %v", err)
			responses.ErrInternalServerError(w, r, "An unknown error has occured")
			return
		}
		if exists && emailReq.Email != email {
			valid = false
		}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]interface{}{
		"allowed": valid,
	})
}
