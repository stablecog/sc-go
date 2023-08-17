package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/database/enttypes"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetUsersUniquePromptIds(t *testing.T) {
	// Create 2 prompts
	p1, _, err := MockRepo.GetOrCreatePrompts("TestGetUsersUniquePromptIds_1", "", prompt.TypeImage, nil)
	assert.Nil(t, err)
	assert.NotNil(t, p1)
	p2, _, err := MockRepo.GetOrCreatePrompts("TestGetUsersUniquePromptIds_2", "", prompt.TypeImage, nil)
	assert.Nil(t, err)
	assert.NotNil(t, p2)

	// Create 3 generations, 1 user sharing a prompt with another user
	// Unique prompt
	g1 := requests.CreateGenerationRequest{
		Prompt:         "TestGetUsersUniquePromptIds_1",
		Width:          utils.ToPtr[int32](512),
		Height:         utils.ToPtr[int32](512),
		InferenceSteps: utils.ToPtr[int32](11),
		GuidanceScale:  utils.ToPtr[float32](2.0),
		ModelId:        utils.ToPtr(uuid.MustParse(MOCK_GENERATION_MODEL_ID)),
		SchedulerId:    utils.ToPtr(uuid.MustParse(MOCK_SCHEDULER_ID)),
		Seed:           utils.ToPtr(1234),
	}
	g1.ApplyDefaults()

	// Shared prompt
	g2 := requests.CreateGenerationRequest{
		Prompt:         "TestGetUsersUniquePromptIds_2",
		Width:          utils.ToPtr[int32](512),
		Height:         utils.ToPtr[int32](512),
		InferenceSteps: utils.ToPtr[int32](11),
		GuidanceScale:  utils.ToPtr[float32](2.0),
		ModelId:        utils.ToPtr(uuid.MustParse(MOCK_GENERATION_MODEL_ID)),
		SchedulerId:    utils.ToPtr(uuid.MustParse(MOCK_SCHEDULER_ID)),
		Seed:           utils.ToPtr(1234),
	}
	g2.ApplyDefaults()

	gen1, err := MockRepo.CreateGeneration(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", g1, nil, nil, enttypes.SourceTypeWebUI, nil)
	assert.Nil(t, err)
	assert.NotNil(t, gen1)
	_, err = MockRepo.SetGenerationSucceeded(gen1.ID.String(), "TestGetUsersUniquePromptIds_1", "", true, requests.CogWebhookOutput{
		Images: []requests.CogWebhookOutputImage{{
			Image: "1.jpeg",
		}},
	}, 0)
	assert.Nil(t, err)

	// 2 different users, same prompt
	gen2, err := MockRepo.CreateGeneration(uuid.MustParse(MOCK_NORMAL_UUID), "browser", "macos", "chrome", "DE", g2, nil, nil, enttypes.SourceTypeWebUI, nil)
	assert.Nil(t, err)
	assert.NotNil(t, gen2)
	_, err = MockRepo.SetGenerationSucceeded(gen2.ID.String(), "TestGetUsersUniquePromptIds_2", "", true, requests.CogWebhookOutput{
		Images: []requests.CogWebhookOutputImage{{
			Image: "2.jpeg",
		}},
	}, 0)
	assert.Nil(t, err)

	gen3, err := MockRepo.CreateGeneration(uuid.MustParse(MOCK_ADMIN_UUID), "browser", "macos", "chrome", "DE", g2, nil, nil, enttypes.SourceTypeWebUI, nil)
	assert.Nil(t, err)
	assert.NotNil(t, gen3)
	_, err = MockRepo.SetGenerationSucceeded(gen3.ID.String(), "TestGetUsersUniquePromptIds_2", "", true, requests.CogWebhookOutput{
		Images: []requests.CogWebhookOutputImage{{
			Image: "3.jpeg",
		}},
	}, 0)
	assert.Nil(t, err)

	// Get prompts not used by user
	// Normal should have p2 not p1
	prompts, err := MockRepo.GetUsersUniquePromptIds([]uuid.UUID{*p2}, uuid.MustParse(MOCK_NORMAL_UUID))
	assert.Nil(t, err)
	assert.Empty(t, prompts)

	// Admin should have 1 unique prompt (p1)
	prompts, err = MockRepo.GetUsersUniquePromptIds([]uuid.UUID{*p1, *p2}, uuid.MustParse(MOCK_ADMIN_UUID))
	assert.Nil(t, err)
	assert.Equal(t, []uuid.UUID{*p1}, prompts)

	// Cleanup
	MockRepo.DB.Prompt.Delete().Where(prompt.IDEQ(*p1)).ExecX(MockRepo.Ctx)
	MockRepo.DB.Prompt.Delete().Where(prompt.IDEQ(*p2)).ExecX(MockRepo.Ctx)
	MockRepo.DB.Generation.Delete().Where(generation.IDEQ(gen1.ID)).ExecX(MockRepo.Ctx)
	MockRepo.DB.Generation.Delete().Where(generation.IDEQ(gen2.ID)).ExecX(MockRepo.Ctx)
	MockRepo.DB.Generation.Delete().Where(generation.IDEQ(gen3.ID)).ExecX(MockRepo.Ctx)
}
