package controller

import (
	"github.com/go-redis/redis/v8"
	"github.com/stablecog/go-apps/database/repository"
)

type HttpController struct {
	Repo  *repository.Repository
	Redis *redis.Client
}
