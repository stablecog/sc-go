package rest

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/responses"
)

func (c *RestAPI) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	res, err := c.Redis.GetStatsCount()
	if err != nil {
		log.Error("Error getting generate upscale count", "err", err)
		responses.ErrInternalServerError(w, r, "Unable to get stats")
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, res)
}
