package qdrant

import (
	"context"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

func TestGetCollectionsResponse(t *testing.T) {
	godotenv.Load("../../.env")
	c, err := NewQdrantClient(context.Background())
	assert.Nil(t, err)
	resp, err := c.GetCollections(false)
	assert.Nil(t, err)
	assert.Len(t, resp.Collections, 1)
	assert.NotNil(t, resp)
}

func TestCreateCollection(t *testing.T) {
	godotenv.Load("../../.env")
	c, err := NewQdrantClient(context.Background())
	assert.Nil(t, err)
	err = c.CreateCollection("qa_generation_outputs_flat", false)
	assert.Nil(t, err)
}
