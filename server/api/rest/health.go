package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/cron/discord"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
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

// Specific for runpod functions
type RunpodHealthResponse struct {
	Jobs struct {
		Completed  int `json:"completed"`
		Failed     int `json:"failed"`
		InProgress int `json:"inProgress"`
		InQueue    int `json:"inQueue"`
		Retried    int `json:"retried"`
	} `json:"jobs"`
	Workers struct {
		Idle      int `json:"idle"`
		Running   int `json:"running"`
		Unhealthy int `json:"unhealthy"`
	} `json:"workers"`
}

type WorkerHealthResponse struct {
	HasUnhealthy bool `json:"has_unhealthy"`
}

type WorkerHealthResponseAll struct {
	Models map[string]WorkerHealthResponse `json:"models"`
}

func (c *RestAPI) HandleRunpodWorkerHealth(w http.ResponseWriter, r *http.Request) {
	// Get optional model name parameter
	model := chi.URLParam(r, "model")
	generationModels := shared.GetCache().GetAllGenerationModels()
	// Discard all models without RunpodEndpoint
	runpodModels := make([]*ent.GenerationModel, 0)
	for _, m := range generationModels {
		if m.RunpodEndpoint != nil && (strings.ToLower(m.NameInWorker) == strings.ToLower(model) || model == "all") {
			runpodModels = append(runpodModels, m)
		}
	}

	// Query all runpod endpoints for health
	healthResponses := make(map[string]bool)
	for _, m := range runpodModels {
		// Get health from API GET request
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/health", *m.RunpodEndpoint), nil)
		if err != nil {
			log.Errorf("http.NewRequest failed runpod healtth: %v", err)
			responses.ErrInternalServerError(w, r, "Error getting runpod health")
			return
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", utils.GetEnv().RunpodApiToken))

		statusResp, err := c.Client.Do(req)
		if err != nil {
			log.Errorf("Error polling runpod status: %v", err)
			responses.ErrInternalServerError(w, r, "Error getting runpod health")
			return
		}
		defer statusResp.Body.Close()
		// Unmarshal
		var healthResponse RunpodHealthResponse
		if err := json.NewDecoder(statusResp.Body).Decode(&healthResponse); err != nil {
			log.Errorf("Error decoding runpod health response: %v", err)
			responses.ErrInternalServerError(w, r, "Error getting runpod health/json decode")
			return
		}

		healthResponses[m.NameInWorker] = healthResponse.Workers.Unhealthy > 0
	}

	// Return health responses
	if model == "all" {
		// Show WorkerHealthResponseAll
		workerHealthResponseAll := WorkerHealthResponseAll{
			Models: make(map[string]WorkerHealthResponse),
		}
		for _, m := range runpodModels {
			healthResponse, ok := healthResponses[m.NameInWorker]
			if !ok {
				continue
			}
			workerHealthResponseAll.Models[m.NameInWorker] = WorkerHealthResponse{
				HasUnhealthy: healthResponse,
			}
		}
		if len(workerHealthResponseAll.Models) == 0 {
			responses.ErrNotFound(w, r, "No models found")
			return
		}
		render.JSON(w, r, workerHealthResponseAll)
		return
	}

	// Show WorkerHealthResponse
	hasUnhealthy, ok := healthResponses[model]
	if !ok {
		responses.ErrNotFound(w, r, "Model not found")
		return
	}
	workerHealthResponse := WorkerHealthResponse{
		HasUnhealthy: hasUnhealthy,
	}
	render.JSON(w, r, workerHealthResponse)

}
