package database

import (
	"context"
	"fmt"

	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/utils"
)

const MILVUS_COLLECTION_NAME = "generation_data"

type MilvusClient struct {
	Client client.Client
	Ctx    context.Context
}

func NewMilvusClient(ctx context.Context) (*MilvusClient, error) {
	c, err := client.NewDefaultGrpcClientWithAuth(ctx, utils.GetEnv("MILVUS_ENDPOINT", ""), utils.GetEnv("MILVUS_USER", ""), utils.GetEnv("MILVUS_PASSWORD", ""))
	if err != nil {
		log.Errorf("failed to connect to milvus, err: %v", err)
		return nil, err
	}
	return &MilvusClient{
		Client: c,
		Ctx:    ctx,
	}, nil
}

func (m *MilvusClient) Close() {
	m.Client.Close()
}

func (m *MilvusClient) CreateCollectionIfNotExists() error {
	hasCollection, err := m.Client.HasCollection(m.Ctx, MILVUS_COLLECTION_NAME)
	if err != nil {
		log.Errorf("failed to check if collection exists, err: %v", err)
		return err
	}
	if hasCollection {
		return nil
	}
	schema := &entity.Schema{
		CollectionName: MILVUS_COLLECTION_NAME,
		Description:    "generation+generation_outputs flat",
		AutoID:         false,
		Fields: []*entity.Field{
			{
				Name:       "id",
				DataType:   entity.FieldTypeVarChar,
				PrimaryKey: true,
				AutoID:     false,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: fmt.Sprintf("%d", 48),
				},
			},
			{
				Name:     "image_embedding",
				DataType: entity.FieldTypeFloatVector,
				TypeParams: map[string]string{
					entity.TypeParamDim: fmt.Sprint(1024),
				},
			},
			{
				Name:       "image_path",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: fmt.Sprintf("%d", 500),
				},
			},
			{
				Name:       "upscaled_image_path",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: fmt.Sprintf("%d", 500),
				},
			},
			{
				Name:       "gallery_status",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: fmt.Sprintf("%d", 50),
				},
			},
			{
				Name:       "is_favorited",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeBool,
			},
			{
				Name:       "width",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeInt32,
			},
			{
				Name:       "height",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeInt32,
			},
			{
				Name:       "model_id",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: fmt.Sprintf("%d", 48),
				},
			},
			{
				Name:       "scheduler_id",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: fmt.Sprintf("%d", 48),
				},
			},
			{
				Name:       "generation_id",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: fmt.Sprintf("%d", 48),
				},
			},
			{
				Name:       "user_id",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: fmt.Sprintf("%d", 48),
				},
			},
			{
				Name:       "prompt_text",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: fmt.Sprintf("%d", 2000),
				},
			},
			{
				Name:       "negative_prompt_text",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeVarChar,
				TypeParams: map[string]string{
					entity.TypeParamMaxLength: fmt.Sprintf("%d", 2000),
				},
			},
			{
				Name:       "created_at",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeInt64,
			},
			{
				Name:       "updated_at",
				PrimaryKey: false,
				AutoID:     false,
				DataType:   entity.FieldTypeInt64,
			},
		},
	}

	// create collection with consistency level, which serves as the default search/query consistency level
	if err := m.Client.CreateCollection(m.Ctx, schema, 2, client.WithConsistencyLevel(entity.ClSession)); err != nil {
		log.Errorf("create collection failed, err: %v", err)
		return err
	}
	return nil
}

func (m *MilvusClient) InsertOutput(output *ent.GenerationOutput, generation *ent.Generation, promptText string, negativePromptText string, imageEmbedding []float32) error {
	var upscaledImagePath string
	if output.UpscaledImagePath != nil {
		upscaledImagePath = *output.UpscaledImagePath
	}
	columns := []entity.Column{
		entity.NewColumnVarChar("id", []string{output.ID.String()}),
		entity.NewColumnFloatVector("image_embedding", 1024, [][]float32{imageEmbedding}),
		entity.NewColumnVarChar("image_path", []string{output.ImagePath}),
		entity.NewColumnVarChar("upscaled_image_path", []string{upscaledImagePath}),
		entity.NewColumnVarChar("gallery_status", []string{output.GalleryStatus.String()}),
		entity.NewColumnBool("is_favorited", []bool{output.IsFavorited}),
		entity.NewColumnInt32("width", []int32{generation.Width}),
		entity.NewColumnInt32("height", []int32{generation.Height}),
		entity.NewColumnVarChar("model_id", []string{generation.ModelID.String()}),
		entity.NewColumnVarChar("scheduler_id", []string{generation.SchedulerID.String()}),
		entity.NewColumnVarChar("generation_id", []string{generation.ID.String()}),
		entity.NewColumnVarChar("user_id", []string{generation.UserID.String()}),
		entity.NewColumnVarChar("prompt_text", []string{promptText}),
		entity.NewColumnVarChar("negative_prompt_text", []string{promptText}),
		entity.NewColumnInt64("created_at", []int64{output.CreatedAt.Unix()}),
		entity.NewColumnInt64("updated_at", []int64{output.UpdatedAt.Unix()}),
	}

	_, err := m.Client.Insert(m.Ctx, MILVUS_COLLECTION_NAME, "", columns...)
	if err != nil {
		log.Errorf("failed to insert film data %v", err)
		return err
	}
	return nil
}
