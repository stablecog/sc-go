package repository

import (
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/apitoken"
	"github.com/stablecog/sc-go/utils"
)

func (r *Repository) NewAPIToken(userId uuid.UUID) (dbToken *ent.ApiToken, token string, err error) {
	// Create a new random 64 character token
	token, err = utils.GenerateRandomHex(nil, 32)
	if err != nil {
		return nil, "", err
	}

	// Create in DB
	dbToken, err = r.DB.ApiToken.Create().SetHashedToken(utils.Sha256(token)).SetUserID(userId).SetIsActive(true).SetUses(0).Save(r.Ctx)
	if err != nil {
		return nil, "", err
	}
	return dbToken, token, nil
}

func (r *Repository) SetTokenUsed(amount int, tokenId uuid.UUID) error {
	return r.DB.ApiToken.Update().Where(apitoken.IDEQ(tokenId)).AddUses(amount).SetLastUsedAt(time.Now()).Exec(r.Ctx)
}

func (r *Repository) DeactivateTokenForUser(id uuid.UUID, userId uuid.UUID) (int, error) {
	return r.DB.ApiToken.Update().Where(apitoken.IDEQ(id), apitoken.UserIDEQ(userId), apitoken.IsActive(true)).SetIsActive(false).Save(r.Ctx)
}
