package database

import (
	"context"

	client "github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
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

func (m *WeaviateClient) CreateSchema() error {
	False := false

	// See if class exists first
	s, _ := m.Client.Schema().ClassGetter().WithClassName("Test").Do(m.Ctx)
	if s != nil {
		return nil
	}
	return m.Client.Schema().ClassCreator().WithClass(&models.Class{
		Class:           "Test",
		Description:     "Testing",
		VectorIndexType: "hnsw",
		Vectorizer:      "none",
		VectorIndexConfig: map[string]interface{}{
			"pq": map[string]interface{}{
				"enabled": true,
			},
		},
		Properties: []*models.Property{
			{
				Name:          "image_path",
				IndexInverted: &False,
				DataType:      []string{"string"},
			},
			{
				Name:          "prompt",
				IndexInverted: &False,
				DataType:      []string{"string"},
			},
		},
	}).Do(m.Ctx)
}

func (w *WeaviateClient) SearchNearVector(vector []float32) (map[string]models.JSONObject, error) {
	imagePath := graphql.Field{Name: "image_path"}
	prompt := graphql.Field{Name: "prompt"}

	ctx := context.Background()
	result, err := w.Client.GraphQL().Get().
		WithClassName("Test").
		WithFields(imagePath, prompt).
		WithNearVector(w.Client.GraphQL().NearVectorArgBuilder().WithVector(vector)).
		WithLimit(50).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	return result.Data, nil
}
