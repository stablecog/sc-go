package repository

import (
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/apitoken"
	"github.com/stablecog/sc-go/server/requests"
)

func (r *Repository) GetToken(id uuid.UUID) (*ent.ApiToken, error) {
	return r.DB.ApiToken.Get(r.Ctx, id)
}

func (r *Repository) GetTokensByUserID(userID uuid.UUID, filters *requests.ApiTokenQueryFilters) ([]*ent.ApiToken, error) {
	q := r.DB.ApiToken.Query().Where(apitoken.UserIDEQ(userID), apitoken.IsActive(true))
	if filters != nil {
		switch filters.ApiTokenType {
		case requests.ApiTokenClient:
			q = q.Where(apitoken.AuthClientIDNotNil())
		case requests.ApiTokenManual:
			q = q.Where(apitoken.AuthClientIDIsNil())
		case requests.ApiTokenAny:
		default:
			// Any
		}
	}
	q = q.Order(ent.Desc(apitoken.FieldCreatedAt))
	return q.All(r.Ctx)
}

func (r *Repository) GetTokenCountByUserID(userID uuid.UUID) (int, error) {
	return r.DB.ApiToken.Query().Where(apitoken.UserIDEQ(userID), apitoken.IsActive(true), apitoken.AuthClientIDIsNil()).Count(r.Ctx)
}

func (r *Repository) GetTokenByHashedToken(hashedToken string) (*ent.ApiToken, error) {
	return r.DB.ApiToken.Query().Where(apitoken.HashedTokenEQ(hashedToken), apitoken.IsActiveEQ(true)).Only(r.Ctx)
}
