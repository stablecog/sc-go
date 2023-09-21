package api

import (
	"github.com/stablecog/sc-go/auth/store"
	"github.com/stablecog/sc-go/database"
	"github.com/stablecog/sc-go/utils"
)

type ApiWrapper struct {
	RedisStore   *store.RedisStore
	SupabaseAuth *database.SupabaseAuth
	AesCrypt     *utils.AESCrypt
}