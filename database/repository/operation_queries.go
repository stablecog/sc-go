package repository

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/upscale"
	"github.com/stablecog/sc-go/database/ent/voiceover"
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
		generation.FieldFromDiscord,
	).Where(generation.StatusEQ(generation.StatusSucceeded), generation.UserIDEQ(userId), generation.StartedAtNotNil(), generation.CompletedAtNotNil())
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
		upscale.FieldFromDiscord,
	).Where(upscale.StatusEQ(upscale.StatusSucceeded), upscale.UserIDEQ(userId), upscale.StartedAtNotNil(), upscale.CompletedAtNotNil())
	if cursor != nil {
		uQuery = uQuery.Where(upscale.CreatedAtLT(*cursor))
	}
	uQuery = uQuery.Order(ent.Desc(upscale.FieldCreatedAt)).Limit(limit + 1)
	ups, err := uQuery.All(r.Ctx)
	if err != nil {
		log.Error("QueryUserOperations query upscales error", "err", err)
		return nil, err
	}

	// Query voiceovers
	voQuery := r.DB.Voiceover.Query().Select(
		voiceover.FieldID,
		voiceover.FieldStatus,
		voiceover.FieldUserID,
		voiceover.FieldCreatedAt,
		voiceover.FieldStartedAt,
		voiceover.FieldCompletedAt,
		voiceover.FieldAPITokenID,
		voiceover.FieldFromDiscord,
		voiceover.FieldCost,
	).Where(voiceover.StatusEQ(voiceover.StatusSucceeded), voiceover.UserIDEQ(userId), voiceover.StartedAtNotNil(), voiceover.CompletedAtNotNil())
	if cursor != nil {
		voQuery = voQuery.Where(voiceover.CreatedAtLT(*cursor))
	}
	voQuery = voQuery.Order(ent.Desc(voiceover.FieldCreatedAt)).Limit(limit + 1)
	vos, err := voQuery.All(r.Ctx)
	if err != nil {
		log.Error("QueryUserOperations query voiceovers error", "err", err)
		return nil, err
	}

	operationQueryResult := []OperationQueryResult{}
	for _, g := range gens {
		source := shared.OperationSourceTypeWebUI
		if g.APITokenID != nil {
			source = shared.OperationSourceTypeAPI
		} else if g.FromDiscord {
			source = shared.OperationSourceTypeDiscord
		}
		operationQueryResult = append(operationQueryResult, OperationQueryResult{
			ID:            g.ID,
			OperationType: shared.GENERATE,
			CreatedAt:     g.CreatedAt,
			StartedAt:     *g.StartedAt,
			CompletedAt:   *g.CompletedAt,
			Source:        source,
			NumOutputs:    int(g.NumOutputs),
			Cost:          g.NumOutputs,
		})
	}

	for _, u := range ups {
		source := shared.OperationSourceTypeWebUI
		if u.APITokenID != nil {
			source = shared.OperationSourceTypeAPI
		} else if u.FromDiscord {
			source = shared.OperationSourceTypeDiscord
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
			Cost:          1,
		})
	}

	for _, vo := range vos {
		source := shared.OperationSourceTypeWebUI
		if vo.APITokenID != nil {
			source = shared.OperationSourceTypeAPI
		} else if vo.FromDiscord {
			source = shared.OperationSourceTypeDiscord
		}
		// Is voiceover
		operationQueryResult = append(operationQueryResult, OperationQueryResult{
			ID:            vo.ID,
			OperationType: shared.VOICEOVER,
			CreatedAt:     vo.CreatedAt,
			StartedAt:     *vo.StartedAt,
			CompletedAt:   *vo.CompletedAt,
			Source:        source,
			NumOutputs:    1, // ! Always 1 for now
			Cost:          vo.Cost,
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
	Cost          int32                      `json:"cost"`
	Source        shared.OperationSourceType `json:"source"`
}

type OperationQueryResultMeta struct {
	Operations []OperationQueryResult `json:"operations"`
	Next       *time.Time             `json:"next,omitempty"`
}
