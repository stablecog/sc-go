package repository

import (
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/upscaleoutput"
)

func (r *Repository) GetUpscaleOutputWithPath(path string) (*ent.UpscaleOutput, error) {
	return r.DB.UpscaleOutput.Query().Where(upscaleoutput.ImagePathEQ(path)).Only(r.Ctx)
}
