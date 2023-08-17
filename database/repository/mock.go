package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credittype"
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
)

// Mock user IDs
const MOCK_ADMIN_UUID = "00000000-0000-0000-0000-000000000000"
const MOCK_NORMAL_UUID = "00000000-0000-0000-0000-000000000001"
const MOCK_NO_CREDITS_UUID = "00000000-0000-0000-0000-000000000002"
const MOCK_ALT_UUID = "00000000-0000-0000-0000-000000000003"

// Mock generation model IDs and scheduler IDs
const MOCK_GENERATION_MODEL_ID = "b972a2b8-f39e-4ee3-a670-05e3acdd821c"
const MOCK_SCHEDULER_ID = "b4dff6e9-91a7-449b-b1a7-c25000e3ccd0"

// Mock upscale
const MOCK_UPSCALE_MODEL_ID = "b972a2b8-f39e-4ee3-a670-05e3acdd821e"

// Mock voiceover
const MOCK_VOICEOVER_MODEL_ID = "b972a2b8-f39e-4ee3-a670-05e3acdd821f"
const MOCK_VOICEOVER_SPEAKER_ID = "b4dff6e9-91a7-449b-b1a7-c25000e3ccd1"

// Just creates some mock data for our tests
func (repo *Repository) CreateMockData(ctx context.Context) error {
	// Drop all data
	repo.DB.User.Delete().ExecX(ctx)
	repo.DB.Role.Delete().ExecX(ctx)
	repo.DB.Credit.Delete().ExecX(ctx)
	repo.DB.CreditType.Delete().ExecX(ctx)
	repo.DB.GenerationModel.Delete().ExecX(ctx)
	repo.DB.Scheduler.Delete().ExecX(ctx)
	repo.DB.UpscaleModel.Delete().ExecX(ctx)
	repo.DB.Generation.Delete().ExecX(ctx)
	repo.DB.GenerationOutput.Delete().ExecX(ctx)
	repo.DB.Upscale.Delete().ExecX(ctx)
	repo.DB.Voiceover.Delete().ExecX(ctx)
	repo.DB.VoiceoverOutput.Delete().ExecX(ctx)
	repo.DB.VoiceoverModel.Delete().ExecX(ctx)
	repo.DB.VoiceoverSpeaker.Delete().ExecX(ctx)

	// Create a credit type
	stripeProductId := "prod_123"
	creditType, err := repo.CreateCreditType("mock", 100, nil, &stripeProductId, credittype.TypeSubscription)
	if err != nil {
		return err
	}

	// ! Mock users
	// Create a user
	u, err := repo.DB.User.Create().SetEmail("mockadmin@stablecog.com").SetID(uuid.MustParse(MOCK_ADMIN_UUID)).SetStripeCustomerID("1").SetUsername("1").Save(ctx)
	if err != nil {
		return err
	}
	// Give user admin role
	sAdminRole, err := repo.DB.Role.Create().SetName("SUPER_ADMIN").Save(ctx)
	if err != nil {
		return err
	}
	repo.DB.User.Update().Where(user.IDEQ(u.ID)).AddRoles(sAdminRole).ExecX(ctx)
	// Give user credits
	_, err = repo.AddCreditsIfEligible(creditType, u.ID, time.Now().AddDate(0, 0, 30), "", nil)
	if err != nil {
		return err
	}
	err = repo.SetActiveProductID(u.ID, stripeProductId, nil)
	if err != nil {
		return err
	}
	// Create another non-admin user
	u, err = repo.DB.User.Create().SetEmail("mockuser@stablecog.com").SetID(uuid.MustParse(MOCK_NORMAL_UUID)).SetStripeCustomerID("2").SetUsername("2").Save(ctx)
	if err != nil {
		return err
	}
	// Give user credits
	_, err = repo.AddCreditsIfEligible(creditType, u.ID, time.Now().AddDate(0, 0, 30), "", nil)
	if err != nil {
		return err
	}
	err = repo.SetActiveProductID(u.ID, stripeProductId, nil)
	if err != nil {
		return err
	}

	// Create another non-admin user
	u, err = repo.DB.User.Create().SetEmail("mockaltuser@stablecog.com").SetID(uuid.MustParse(MOCK_ALT_UUID)).SetStripeCustomerID("3").SetUsername("3").Save(ctx)
	if err != nil {
		return err
	}
	// Give user credits
	_, err = repo.AddCreditsIfEligible(creditType, u.ID, time.Now().AddDate(0, 0, 30), "", nil)
	if err != nil {
		return err
	}
	err = repo.SetActiveProductID(u.ID, stripeProductId, nil)
	if err != nil {
		return err
	}
	// Give user more credits
	_, err = repo.DB.Credit.Create().SetCreditTypeID(creditType.ID).SetUserID(u.ID).SetRemainingAmount(1234).SetExpiresAt(time.Now().AddDate(1000, 0, 0)).Save(ctx)
	if err != nil {
		return err
	}

	// Create another non-admin user with no credits
	u, err = repo.DB.User.Create().SetEmail("mocknocredituser@stablecog.com").SetID(uuid.MustParse(MOCK_NO_CREDITS_UUID)).SetStripeCustomerID("4").SetLastSignInAt(time.Now()).SetUsername("4").Save(ctx)
	if err != nil {
		return err
	}

	// ! Mock generation models
	// Create a generation model for the free user
	_, err = repo.DB.GenerationModel.Create().SetID(uuid.MustParse(MOCK_GENERATION_MODEL_ID)).SetNameInWorker("mockfreemodel").Save(ctx)
	if err != nil {
		return err
	}
	// ! Mock upscale models
	_, err = repo.DB.UpscaleModel.Create().SetID(uuid.MustParse(MOCK_UPSCALE_MODEL_ID)).SetNameInWorker("mockupscalemodel").Save(ctx)
	if err != nil {
		return err
	}

	// ! Mock schedulers
	// Create a scheduler for the free user
	_, err = repo.DB.Scheduler.Create().SetID(uuid.MustParse(MOCK_SCHEDULER_ID)).SetNameInWorker("mockfreescheduler").Save(ctx)
	if err != nil {
		return err
	}

	// ! Mock voiceover model and speaker
	_, err = repo.DB.VoiceoverModel.Create().SetID(uuid.MustParse(MOCK_VOICEOVER_MODEL_ID)).SetNameInWorker("mockvoiceovermodel").Save(ctx)
	if err != nil {
		return err
	}

	// ! Mock voiceover speaker
	_, err = repo.DB.VoiceoverSpeaker.Create().SetID(uuid.MustParse(MOCK_VOICEOVER_SPEAKER_ID)).SetNameInWorker("mockvoiceoverspeaker").SetModelID(uuid.MustParse(MOCK_VOICEOVER_MODEL_ID)).Save(ctx)
	if err != nil {
		return err
	}

	// ! Mock some generations
	// With negative prompt, success, and outpts
	gen, err := repo.CreateGeneration(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", requests.CreateGenerationRequest{
		Prompt:         "This is a prompt",
		NegativePrompt: "This is a negative prompt",
		Width:          utils.ToPtr[int32](512),
		Height:         utils.ToPtr[int32](512),
		InferenceSteps: utils.ToPtr[int32](11),
		GuidanceScale:  utils.ToPtr[float32](2.0),
		ModelId:        utils.ToPtr(uuid.MustParse(MOCK_GENERATION_MODEL_ID)),
		SchedulerId:    utils.ToPtr(uuid.MustParse(MOCK_SCHEDULER_ID)),
		Seed:           utils.ToPtr(1234),
		NumOutputs:     utils.ToPtr[int32](3),
	}, nil, nil, enttypes.SourceTypeWebUI, nil)
	if err != nil {
		return err
	}
	err = repo.SetGenerationStarted(gen.ID.String())
	if err != nil {
		return err
	}
	_, err = repo.SetGenerationSucceeded(gen.ID.String(), "This is a prompt", "This is a negative prompt", false,
		requests.CogWebhookOutput{
			Images: []requests.CogWebhookOutputImage{
				{
					Image: "output_1",
				},
				{
					Image: "output_2",
				},
				{
					Image: "output_3",
				},
			},
		}, 0)
	if err != nil {
		return err
	}

	// Without negative prompt, also success
	gen, err = repo.CreateGeneration(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", requests.CreateGenerationRequest{
		Prompt:          "This is a prompt 2",
		Width:           utils.ToPtr[int32](512),
		Height:          utils.ToPtr[int32](512),
		InferenceSteps:  utils.ToPtr[int32](30),
		GuidanceScale:   utils.ToPtr[float32](1.0),
		ModelId:         utils.ToPtr(uuid.MustParse(MOCK_GENERATION_MODEL_ID)),
		SchedulerId:     utils.ToPtr(uuid.MustParse(MOCK_SCHEDULER_ID)),
		Seed:            utils.ToPtr(1234),
		NumOutputs:      utils.ToPtr[int32](1),
		SubmitToGallery: true,
	}, nil, nil, enttypes.SourceTypeWebUI, nil)
	if err != nil {
		return err
	}
	err = repo.SetGenerationStarted(gen.ID.String())
	if err != nil {
		return err
	}
	_, err = repo.SetGenerationSucceeded(gen.ID.String(), "This is a prompt 2", "", true, requests.CogWebhookOutput{
		Images: []requests.CogWebhookOutputImage{
			{
				Image: "output_4",
			},
			{
				Image: "output_5",
			},
			{
				Image: "output_6",
			},
		},
	}, 0)
	if err != nil {
		return err
	}

	// Failure
	gen, err = repo.CreateGeneration(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", requests.CreateGenerationRequest{
		Prompt:         "This is a prompt 3",
		Width:          utils.ToPtr[int32](512),
		Height:         utils.ToPtr[int32](512),
		InferenceSteps: utils.ToPtr[int32](30),
		GuidanceScale:  utils.ToPtr[float32](1.0),
		ModelId:        utils.ToPtr(uuid.MustParse(MOCK_GENERATION_MODEL_ID)),
		SchedulerId:    utils.ToPtr(uuid.MustParse(MOCK_SCHEDULER_ID)),
		Seed:           utils.ToPtr(1234),
		NumOutputs:     utils.ToPtr[int32](1),
	}, nil, nil, enttypes.SourceTypeWebUI, nil)
	if err != nil {
		return err
	}
	err = repo.SetGenerationStarted(gen.ID.String())
	if err != nil {
		return err
	}
	err = repo.SetGenerationFailed(gen.ID.String(), "Failed to generate", 0, nil)
	if err != nil {
		return err
	}

	// In progress
	gen, err = repo.CreateGeneration(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", requests.CreateGenerationRequest{
		Prompt:         "This is a prompt 4",
		Width:          utils.ToPtr[int32](512),
		Height:         utils.ToPtr[int32](512),
		InferenceSteps: utils.ToPtr[int32](30),
		GuidanceScale:  utils.ToPtr[float32](1.0),
		ModelId:        utils.ToPtr(uuid.MustParse(MOCK_GENERATION_MODEL_ID)),
		SchedulerId:    utils.ToPtr(uuid.MustParse(MOCK_SCHEDULER_ID)),
		Seed:           utils.ToPtr(1234),
		NumOutputs:     utils.ToPtr[int32](1),
	}, nil, nil, enttypes.SourceTypeWebUI, nil)
	if err != nil {
		return err
	}
	err = repo.SetGenerationStarted(gen.ID.String())
	if err != nil {
		return err
	}

	// ! Mock voiceover
	// ! Mock some generations
	// With negative prompt, success, and outpts
	vo, err := repo.CreateVoiceover(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", requests.CreateVoiceoverRequest{
		Prompt:      "This is a prompt",
		ModelId:     utils.ToPtr(uuid.MustParse(MOCK_VOICEOVER_MODEL_ID)),
		SpeakerId:   utils.ToPtr(uuid.MustParse(MOCK_VOICEOVER_SPEAKER_ID)),
		Seed:        utils.ToPtr(1234),
		Temperature: utils.ToPtr(float32(0.5)),
	}, nil, nil, enttypes.SourceTypeWebUI, nil)
	if err != nil {
		return err
	}
	err = repo.SetVoiceoverStarted(vo.ID.String())
	if err != nil {
		return err
	}
	_, err = repo.SetVoiceoverSucceeded(vo.ID.String(), "This is a prompt", true,
		requests.CogWebhookOutput{
			AudioFiles: []requests.CogWebhookOutputAudio{
				{
					AudioFile: "output_1",
				},
			},
		})
	if err != nil {
		return err
	}

	return nil
}

