package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/credittype"
	"github.com/stablecog/sc-go/database/ent/userrole"
	"github.com/stablecog/sc-go/server/requests"
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

// Just creates some mock data for our tests
func (repo *Repository) CreateMockData(ctx context.Context) error {
	// Drop all data
	repo.DB.User.Delete().ExecX(ctx)
	repo.DB.UserRole.Delete().ExecX(ctx)
	repo.DB.Credit.Delete().ExecX(ctx)
	repo.DB.CreditType.Delete().ExecX(ctx)
	repo.DB.GenerationModel.Delete().ExecX(ctx)
	repo.DB.Scheduler.Delete().ExecX(ctx)
	repo.DB.UpscaleModel.Delete().ExecX(ctx)
	repo.DB.Generation.Delete().ExecX(ctx)
	repo.DB.GenerationOutput.Delete().ExecX(ctx)
	repo.DB.Upscale.Delete().ExecX(ctx)

	// Create a credit type
	stripeProductId := "prod_123"
	creditType, err := repo.CreateCreditType("mock", 100, nil, &stripeProductId, credittype.TypeSubscription)
	if err != nil {
		return err
	}

	// ! Mock users
	// Create a user
	u, err := repo.DB.User.Create().SetEmail("mockadmin@stablecog.com").SetID(uuid.MustParse(MOCK_ADMIN_UUID)).SetStripeCustomerID("1").Save(ctx)
	if err != nil {
		return err
	}
	// Give user admin role
	_, err = repo.DB.UserRole.Create().SetRoleName(userrole.RoleNameSUPER_ADMIN).SetUserID(u.ID).Save(ctx)
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
	u, err = repo.DB.User.Create().SetEmail("mockuser@stablecog.com").SetID(uuid.MustParse(MOCK_NORMAL_UUID)).SetStripeCustomerID("2").Save(ctx)
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
	u, err = repo.DB.User.Create().SetEmail("mockaltuser@stablecog.com").SetID(uuid.MustParse(MOCK_ALT_UUID)).SetStripeCustomerID("3").Save(ctx)
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
	u, err = repo.DB.User.Create().SetEmail("mocknocredituser@stablecog.com").SetID(uuid.MustParse(MOCK_NO_CREDITS_UUID)).SetStripeCustomerID("4").SetLastSignInAt(time.Now()).Save(ctx)
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

	// ! Mock some generations
	// With negative prompt, success, and outpts
	gen, err := repo.CreateGeneration(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", requests.CreateGenerationRequest{
		Prompt:         "This is a prompt",
		NegativePrompt: "This is a negative prompt",
		Width:          512,
		Height:         512,
		InferenceSteps: 11,
		GuidanceScale:  2.0,
		ModelId:        uuid.MustParse(MOCK_GENERATION_MODEL_ID),
		SchedulerId:    uuid.MustParse(MOCK_SCHEDULER_ID),
		Seed:           1234,
	}, nil, nil, nil)
	if err != nil {
		return err
	}
	err = repo.SetGenerationStarted(gen.ID.String())
	if err != nil {
		return err
	}
	_, err = repo.SetGenerationSucceeded(gen.ID.String(), "This is a prompt", "This is a negative prompt",
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
		Width:           512,
		Height:          512,
		InferenceSteps:  30,
		GuidanceScale:   1.0,
		ModelId:         uuid.MustParse(MOCK_GENERATION_MODEL_ID),
		SchedulerId:     uuid.MustParse(MOCK_SCHEDULER_ID),
		Seed:            1234,
		NumOutputs:      1,
		SubmitToGallery: true,
	}, nil, nil, nil)
	if err != nil {
		return err
	}
	err = repo.SetGenerationStarted(gen.ID.String())
	if err != nil {
		return err
	}
	_, err = repo.SetGenerationSucceeded(gen.ID.String(), "This is a prompt 2", "", requests.CogWebhookOutput{
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
		Width:          512,
		Height:         512,
		InferenceSteps: 30,
		GuidanceScale:  1.0,
		ModelId:        uuid.MustParse(MOCK_GENERATION_MODEL_ID),
		SchedulerId:    uuid.MustParse(MOCK_SCHEDULER_ID),
		Seed:           1234,
		NumOutputs:     1,
	}, nil, nil, nil)
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
		Width:          512,
		Height:         512,
		InferenceSteps: 30,
		GuidanceScale:  1.0,
		ModelId:        uuid.MustParse(MOCK_GENERATION_MODEL_ID),
		SchedulerId:    uuid.MustParse(MOCK_SCHEDULER_ID),
		Seed:           1234,
		NumOutputs:     1,
	}, nil, nil, nil)
	if err != nil {
		return err
	}
	err = repo.SetGenerationStarted(gen.ID.String())
	if err != nil {
		return err
	}

	return nil
}

func (repo *Repository) CreateMockGenerationForDeletion(ctx context.Context) (*ent.Generation, error) {
	gen, err := repo.CreateGeneration(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", requests.CreateGenerationRequest{
		Prompt:         "to_delete",
		Width:          512,
		Height:         512,
		InferenceSteps: 30,
		GuidanceScale:  1.0,
		ModelId:        uuid.MustParse(MOCK_GENERATION_MODEL_ID),
		SchedulerId:    uuid.MustParse(MOCK_SCHEDULER_ID),
		Seed:           1234,
		NumOutputs:     1,
	}, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	err = repo.SetGenerationStarted(gen.ID.String())
	if err != nil {
		return nil, err
	}
	_, err = repo.SetGenerationSucceeded(gen.ID.String(), "to_delete", "", requests.CogWebhookOutput{
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
		Type:    requests.UpscaleRequestTypeOutput,
		ModelId: uuid.MustParse(MOCK_UPSCALE_MODEL_ID),
		Input:   uuid.NewString(),
	}, nil, false, nil, nil)
	if err != nil {
		return nil, err
	}

	return up, nil
}
