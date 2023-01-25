package controller

import (
	"github.com/stablecog/go-apps/database"
	"github.com/stablecog/go-apps/database/repository"
)

type HttpController struct {
	Repo  *repository.Repository
	Redis *database.RedisWrapper
}
