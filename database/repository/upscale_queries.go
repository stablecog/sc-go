package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/upscale"
)

// Get upscale by ID
func (r *Repository) GetUpscale(id uuid.UUID) (*ent.Upscale, error) {
	return r.DB.Upscale.Query().Where(upscale.IDEQ(id)).First(r.Ctx)
}

func (r *Repository) GetUpscalesQueuedOrStarted() ([]*ent.Upscale, error) {
	// Get generations that are started/queued and older than 5 minutes
	return r.DB.Upscale.Query().
		Where(
			upscale.StatusIn(
				upscale.StatusQueued,
				upscale.StatusStarted,
			),
			upscale.CreatedAtLT(time.Now().Add(-5*time.Minute)),
		).
		Order(ent.Desc(generation.FieldCreatedAt)).
		Limit(100).
		All(r.Ctx)
}
