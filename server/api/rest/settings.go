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
			m := responses.SettingsResponseItem{
				ID:            model.ID,
				Name:          model.NameInWorker,
				Default:       utils.ToPtr(model.IsDefault),
				DefaultWidth:  utils.ToPtr(model.DefaultWidth),
				DefaultHeight: utils.ToPtr(model.DefaultHeight),
			}
			m.AvailableSchedulers = make([]responses.AvailableScheduler, len(model.Edges.Schedulers))
			for i, scheduler := range model.Edges.Schedulers {
				m.AvailableSchedulers[i] = responses.AvailableScheduler{
					ID:   scheduler.ID,
					Name: scheduler.NameInWorker,
				}
			}
			generationModels = append(generationModels, m)
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
		GenerationDefaults: responses.ImageGenerationSettingsResponse{
			Model:          shared.GetCache().GetDefaultGenerationModel().ID,
			Scheduler:      shared.GetCache().GetDefaultScheduler().ID,
			NumOutputs:     shared.DEFAULT_GENERATE_NUM_OUTPUTS,
			GuidanceScale:  shared.DEFAULT_GENERATE_GUIDANCE_SCALE,
			InferenceSteps: shared.DEFAULT_GENERATE_INFERENCE_STEPS,
			Width:          shared.GetCache().GetDefaultGenerationModel().DefaultWidth,
			Height:         shared.GetCache().GetDefaultGenerationModel().DefaultHeight,
		},
	})
}
