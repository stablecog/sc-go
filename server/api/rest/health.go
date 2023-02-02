package rest

import (
	"net/http"

	"github.com/go-chi/render"
)

// GET health endpoint
func (c *RestAPI) HandleHealth(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{
		"status": "ok",
	})
}
