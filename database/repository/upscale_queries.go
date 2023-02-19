package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/upscale"
)

// Get upscale by ID
func (r *Repository) GetUpscale(id uuid.UUID) (*ent.Upscale, error) {
	return r.DB.Upscale.Query().Where(upscale.IDEQ(id)).First(r.Ctx)
}
