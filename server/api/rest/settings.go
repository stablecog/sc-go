package rest

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
)

func (c *RestAPI) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	cache := shared.GetCache()

	generationModels := make([]responses.SettingsResponseItem, len(cache.GenerateModels))
	upscaleModels := make([]responses.SettingsResponseItem, len(cache.UpscaleModels))
	schedulers := make([]responses.SettingsResponseItem, len(cache.Schedulers))

	for i, model := range cache.GenerateModels {
		generationModels[i] = responses.SettingsResponseItem{
			ID:   model.ID,
			Name: model.NameInWorker,
		}
	}
	for i, model := range cache.UpscaleModels {
		upscaleModels[i] = responses.SettingsResponseItem{
			ID:   model.ID,
			Name: model.NameInWorker,
		}
	}
	for i, scheduler := range cache.Schedulers {
		schedulers[i] = responses.SettingsResponseItem{
			ID:   scheduler.ID,
			Name: scheduler.NameInWorker,
		}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses.SettingsResponse{
		GenerationModels: generationModels,
		UpscaleModels:    upscaleModels,
		Schedulers:       schedulers,
	})
}
