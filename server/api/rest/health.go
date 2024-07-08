package rest

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/cron/discord"
)

var healthTracker *discord.DiscordHealthTracker

// GET health endpoint
func (c *RestAPI) HandleHealth(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{
		"status": "ok",
	})
	render.Status(r, http.StatusOK)
}

func (c *RestAPI) HandleSCWorkerHealth(w http.ResponseWriter, r *http.Request) {
	status := "ok"
	render.JSON(w, r, map[string]string{
		"status": status,
	})
	render.Status(r, http.StatusServiceUnavailable)
}
