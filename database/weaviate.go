package database

import (
	"context"

	client "github.com/weaviate/weaviate-go-client/v4/weaviate"
)

type WeaviateClient struct {
	Client *client.Client
	Ctx    context.Context
}

func NewWeaviateClient(ctx context.Context) *WeaviateClient {
	c := client.New(client.Config{
		Scheme: "http",
		Host:   "weaviate.weaviate:80",
	})
	return &WeaviateClient{
		Client: c,
		Ctx:    ctx,
	}
}