func (repo *Repository) CreateMockGenerationForDeletion(ctx context.Context) (*ent.Generation, error) {
	gen, err := repo.CreateGeneration(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", requests.CreateGenerationRequest{
		Prompt:         "to_delete",
		Width:          utils.ToPtr[int32](512),
		Height:         utils.ToPtr[int32](512),
		InferenceSteps: utils.ToPtr[int32](30),
		GuidanceScale:  utils.ToPtr[float32](1.0),
		ModelId:        utils.ToPtr(uuid.MustParse(MOCK_GENERATION_MODEL_ID)),
		SchedulerId:    utils.ToPtr(uuid.MustParse(MOCK_SCHEDULER_ID)),
		Seed:           utils.ToPtr(1234),
		NumOutputs:     utils.ToPtr[int32](1),
	}, nil, nil, enttypes.SourceTypeWebUI, nil)
	if err != nil {
		return nil, err
	}
	err = repo.SetGenerationStarted(gen.ID.String())
	if err != nil {
		return nil, err
	}
	_, err = repo.SetGenerationSucceeded(gen.ID.String(), "to_delete", "", true, requests.CogWebhookOutput{
		Images: []requests.CogWebhookOutputImage{
			{
				Image: "output_4",
			},
			{
				Image: "output_5",
			},
			{
				Image: "output_6",
			},
		},
	}, 0)
	if err != nil {
		return nil, err
	}

	return gen, nil
}

func (repo *Repository) CreateMockUpscaleForDeletion(ctx context.Context) (*ent.Upscale, error) {
	up, err := repo.CreateUpscale(uuid.MustParse(MOCK_ADMIN_UUID), 512, 512, "browser", "macos", "chrome", "DE", requests.CreateUpscaleRequest{
		Type:    utils.ToPtr(requests.UpscaleRequestTypeOutput),
		ModelId: utils.ToPtr(uuid.MustParse(MOCK_UPSCALE_MODEL_ID)),
		Input:   uuid.NewString(),
	}, nil, false, nil, enttypes.SourceTypeWebUI, nil)
	if err != nil {
		return nil, err
	}

	return up, nil
}
