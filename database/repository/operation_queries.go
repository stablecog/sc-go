package repository

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/upscale"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/shared"
)

// Query generations and upscales combined for user
func (r *Repository) QueryUserOperations(userId uuid.UUID, limit int, cursor *time.Time) (*OperationQueryResultMeta, error) {
	// Generations first
	query := r.DB.Generation.Query().Select(
		generation.FieldID,
		generation.FieldStatus,
		generation.FieldUserID,
		generation.FieldCreatedAt,
		generation.FieldStartedAt,
		generation.FieldCompletedAt,
		generation.FieldAPITokenID,
		generation.FieldNumOutputs,
	).Where(generation.StatusEQ(generation.StatusSucceeded), generation.UserIDEQ(userId), generation.StartedAtNotNil(), generation.CompletedAtNotIn())
	if cursor != nil {
		query = query.Where(generation.CreatedAtLT(*cursor))
	}
	query = query.Order(ent.Desc(generation.FieldCreatedAt)).Limit(limit + 1)
	gens, err := query.All(r.Ctx)
	if err != nil {
		log.Error("QueryUserOperations query generations error", "err", err)
		return nil, err
	}

	// Query upscales
	uQuery := r.DB.Upscale.Query().Select(
		upscale.FieldID,
		upscale.FieldStatus,
		upscale.FieldUserID,
		upscale.FieldCreatedAt,
		upscale.FieldStartedAt,
		upscale.FieldCompletedAt,
		upscale.FieldAPITokenID,
	).Where(upscale.StatusEQ(upscale.StatusSucceeded), upscale.UserIDEQ(userId), upscale.StartedAtNotNil(), upscale.CompletedAtNotIn())
	if cursor != nil {
		uQuery = uQuery.Where(upscale.CreatedAtLT(*cursor))
	}
	uQuery = uQuery.Order(ent.Desc(upscale.FieldCreatedAt)).Limit(limit + 1)
	ups, err := uQuery.All(r.Ctx)
	if err != nil {
		log.Error("QueryUserOperations query upscales error", "err", err)
		return nil, err
	}

	operationQueryResult := []OperationQueryResult{}
	for _, g := range gens {
		source := shared.OperationSourceTypeWebUI
		if g.APITokenID != nil {
			source = shared.OperationSourceTypeAPI
		}
		operationQueryResult = append(operationQueryResult, OperationQueryResult{
			ID:            g.ID,
			OperationType: shared.GENERATE,
			CreatedAt:     g.CreatedAt,
			StartedAt:     *g.StartedAt,
			CompletedAt:   *g.CompletedAt,
			Source:        source,
			NumOutputs:    int(g.NumOutputs),
		})
	}

	for _, u := range ups {
		source := shared.OperationSourceTypeWebUI
		if u.APITokenID != nil {
			source = shared.OperationSourceTypeAPI
		}
		// Is upscale
		operationQueryResult = append(operationQueryResult, OperationQueryResult{
			ID:            u.ID,
			OperationType: shared.UPSCALE,
			CreatedAt:     u.CreatedAt,
			StartedAt:     *u.StartedAt,
			CompletedAt:   *u.CompletedAt,
			Source:        source,
			NumOutputs:    1, // ! Always 1 for now
		})
	}

	// Sort all operations by created at
	sort.Slice(operationQueryResult, func(i, j int) bool {
		return operationQueryResult[i].CreatedAt.After(operationQueryResult[j].CreatedAt)
	})

	// Truncate to limit + 1
	if len(operationQueryResult) > limit {
		operationQueryResult = operationQueryResult[:limit+1]
	}

	// Compute cursor
	var nextCursor *time.Time
	if len(operationQueryResult) > limit {
		nextCursor = &operationQueryResult[limit].CreatedAt
		operationQueryResult = operationQueryResult[:limit]
	}

	return &OperationQueryResultMeta{
		Operations: operationQueryResult,
		Next:       nextCursor,
	}, nil
}

type OperationQueryResultRaw struct {
	ID                 *uuid.UUID `json:"id" sql:"id"`
	Status             *string    `json:"status" sql:"status"`
	UserID             *uuid.UUID `json:"user_id" sql:"user_id"`
	CreatedAt          *time.Time `json:"created_at" sql:"created_at"`
	StartedAt          *time.Time `json:"started_at" sql:"started_at"`
	CompletedAt        *time.Time `json:"completed_at" sql:"completed_at"`
	ApiTokenID         *uuid.UUID `json:"api_token_id" sql:"api_token_id"`
	NumOutputs         *int       `json:"num_outputs" sql:"num_outputs"`
	UpscaleID          *uuid.UUID `json:"upscale_id" sql:"upscale_id"`
	UpscaleStatus      *string    `json:"upscale_status" sql:"upscale_status"`
	UpscaleUserID      *uuid.UUID `json:"upscale_user_id" sql:"upscale_user_id"`
	UpscaleCreatedAt   *time.Time `json:"upscale_created_at" sql:"upscale_created_at"`
	UpscaleStartedAt   *time.Time `json:"upscale_started_at" sql:"upscale_started_at"`
	UpscaleCompletedAt *time.Time `json:"upscale_completed_at" sql:"upscale_completed_at"`
	UpscaleApiTokenID  *uuid.UUID `json:"upscale_api_token_id" sql:"upscale_api_token_id"`
}

type OperationType string

const (
	OperationTypeGeneration OperationType = "generation"
	OperationTypeUpscale    OperationType = "upscale"
)

type OperationQueryResult struct {
	ID            uuid.UUID                  `json:"id"`
	OperationType shared.ProcessType         `json:"operation_type"`
	CreatedAt     time.Time                  `json:"created_at"`
	StartedAt     time.Time                  `json:"started_at"`
	CompletedAt   time.Time                  `json:"completed_at"`
	NumOutputs    int                        `json:"num_outputs"`
	Source        shared.OperationSourceType `json:"source"`
}

type OperationQueryResultMeta struct {
	Operations []OperationQueryResult `json:"operations"`
	Next       *time.Time             `json:"next,omitempty"`
}
