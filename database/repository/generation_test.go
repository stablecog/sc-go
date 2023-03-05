package repository

import (
	"context"
	"testing"

	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stretchr/testify/assert"
)

func TestGetSingleGenerationQueryWithOutputsResultFormatted(t *testing.T) {
	// Create some generations and outputs and approve in gallery
	g, err := MockRepo.CreateMockGenerationForDeletion(context.Background())
	assert.Nil(t, err)

	// Update to approved all outputs
	_, err = MockRepo.DB.GenerationOutput.Update().Where(generationoutput.GenerationIDEQ(g.ID)).SetGalleryStatus(generationoutput.GalleryStatusApproved).Save(MockRepo.Ctx)
	assert.Nil(t, err)

	// Get a single output id that is approved
	goutput, err := MockRepo.DB.Generation.Query().Where(generation.IDEQ(g.ID)).QueryGenerationOutputs().Where(generationoutput.GalleryStatusEQ(generationoutput.GalleryStatusApproved)).First(MockRepo.Ctx)
	assert.Nil(t, err)

	// Get the generation
	format, err := MockRepo.GetSingleGenerationQueryWithOutputsResultFormatted(goutput.ID)
	assert.Nil(t, err)
	assert.Len(t, format.Generation.Outputs, 3)

	// Delete the generation
	assert.Nil(t, MockRepo.DB.Generation.DeleteOne(g).Exec(MockRepo.Ctx))
}
