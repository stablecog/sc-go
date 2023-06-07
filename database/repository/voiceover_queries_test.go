package repository

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestQueryVoiceoversForUser(t *testing.T) {
	filters := requests.QueryVoiceoverFilters{
		UserID: utils.ToPtr(uuid.MustParse(MOCK_ADMIN_UUID)),
	}

	voiceovers, err := MockRepo.QueryVoiceovers(50, nil, &filters)
	assert.Nil(t, err)
	assert.Nil(t, voiceovers.Next)
	assert.Len(t, voiceovers.Outputs, 1)
}
