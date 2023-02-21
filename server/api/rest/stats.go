package rest

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/server/responses"
	"k8s.io/klog/v2"
)

func (c *RestAPI) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	res, err := c.Redis.GetGenerateUpscaleCount()
	if err != nil {
		klog.Errorf("Error getting generate upscale count: %v", err)
		responses.ErrInternalServerError(w, r, "Unable to get stats")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}
