package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetGalleryData(t *testing.T) {
	// Approve generations
	gOutputs := MockRepo.DB.GenerationOutput.Query().Where(generationoutput.GalleryStatusEQ(generationoutput.GalleryStatusSubmitted)).AllX(MockRepo.Ctx)
	gOutputIDs := make([]uuid.UUID, len(gOutputs))
	for i, gOutput := range gOutputs {
		gOutputIDs[i] = gOutput.ID
	}
	_, err := MockRepo.BulkUpdateGalleryStatusForOutputs(gOutputIDs, generationoutput.GalleryStatusApproved)
	assert.Nil(t, err)

	// Check data
	gData, _, _, err := MockRepo.RetrieveMostRecentGalleryDataV2(&requests.QueryGenerationFilters{
		GalleryStatus: []generationoutput.GalleryStatus{generationoutput.GalleryStatusApproved},
	}, nil, 100, nil, nil)
	assert.Nil(t, err)
	assert.Len(t, gData, 3)

	// Check data with calling user ID
	gData, _, _, err = MockRepo.RetrieveMostRecentGalleryDataV2(&requests.QueryGenerationFilters{
		GalleryStatus: []generationoutput.GalleryStatus{generationoutput.GalleryStatusApproved},
	}, utils.ToPtr(uuid.MustParse(MOCK_ADMIN_UUID)), 100, nil, nil)
	assert.Nil(t, err)
	assert.Len(t, gData, 3)
}
