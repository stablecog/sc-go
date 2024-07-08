package rest

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/cron/discord"
	"github.com/stablecog/sc-go/cron/jobs"
)

// GET health endpoint
func (c *RestAPI) HandleHealth(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, map[string]string{
		"status": "ok",
	})
	render.Status(r, http.StatusOK)
}

func (c *RestAPI) HandleWorkerHealth(w http.ResponseWriter, r *http.Request) {
	healthStatus := jobs.GetWorkerHealthStatus()
	status := "ok"
	if healthStatus != discord.HEALTHY {
		status = "unhealthy"
	}
	render.JSON(w, r, map[string]string{
		"status": status,
	})
	render.Status(r, http.StatusServiceUnavailable)
}
