package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stretchr/testify/assert"
)

func TestQueryUserOperations(t *testing.T) {
	// Create mocks
	u, err := MockRepo.CreateMockUpscaleForDeletion(MockRepo.Ctx)
	assert.Nil(t, err)
	assert.NotNil(t, u)
	assert.Nil(t, err)
	g, err := MockRepo.CreateMockGenerationForDeletion(MockRepo.Ctx)
	assert.Nil(t, err)
	assert.NotNil(t, g)
	assert.Nil(t, MockRepo.SetGenerationStarted(g.ID.String()))
	outputs, err := MockRepo.SetGenerationSucceeded(g.ID.String(), "TestQueryUserOperations_!", "", requests.CogWebhookOutput{
		Images: []requests.CogWebhookOutputImage{{
			Image: "3.jpeg",
		}},
	}, 0)
	assert.Nil(t, err)
	assert.Nil(t, MockRepo.SetUpscaleStarted(u.ID.String()))
	_, err = MockRepo.SetUpscaleSucceeded(u.ID.String(), outputs[0].ID.String(), "", requests.CogWebhookOutput{
		Images: []requests.CogWebhookOutputImage{{
			Image: "3.jpeg",
		}},
	})
	assert.Nil(t, err)

	// Query
	ops, err := MockRepo.QueryUserOperations(uuid.MustParse(MOCK_ADMIN_UUID), 2, nil)
	assert.Nil(t, err)
	assert.NotNil(t, ops)
	assert.Len(t, ops.Operations, 2)
	assert.NotNil(t, ops.Next)
	assert.Equal(t, ops.Operations[0].ID, g.ID)
	assert.Equal(t, ops.Operations[1].ID, u.ID)

	// Delete
	MockRepo.DB.Upscale.DeleteOne(u).ExecX(MockRepo.Ctx)
	MockRepo.DB.Generation.DeleteOne(g).ExecX(MockRepo.Ctx)
}
