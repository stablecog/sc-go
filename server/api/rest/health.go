package rest

import (
	"net/http"
	"strconv"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/cron/discord"
	"github.com/stablecog/sc-go/shared"
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
	healthStatusStr, _ := c.Redis.Client.Get(c.Redis.Ctx, shared.REDIS_SC_WORKER_HEALTH_KEY).Result()

	if healthStatusStr != "" {
		healthStatusInt, _ := strconv.Atoi(healthStatusStr)
		retrievedStatus := discord.HEALTH_STATUS(healthStatusInt)
		if retrievedStatus != discord.HEALTHY {
			status = "unhealthy"
		}
	}

	render.JSON(w, r, map[string]string{
		"status": status,
	})
	render.Status(r, http.StatusServiceUnavailable)
}
