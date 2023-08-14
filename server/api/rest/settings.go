package rest

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/stablecog/sc-go/server/responses"
	"github.com/stablecog/sc-go/shared"
	"github.com/stablecog/sc-go/utils"
)

// Generation defaults and models
func (c *RestAPI) HandleGetGenerationDefaults(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses.ImageGenerationSettingsResponse{
		ModelId:        shared.GetCache().GetDefaultGenerationModel().ID,
		SchedulerId:    shared.GetCache().GetDefaultScheduler().ID,
		NumOutputs:     shared.DEFAULT_GENERATE_NUM_OUTPUTS,
		GuidanceScale:  shared.DEFAULT_GENERATE_GUIDANCE_SCALE,
		InferenceSteps: shared.DEFAULT_GENERATE_INFERENCE_STEPS,
		Width:          shared.GetCache().GetDefaultGenerationModel().DefaultWidth,
		Height:         shared.GetCache().GetDefaultGenerationModel().DefaultHeight,
		PromptStrength: utils.ToPtr(shared.DEFAULT_GENERATE_PROMPT_STRENGTH),
	})
}

func (c *RestAPI) HandleGetGenerationModels(w http.ResponseWriter, r *http.Request) {
	var generationModels []responses.SettingsResponseItem

	for _, model := range shared.GetCache().GenerationModels() {
		if model.IsActive && !model.IsHidden {
			m := responses.SettingsResponseItem{
				ID:            model.ID,
				Name:          model.NameInWorker,
				IsDefault:     utils.ToPtr(model.IsDefault),
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

	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses.ImageModelsResponse{
		Models: generationModels,
	})
}

// Upscale defaults and models
func (c *RestAPI) HandleGetUpscaleDefaults(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses.ImageUpscaleSettingsResponse{
		ModelId: shared.GetCache().GetDefaultUpscaleModel().ID,
	})
}

func (c *RestAPI) HandleGetUpscaleModels(w http.ResponseWriter, r *http.Request) {
	var upscaleModels []responses.SettingsResponseItem

	for _, model := range shared.GetCache().UpscaleModels() {
		if model.IsActive && !model.IsHidden {
			upscaleModels = append(upscaleModels, responses.SettingsResponseItem{
				ID:        model.ID,
				Name:      model.NameInWorker,
				IsDefault: utils.ToPtr(model.IsDefault),
			})
		}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses.ImageModelsResponse{
		Models: upscaleModels,
	})
}

// Voiceover defaults and models
func (c *RestAPI) HandleGetVoiceoverDefaults(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses.VoiceoverSettingsResponse{
		ModelId:       shared.GetCache().GetDefaultVoiceoverModel().ID,
		SpeakerId:     shared.GetCache().GetDefaultVoiceoverSpeaker().ID,
		Temperature:   shared.DEFAULT_VOICEOVER_TEMPERATURE,
		DenoiseAudio:  shared.DEFAULT_VOICEOVER_DENOISE_AUDIO,
		RemoveSilence: shared.DEFAULT_VOICEOVER_REMOVE_SILENCE,
	})
}

func (c *RestAPI) HandleGetVoiceoverModels(w http.ResponseWriter, r *http.Request) {
	var voiceoverModels []responses.SettingsResponseItem

	for _, model := range shared.GetCache().VoiceoverModels() {
		if model.IsActive && !model.IsHidden {
			speakers := shared.GetCache().GetVoiceoverSpeakersForModel(model.ID)
			availableSpeakers := make([]responses.AvailableSpeaker, len(speakers))
			for i, speaker := range speakers {
				availableSpeakers[i] = responses.AvailableSpeaker{
					ID:   speaker.ID,
					Name: speaker.NameInWorker,
				}
			}
			voiceoverModels = append(voiceoverModels, responses.SettingsResponseItem{
				ID:                model.ID,
				Name:              model.NameInWorker,
				IsDefault:         utils.ToPtr(model.IsDefault),
				AvailableSpeakers: availableSpeakers,
			})
		}
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, responses.ImageModelsResponse{
		Models: voiceoverModels,
	})
}
