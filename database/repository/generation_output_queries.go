package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
)

func (r *Repository) GetNonUpscaledGalleryItems(limit int) ([]*ent.GenerationOutput, error) {
	return r.DB.GenerationOutput.Query().
		Where(
			generationoutput.GalleryStatusEQ(generationoutput.GalleryStatusApproved),
			generationoutput.Or(
				generationoutput.UpscaledImagePathIsNil(),
				generationoutput.UpscaledImagePathEQ(""),
			),
		).
		Order(ent.Desc(generationoutput.FieldUpdatedAt, generationoutput.FieldCreatedAt)).
		Limit(limit).
		All(r.Ctx)
}

func (r *Repository) GetUserGenerationOutputs(userId uuid.UUID) ([]*ent.GenerationOutput, error) {
	return r.DB.Generation.Query().Where(
		generation.UserIDEQ(userId),
	).QueryGenerationOutputs().All(r.Ctx)
}

func (r *Repository) GetPromptFromOutputID(outputID uuid.UUID) (string, error) {
	output, err := r.DB.GenerationOutput.Get(r.Ctx, outputID)
	if err != nil {
		return "", err
	}

	prompt, err := output.QueryGenerations().QueryPrompt().Only(r.Ctx)
	if err != nil {
		return "", err
	}

	return prompt.Text, nil
}
