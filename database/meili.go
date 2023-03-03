package database

import (
	"github.com/meilisearch/meilisearch-go"
	"github.com/stablecog/sc-go/utils"
)

func NewMeiliSearchClient() *meilisearch.Client {
	return meilisearch.NewClient(meilisearch.ClientConfig{
		Host:   getMeiliUrl(),
		APIKey: utils.GetEnv("MEILI_MASTER_KEY", ""),
	})
}

func getMeiliUrl() string {
	return utils.GetEnv("MEILI_URL", "")
}
