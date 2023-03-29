package repository

import (
	"github.com/stablecog/sc-go/database/ent"
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
		Order(ent.Desc(generationoutput.FieldCreatedAt)).
		Limit(limit).
		All(r.Ctx)
}
