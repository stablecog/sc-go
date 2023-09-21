package api

import (
	"github.com/stablecog/sc-go/auth/store"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/repository"
	"github.com/stablecog/sc-go/utils"
)

type ApiWrapper struct {
	RedisStore   *store.RedisStore
	SupabaseAuth *database.SupabaseAuth
	AesCrypt     *utils.AESCrypt
	DB           *ent.Client
	Repo         *repository.Repository
}
