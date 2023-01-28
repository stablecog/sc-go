package database

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent"
	"github.com/stablecog/go-apps/database/ent/userrole"
	"github.com/stablecog/go-apps/database/repository"
	"github.com/stablecog/go-apps/server/requests"
)

// Mock user IDs
const MOCK_ADMIN_UUID = "00000000-0000-0000-0000-000000000000"
const MOCK_PRO_UUID = "00000000-0000-0000-0000-000000000001"
const MOCK_FREE_UUID = "00000000-0000-0000-0000-000000000002"

// Mock generation model IDs and scheduler IDs
const MOCK_GENERATION_MODEL_ID_FREE = "b972a2b8-f39e-4ee3-a670-05e3acdd821c"
const MOCK_GENERATION_MODEL_ID_PRO = "b972a2b8-f39e-4ee3-a670-05e3acdd821d"
const MOCK_SCHEDULER_ID_FREE = "b4dff6e9-91a7-449b-b1a7-c25000e3ccd0"
const MOCK_SCHEDULER_ID_PRO = "b4dff6e9-91a7-449b-b1a7-c25000e3ccd1"

// Just creates some mock data for our tests
func CreateMockData(ctx context.Context, db *ent.Client, repo *repository.Repository) error {
	// ! Mock users
	// Create a user
	u, err := db.User.Create().SetEmail("mockadmin@stablecog.com").SetID(uuid.MustParse(MOCK_ADMIN_UUID)).SetConfirmedAt(time.Now()).Save(ctx)
	if err != nil {
		return err
	}
	// Give user admin role
	_, err = db.UserRole.Create().SetRoleName(userrole.RoleNameADMIN).SetUserID(u.ID).Save(ctx)
	if err != nil {
		return err
	}
	// Create two more users, "PRO" and free tier
	u, err = db.User.Create().SetEmail("mockpro@stablecog.com").SetID(uuid.MustParse(MOCK_PRO_UUID)).SetConfirmedAt(time.Now()).Save(ctx)
	if err != nil {
		return err
	}
	// Give user PRO role
	_, err = db.UserRole.Create().SetRoleName(userrole.RoleNamePRO).SetUserID(u.ID).Save(ctx)
	if err != nil {
		return err
	}
	u, err = db.User.Create().SetEmail("mockbasic@stablecog.com").SetID(uuid.MustParse(MOCK_FREE_UUID)).SetConfirmedAt(time.Now()).Save(ctx)
	if err != nil {
		return err
	}

	// ! Mock generation models
	// Create a generation model for the free user
	_, err = db.GenerationModel.Create().SetID(uuid.MustParse(MOCK_GENERATION_MODEL_ID_FREE)).SetName("mockfreemodel").SetIsFree(true).Save(ctx)
	if err != nil {
		return err
	}
	// Create a generation model for the pro user
	_, err = db.GenerationModel.Create().SetID(uuid.MustParse(MOCK_GENERATION_MODEL_ID_PRO)).SetName("mockpromodel").SetIsFree(false).Save(ctx)
	if err != nil {
		return err
	}

	// ! Mock schedulers
	// Create a scheduler for the free user
	_, err = db.Scheduler.Create().SetID(uuid.MustParse(MOCK_SCHEDULER_ID_FREE)).SetName("mockfreescheduler").SetIsFree(true).Save(ctx)
	if err != nil {
		return err
	}
	// Create a scheduler for the pro user
	_, err = db.Scheduler.Create().SetID(uuid.MustParse(MOCK_SCHEDULER_ID_PRO)).SetName("mockproscheduler").SetIsFree(false).Save(ctx)
	if err != nil {
		return err
	}

	// ! Mock some generations
	// With negative prompt, success, and outpts
	gen, err := repo.CreateGeneration(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", requests.GenerateRequestBody{
		Prompt:            "This is a prompt",
		NegativePrompt:    "This is a negative prompt",
		Width:             512,
		Height:            512,
		NumInferenceSteps: 30,
		GuidanceScale:     1.0,
		ModelId:           uuid.MustParse(MOCK_GENERATION_MODEL_ID_FREE),
		SchedulerId:       uuid.MustParse(MOCK_SCHEDULER_ID_FREE),
		Seed:              1234,
	})
	if err != nil {
		return err
	}
	err = repo.SetGenerationStarted(gen.ID.String())
	if err != nil {
		return err
	}
	err = repo.SetGenerationSucceeded(gen.ID.String(), []string{"output_1", "output_2", "output_3"})
	if err != nil {
		return err
	}

	// Without negative prompt, also success
	gen, err = repo.CreateGeneration(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", requests.GenerateRequestBody{
		Prompt:            "This is a prompt 2",
		Width:             512,
		Height:            512,
		NumInferenceSteps: 30,
		GuidanceScale:     1.0,
		ModelId:           uuid.MustParse(MOCK_GENERATION_MODEL_ID_FREE),
		SchedulerId:       uuid.MustParse(MOCK_SCHEDULER_ID_FREE),
		Seed:              1234,
	})
	if err != nil {
		return err
	}
	err = repo.SetGenerationStarted(gen.ID.String())
	if err != nil {
		return err
	}
	err = repo.SetGenerationSucceeded(gen.ID.String(), []string{"output_4", "output_5", "output_6"})
	if err != nil {
		return err
	}

	// Failure
	gen, err = repo.CreateGeneration(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", requests.GenerateRequestBody{
		Prompt:            "This is a prompt 3",
		Width:             512,
		Height:            512,
		NumInferenceSteps: 30,
		GuidanceScale:     1.0,
		ModelId:           uuid.MustParse(MOCK_GENERATION_MODEL_ID_FREE),
		SchedulerId:       uuid.MustParse(MOCK_SCHEDULER_ID_FREE),
		Seed:              1234,
	})
	if err != nil {
		return err
	}
	err = repo.SetGenerationStarted(gen.ID.String())
	if err != nil {
		return err
	}
	err = repo.SetGenerationFailed(gen.ID.String(), "Failed to generate")
	if err != nil {
		return err
	}

	// In progress
	gen, err = repo.CreateGeneration(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", requests.GenerateRequestBody{
		Prompt:            "This is a prompt 4",
		Width:             512,
		Height:            512,
		NumInferenceSteps: 30,
		GuidanceScale:     1.0,
		ModelId:           uuid.MustParse(MOCK_GENERATION_MODEL_ID_FREE),
		SchedulerId:       uuid.MustParse(MOCK_SCHEDULER_ID_FREE),
		Seed:              1234,
	})
	if err != nil {
		return err
	}
	err = repo.SetGenerationStarted(gen.ID.String())
	if err != nil {
		return err
	}

	return nil
}
