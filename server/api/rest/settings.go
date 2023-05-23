package rest

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

func (c *RestAPI) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	cache := shared.GetCache()

	var generationModels []responses.SettingsResponseItem
	var upscaleModels []responses.SettingsResponseItem
	var schedulers []responses.SettingsResponseItem

	for _, model := range cache.GenerateModels {
		if model.IsActive && !model.IsHidden {
			generationModels = append(generationModels, responses.SettingsResponseItem{
				ID:      model.ID,
				Name:    model.NameInWorker,
				Default: utils.ToPtr(model.IsDefault),
			})
		}
	}
	for _, model := range cache.UpscaleModels {
		if model.IsActive && !model.IsHidden {
			upscaleModels = append(upscaleModels, responses.SettingsResponseItem{
				ID:      model.ID,
				Name:    model.NameInWorker,
				Default: utils.ToPtr(model.IsDefault),
			})
		}
	}
	for _, scheduler := range cache.Schedulers {
		if scheduler.IsActive && !scheduler.IsHidden {
			schedulers = append(schedulers, responses.SettingsResponseItem{
				ID:      scheduler.ID,
				Name:    scheduler.NameInWorker,
				Default: utils.ToPtr(scheduler.IsDefault),
			})
		}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses.SettingsResponse{
		GenerationModels: generationModels,
		UpscaleModels:    upscaleModels,
		Schedulers:       schedulers,
	})
}
