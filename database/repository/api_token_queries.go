package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/apitoken"
)

func (r *Repository) GetToken(id uuid.UUID) (*ent.ApiToken, error) {
	return r.DB.ApiToken.Get(r.Ctx, id)
}

func (r *Repository) GetTokensByUserID(userID uuid.UUID, activeOnly bool) ([]*ent.ApiToken, error) {
	q := r.DB.ApiToken.Query().Where(apitoken.UserIDEQ(userID))
	if activeOnly {
		q = q.Where(apitoken.IsActive(true))
	}
	return q.All(r.Ctx)
}

func (r *Repository) GetTokenCountByUserID(userID uuid.UUID) (int, error) {
	return r.DB.ApiToken.Query().Where(apitoken.UserIDEQ(userID)).Count(r.Ctx)
}

func (r *Repository) GetTokenByHashedToken(hashedToken string) (*ent.ApiToken, error) {
	return r.DB.ApiToken.Query().Where(apitoken.HashedTokenEQ(hashedToken)).Only(r.Ctx)
}
