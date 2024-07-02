package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/ent/generationoutputlike"
	"github.com/stablecog/sc-go/database/ent/negativeprompt"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
)

// Retrieved a single generation output by ID, in GalleryData format
func (r *Repository) RetrieveGalleryDataByID(id uuid.UUID, userId *uuid.UUID, callingUserId *uuid.UUID, all bool) (*GalleryData, error) {
	var q *ent.GenerationOutputQuery
	if userId != nil {
		q = r.DB.Generation.Query().Where(generation.UserIDEQ(*userId)).QueryGenerationOutputs()
	} else {
		q = r.DB.GenerationOutput.Query()
	}
	q = q.Where(generationoutput.IDEQ(id))
	if !all {
		q = q.Where(generationoutput.GalleryStatusEQ(generationoutput.GalleryStatusApproved))
	}
	if callingUserId != nil {
		q = q.WithGenerationOutputLikes(func(gql *ent.GenerationOutputLikeQuery) {
			gql.Where(generationoutputlike.LikedByUserID(*callingUserId))
		})
	}
	output, err := q.WithGenerations(func(gq *ent.GenerationQuery) {
		gq.WithPrompt()
		gq.WithNegativePrompt()
		gq.WithUser()
	}).Only(r.Ctx)
	if err != nil {
		return nil, err
	}
	data := GalleryData{
		ID:             output.ID,
		ImageURL:       utils.GetEnv().GetURLFromImagePath(output.ImagePath),
		CreatedAt:      output.CreatedAt,
		UpdatedAt:      output.UpdatedAt,
		Width:          output.Edges.Generations.Width,
		Height:         output.Edges.Generations.Height,
		InferenceSteps: output.Edges.Generations.InferenceSteps,
		GuidanceScale:  output.Edges.Generations.GuidanceScale,
		Seed:           output.Edges.Generations.Seed,
		ModelID:        output.Edges.Generations.ModelID,
		SchedulerID:    output.Edges.Generations.SchedulerID,
		PromptID:       output.Edges.Generations.Edges.Prompt.ID,
		PromptText:     output.Edges.Generations.Edges.Prompt.Text,
		PromptStrength: output.Edges.Generations.PromptStrength,
		User: &UserType{
			Username:   output.Edges.Generations.Edges.User.Username,
			Identifier: utils.Sha256(output.Edges.Generations.Edges.User.ID.String()),
		},
		LikeCount: output.LikeCount,
		IsLiked:   utils.ToPtr(len(output.Edges.GenerationOutputLikes) > 0),
	}
	if all {
		data.IsPublic = output.IsPublic
		data.WasAutoSubmitted = output.Edges.Generations.WasAutoSubmitted
	}
	if output.Edges.Generations.Edges.NegativePrompt != nil {
		data.NegativePromptID = &output.Edges.Generations.Edges.NegativePrompt.ID
		data.NegativePromptText = output.Edges.Generations.Edges.NegativePrompt.Text
	}
	if output.UpscaledImagePath != nil {
		data.UpscaledImageURL = utils.GetEnv().GetURLFromImagePath(*output.UpscaledImagePath)
	}
	return &data, nil
}

func (r *Repository) RetrieveMostRecentGalleryDataV3(filters *requests.QueryGenerationFilters, callingUserId *uuid.UUID, per_page int, cursor *time.Time, offset *int) ([]GalleryData, *time.Time, *int, error) {
	// Base query parts
	baseQuery := `
    WITH like_counts AS (
        SELECT 
            output_id, 
            COUNT(*) AS like_count_trending 
        FROM 
            generation_output_likes 
        WHERE 
            created_at > $1 
        GROUP BY 
            output_id
    )
    SELECT 
        go.id AS id, 
        go.image_path AS image_url,
        go.upscaled_image_path AS upscaled_image_url,
        go.created_at,
        go.updated_at,
        g.width, 
        g.height, 
        g.inference_steps, 
        g.guidance_scale, 
        g.seed, 
        g.model_id, 
        g.scheduler_id, 
        p.text AS prompt_text,
        g.prompt_id,
        np.text AS negative_prompt_text,
        g.negative_prompt_id,
        u.id AS user_id,
        u.username,
        g.prompt_strength, 
        g.was_auto_submitted, 
        go.is_public, 
        go.like_count, 
        COALESCE(lc.like_count_trending, 0) AS like_count_trending 
    FROM 
        generations g
    JOIN 
        generation_outputs go 
        ON g.id = go.generation_id 
        AND go.deleted_at IS NULL 
    LEFT JOIN 
        like_counts lc 
        ON go.id = lc.output_id 
    LEFT JOIN 
        users u 
        ON g.user_id = u.id 
    LEFT JOIN 
        prompts p
        ON g.prompt_id = p.id
    LEFT JOIN 
        negative_prompts np
        ON g.negative_prompt_id = np.id
    WHERE 
        g.status = $2 
        AND go.gallery_status = $3 
        AND go.is_public`

	// Arguments for the query
	args := []interface{}{
		time.Now().AddDate(0, 0, -7), // for like_counts CTE
		"succeeded",                  // status
		"approved",                   // gallery_status
	}

	// Apply the username filter if it exists
	if len(filters.Username) > 0 {
		placeholders := make([]string, len(filters.Username))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", len(args)+i+1)
		}
		baseQuery += fmt.Sprintf(" AND u.username IN (%s)", strings.Join(placeholders, ","))
		for _, username := range filters.Username {
			args = append(args, username)
		}
	}

	// Apply the model_ids filter if it exists
	if len(filters.ModelIDs) > 0 {
		placeholders := make([]string, len(filters.ModelIDs))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", len(args)+i+1)
		}
		baseQuery += fmt.Sprintf(" AND g.model_id IN (%s)", strings.Join(placeholders, ","))
		for _, modelID := range filters.ModelIDs {
			args = append(args, modelID)
		}
	}

	// Apply the scheduler_ids filter if it exists
	if len(filters.SchedulerIDs) > 0 {
		placeholders := make([]string, len(filters.SchedulerIDs))
		for i := range placeholders {
			placeholders[i] = fmt.Sprintf("$%d", len(args)+i+1)
		}
		baseQuery += fmt.Sprintf(" AND g.scheduler_id IN (%s)", strings.Join(placeholders, ","))
		for _, schedulerID := range filters.SchedulerIDs {
			args = append(args, schedulerID)
		}
	}

	// Apply the aspect ratio filter if it exists
	if len(filters.AspectRatio) > 0 {
		var widthHeightConditions []string
		for _, aspectRatio := range filters.AspectRatio {
			widths, heights := aspectRatio.GetAllWidthHeightCombos()
			for i := 0; i < len(widths); i++ {
				if i < len(heights) {
					condition := fmt.Sprintf("(g.width = %d AND g.height = %d)", widths[i], heights[i])
					widthHeightConditions = append(widthHeightConditions, condition)
				}
			}
		}
		if len(widthHeightConditions) > 0 {
			baseQuery += " AND (" + strings.Join(widthHeightConditions, " OR ") + ")"
		}
	}

	// Determine the order direction
	orderDir := "asc"
	if filters == nil || (filters != nil && filters.Order == requests.SortOrderDescending) {
		orderDir = "desc"
	}

	// Construct the ORDER BY clause
	orderByClause := ""
	if filters != nil {
		if filters.OrderBy == requests.OrderByLikeCount {
			orderByClause = fmt.Sprintf("ORDER BY go.like_count %s", orderDir)
		} else if filters.OrderBy == requests.OrderByLikeCountTrending {
			orderByClause = fmt.Sprintf("ORDER BY like_count_trending %s", orderDir)
		} else {
			orderByClause = fmt.Sprintf("ORDER BY g.created_at %s, go.created_at %s", orderDir, orderDir)
		}
	} else {
		orderByClause = fmt.Sprintf("ORDER BY g.created_at %s, go.created_at %s", orderDir, orderDir)
	}

	// Add the ORDER BY clause and LIMIT
	baseQuery += fmt.Sprintf(" %s LIMIT $4", orderByClause)

	// Add the limit argument
	args = append(args, per_page+1) // +1 to check if there's more data for pagination

	// Apply cursor for pagination if ordering by created_at
	if filters == nil || (filters != nil && (filters.OrderBy != requests.OrderByLikeCount && filters.OrderBy != requests.OrderByLikeCountTrending)) {
		if cursor != nil {
			baseQuery += " AND g.created_at < $5"
			args = append(args, *cursor)
		}
	} else if offset != nil {
		baseQuery += " OFFSET $5"
		args = append(args, *offset)
	}

	// Execute the query
	rows, err := r.DB.QueryContext(context.Background(), baseQuery, args...)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var results []GalleryData
	for rows.Next() {
		var data GalleryData
		var userID, promptID, negativePromptID uuid.UUID
		var username sql.NullString
		var negativePromptText sql.NullString
		var likeCountTrending sql.NullInt64
		var promptStrength sql.NullFloat64

		if err := rows.Scan(
			&data.ID,
			&data.ImageURL,
			&data.UpscaledImageURL,
			&data.CreatedAt,
			&data.UpdatedAt,
			&data.Width,
			&data.Height,
			&data.InferenceSteps,
			&data.GuidanceScale,
			&data.Seed,
			&data.ModelID,
			&data.SchedulerID,
			&data.PromptText,
			&promptID,
			&negativePromptText,
			&negativePromptID,
			&userID,
			&username,
			&promptStrength,
			&data.WasAutoSubmitted,
			&data.IsPublic,
			&data.LikeCount,
			&likeCountTrending,
		); err != nil {
			return nil, nil, nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if username.Valid {
			data.Username = &username.String
		}

		if likeCountTrending.Valid {
			count := int(likeCountTrending.Int64)
			data.LikeCountTrending = &count
		}

		if promptStrength.Valid {
			strength := float32(promptStrength.Float64)
			data.PromptStrength = &strength
		}

		if negativePromptText.Valid {
			data.NegativePromptText = negativePromptText.String
		}

		data.PromptID = promptID
		data.NegativePromptID = &negativePromptID
		data.UserID = &userID

		results = append(results, data)
	}

	// Handle pagination
	var nextCursor *time.Time
	var nextOffset *int
	if len(results) > per_page {
		results = results[:len(results)-1]
		if filters != nil && (filters.OrderBy == requests.OrderByLikeCount || filters.OrderBy == requests.OrderByLikeCountTrending) {
			if offset == nil {
				nextOffset = new(int)
				*nextOffset = len(results)
			} else {
				nextOffset = new(int)
				*nextOffset = *offset + len(results)
			}
		} else {
			nextCursor = &results[len(results)-1].CreatedAt
		}
	}

	return results, nextCursor, nextOffset, nil
}

func (r *Repository) RetrieveMostRecentGalleryDataV2(filters *requests.QueryGenerationFilters, callingUserId *uuid.UUID, per_page int, cursor *time.Time, offset *int) ([]GalleryData, *time.Time, *int, error) {
	// Base fields to select in our query
	selectFields := []string{
		generation.FieldID,
		generation.FieldWidth,
		generation.FieldHeight,
		generation.FieldInferenceSteps,
		generation.FieldSeed,
		generation.FieldStatus,
		generation.FieldGuidanceScale,
		generation.FieldSchedulerID,
		generation.FieldModelID,
		generation.FieldPromptID,
		generation.FieldNegativePromptID,
		generation.FieldCreatedAt,
		generation.FieldUpdatedAt,
		generation.FieldStartedAt,
		generation.FieldCompletedAt,
		generation.FieldWasAutoSubmitted,
		generation.FieldInitImageURL,
		generation.FieldPromptStrength,
	}
	var query *ent.GenerationQuery
	var gQueryResult []GenerationQueryWithOutputsResult

	// Figure out order bys
	var orderByGeneration []string
	var orderByOutput []string
	if filters == nil || (filters != nil && (filters.OrderBy == requests.OrderByCreatedAt || filters.OrderBy == requests.OrderByLikeCount || filters.OrderBy == requests.OrderByLikeCountTrending)) {
		orderByGeneration = []string{generation.FieldCreatedAt}
		orderByOutput = []string{generationoutput.FieldCreatedAt}
	} else {
		orderByGeneration = []string{generation.FieldCreatedAt, generation.FieldUpdatedAt}
		orderByOutput = []string{generationoutput.FieldCreatedAt, generationoutput.FieldUpdatedAt}
	}

	query = r.DB.Debug().Generation.Query().Select(selectFields...).
		Where(generation.StatusEQ(generation.StatusSucceeded))
	if cursor != nil {
		query = query.Where(generation.CreatedAtLT(*cursor))
	}
	if offset != nil {
		query = query.Offset(*offset)
	}

	// Apply filters
	query = r.ApplyUserGenerationsFilters(query, filters, false)

	// Limits is + 1 so we can check if there are more pages
	query = query.Limit(per_page + 1)

	// Create the subquery for likes count within last 7 days
	likeT := sql.Table(generationoutputlike.Table)
	likeSubQuery := sql.Select(
		sql.As(likeT.C(generationoutputlike.FieldOutputID), "output_id"),
		sql.As(sql.Count("*"), "like_count_trending"),
	).From(likeT).
		Where(
			sql.GT(likeT.C(generationoutputlike.FieldCreatedAt), time.Now().AddDate(0, 0, -7)),
		).
		GroupBy(likeT.C(generationoutputlike.FieldOutputID))

	// Modify the main query to use the CTE
	err := query.Modify(func(s *sql.Selector) {
		gt := sql.Table(generation.Table)
		got := sql.Table(generationoutput.Table)
		ut := sql.Table(user.Table)

		// Join generation_outputs table
		s.Join(got).OnP(
			sql.And(
				sql.ColumnsEQ(gt.C(generation.FieldID), got.C(generationoutput.FieldGenerationID)),
				sql.IsNull(got.C(generationoutput.FieldDeletedAt)),
			),
		)

		// Left join the like_subquery
		s.LeftJoin(likeSubQuery.As("like_subquery")).OnP(
			sql.ColumnsEQ(got.C(generationoutput.FieldID), sql.Table("like_subquery").C("output_id")),
		)

		// Join users table if filters are applied
		if filters != nil && filters.UserID != nil {
			s.Join(ut).OnP(
				sql.And(
					sql.ColumnsEQ(gt.C(generation.FieldUserID), ut.C(user.FieldID)),
					sql.EQ(ut.C(user.FieldID), *filters.UserID),
				),
			)
		} else if filters != nil && len(filters.Username) > 0 {
			v := make([]any, len(filters.Username))
			for i := range v {
				v[i] = filters.Username[i]
			}
			s.Join(ut).OnP(
				sql.And(
					sql.ColumnsEQ(gt.C(generation.FieldUserID), ut.C(user.FieldID)),
					sql.In(sql.Lower(ut.C(user.FieldUsername)), v...),
				),
			)
		} else {
			s.LeftJoin(ut).OnP(
				sql.ColumnsEQ(gt.C(generation.FieldUserID), ut.C(user.FieldID)),
			)
		}

		// Append necessary select fields
		s.AppendSelect(
			sql.As(got.C(generationoutput.FieldID), "output_id"),
			sql.As(got.C(generationoutput.FieldLikeCount), "like_count"),
			sql.As(got.C(generationoutput.FieldGalleryStatus), "output_gallery_status"),
			sql.As(got.C(generationoutput.FieldImagePath), "image_path"),
			sql.As(got.C(generationoutput.FieldUpscaledImagePath), "upscaled_image_path"),
			sql.As(got.C(generationoutput.FieldDeletedAt), "deleted_at"),
			sql.As(got.C(generationoutput.FieldIsFavorited), "is_favorited"),
			sql.As(ut.C(user.FieldUsername), "username"),
			sql.As(ut.C(user.FieldID), "user_id"),
			sql.As(got.C(generationoutput.FieldIsPublic), "is_public"),
			sql.As(fmt.Sprintf("coalesce(%s, 0)", sql.Table("like_subquery").C("like_count_trending")), "like_count_trending"),
		)

		// Group by necessary fields
		s.GroupBy(
			gt.C(generation.FieldID),
			gt.C(generation.FieldWidth),
			gt.C(generation.FieldHeight),
			gt.C(generation.FieldInferenceSteps),
			gt.C(generation.FieldSeed),
			gt.C(generation.FieldStatus),
			gt.C(generation.FieldGuidanceScale),
			gt.C(generation.FieldSchedulerID),
			gt.C(generation.FieldModelID),
			gt.C(generation.FieldPromptID),
			gt.C(generation.FieldNegativePromptID),
			gt.C(generation.FieldCreatedAt),
			gt.C(generation.FieldUpdatedAt),
			gt.C(generation.FieldStartedAt),
			gt.C(generation.FieldCompletedAt),
			gt.C(generation.FieldWasAutoSubmitted),
			gt.C(generation.FieldInitImageURL),
			gt.C(generation.FieldPromptStrength),
			got.C(generationoutput.FieldID),
			got.C(generationoutput.FieldGalleryStatus),
			got.C(generationoutput.FieldImagePath),
			got.C(generationoutput.FieldUpscaledImagePath),
			got.C(generationoutput.FieldLikeCount),
			got.C(generationoutput.FieldDeletedAt),
			got.C(generationoutput.FieldIsFavorited),
			got.C(generationoutput.FieldIsPublic),
			got.C(generationoutput.FieldCreatedAt),
			got.C(generationoutput.FieldUpdatedAt),
			ut.C(user.FieldUsername),
			ut.C(user.FieldID),
			sql.Table("like_subquery").C("like_count_trending"),
		)

		// Define ordering
		orderDir := "asc"
		if filters == nil || (filters != nil && filters.Order == requests.SortOrderDescending) {
			orderDir = "desc"
		}
		var orderByGeneration2 []string
		var orderByOutput2 []string
		for _, o := range orderByGeneration {
			if orderDir == "desc" {
				orderByGeneration2 = append(orderByGeneration2, sql.Desc(gt.C(o)))
			} else {
				orderByGeneration2 = append(orderByGeneration2, sql.Asc(gt.C(o)))
			}
		}
		for _, o := range orderByOutput {
			if orderDir == "desc" {
				orderByOutput2 = append(orderByOutput2, sql.Desc(got.C(o)))
			} else {
				orderByOutput2 = append(orderByOutput2, sql.Asc(got.C(o)))
			}
		}
		orderByLikes := []string{}
		if filters != nil && filters.OrderBy == requests.OrderByLikeCount {
			if orderDir == "desc" {
				orderByLikes = append(orderByLikes, sql.Desc(got.C(generationoutput.FieldLikeCount)))
			} else {
				orderByLikes = append(orderByLikes, sql.Asc(got.C(generationoutput.FieldLikeCount)))
			}
		}
		if filters != nil && filters.OrderBy == requests.OrderByLikeCountTrending {
			if orderDir == "desc" {
				orderByLikes = append(orderByLikes, sql.Desc("like_count_trending"))
			} else {
				orderByLikes = append(orderByLikes, sql.Asc("like_count_trending"))
			}
		}
		// Order by likes, generation, then output
		orderByCombined := append(orderByLikes, orderByGeneration2...)
		orderByCombined = append(orderByCombined, orderByOutput2...)
		s.OrderBy(orderByCombined...)
	}).Scan(r.Ctx, &gQueryResult)

	if err != nil {
		log.Error("Error retrieving generations", "err", err)
		return nil, nil, nil, err
	}

	if len(gQueryResult) == 0 {
		return []GalleryData{}, nil, nil, nil
	}

	// Get prompt texts
	promptIDsMap := make(map[uuid.UUID]string)
	negativePromptIdsMap := make(map[uuid.UUID]string)
	for _, g := range gQueryResult {
		if g.PromptID != nil {
			promptIDsMap[*g.PromptID] = ""
		}
		if g.NegativePromptID != nil {
			negativePromptIdsMap[*g.NegativePromptID] = ""
		}
	}
	promptIDs := make([]uuid.UUID, len(promptIDsMap))
	negativePromptId := make([]uuid.UUID, len(negativePromptIdsMap))

	i := 0
	for k := range promptIDsMap {
		promptIDs[i] = k
		i++
	}
	i = 0
	for k := range negativePromptIdsMap {
		negativePromptId[i] = k
		i++
	}

	prompts, err := r.DB.Prompt.Query().Select(prompt.FieldText).Where(prompt.IDIn(promptIDs...)).All(r.Ctx)
	if err != nil {
		log.Error("Error retrieving prompts", "err", err)
		return nil, nil, nil, err
	}
	negativePrompts, err := r.DB.NegativePrompt.Query().Select(negativeprompt.FieldText).Where(negativeprompt.IDIn(negativePromptId...)).All(r.Ctx)
	if err != nil {
		log.Error("Error retrieving prompts", "err", err)
		return nil, nil, nil, err
	}
	for _, p := range prompts {
		promptIDsMap[p.ID] = p.Text
	}
	for _, p := range negativePrompts {
		negativePromptIdsMap[p.ID] = p.Text
	}

	var nextCursor *time.Time
	var nextOffset *int
	if filters != nil && (filters.OrderBy == requests.OrderByLikeCountTrending || filters.OrderBy == requests.OrderByLikeCount) && len(gQueryResult) > per_page {
		if offset == nil {
			gQueryResult = gQueryResult[:len(gQueryResult)-1]
			nextOffset = utils.ToPtr(len(gQueryResult))
		} else {
			// Max offset
			if *offset < 50000 {
				gQueryResult = gQueryResult[:len(gQueryResult)-1]
				nextOffset = utils.ToPtr(*offset + len(gQueryResult))
			}
		}
	} else if len(gQueryResult) > per_page {
		gQueryResult = gQueryResult[:len(gQueryResult)-1]
		nextCursor = &gQueryResult[len(gQueryResult)-1].CreatedAt
	}

	// Figure out liked by in another query, if calling user is provided
	likedByMap := make(map[uuid.UUID]struct{})
	if callingUserId != nil && len(gQueryResult) > 0 {
		outputIds := make([]uuid.UUID, len(gQueryResult))
		for i, g := range gQueryResult {
			outputIds[i] = *g.OutputID
		}
		likedByMap, err = r.GetGenerationOutputsLikedByUser(*callingUserId, outputIds)
		if err != nil {
			log.Error("Error getting liked by map", "err", err)
			return nil, nil, nil, err
		}
	}

	galleryData := make([]GalleryData, len(gQueryResult))
	for i, g := range gQueryResult {
		likedByUser := false
		if _, ok := likedByMap[*g.OutputID]; ok {
			likedByUser = true
		}
		promptText, _ := promptIDsMap[*g.PromptID]
		galleryData[i] = GalleryData{
			ID:             *g.OutputID,
			ImageURL:       utils.GetEnv().GetURLFromImagePath(g.ImageUrl),
			CreatedAt:      g.CreatedAt,
			UpdatedAt:      g.UpdatedAt,
			Width:          g.Width,
			Height:         g.Height,
			InferenceSteps: g.InferenceSteps,
			GuidanceScale:  g.GuidanceScale,
			Seed:           g.Seed,
			ModelID:        g.ModelID,
			SchedulerID:    g.SchedulerID,
			PromptText:     promptText,
			PromptID:       *g.PromptID,
			PromptStrength: g.PromptStrength,
			User: &UserType{
				Username:   g.Username,
				Identifier: utils.Sha256(g.UserID.String()),
			},
			WasAutoSubmitted:  g.WasAutoSubmitted,
			IsPublic:          g.IsPublic,
			LikeCount:         g.LikeCount,
			LikeCountTrending: g.LikeCountTrending,
			IsLiked:           utils.ToPtr(likedByUser),
		}

		if g.NegativePromptID != nil {
			galleryData[i].NegativePromptText, _ = negativePromptIdsMap[*g.NegativePromptID]
			galleryData[i].NegativePromptID = g.NegativePromptID
		}

		if g.UpscaledImageUrl != "" {
			galleryData[i].UpscaledImageURL = utils.GetEnv().GetURLFromImagePath(g.UpscaledImageUrl)
		}
	}

	return galleryData, nextCursor, nextOffset, nil
}

// Retrieves data in gallery format given  output IDs
// Returns data, next cursor, error
func (r *Repository) RetrieveMostRecentGalleryData(filters *requests.QueryGenerationFilters, callingUserId *uuid.UUID, per_page int, cursor *time.Time) ([]GalleryData, *time.Time, error) {
	// Apply filters
	queryG := r.DB.Generation.Query().Where(
		generation.StatusEQ(generation.StatusSucceeded),
	)
	queryG = r.ApplyUserGenerationsFilters(queryG, filters, true)
	query := queryG.QueryGenerationOutputs().Where(
		generationoutput.DeletedAtIsNil(),
	)
	if cursor != nil {
		query = query.Where(generationoutput.CreatedAtLT(*cursor))
	}
	if filters != nil {
		if filters.UpscaleStatus == requests.UpscaleStatusNot {
			query = query.Where(generationoutput.UpscaledImagePathIsNil())
		}
		if filters.UpscaleStatus == requests.UpscaleStatusOnly {
			query = query.Where(generationoutput.UpscaledImagePathNotNil())
		}
		if len(filters.GalleryStatus) > 0 {
			query = query.Where(generationoutput.GalleryStatusIn(filters.GalleryStatus...))
		}
		if filters.IsPublic != nil {
			query = query.Where(generationoutput.IsPublic(*filters.IsPublic))
		}
	}
	if callingUserId != nil {
		query = query.WithGenerationOutputLikes(func(gql *ent.GenerationOutputLikeQuery) {
			gql.Where(generationoutputlike.LikedByUserID(*callingUserId))
		})
	}
	query = query.WithGenerations(func(s *ent.GenerationQuery) {
		s.WithPrompt()
		s.WithNegativePrompt()
		s.WithGenerationOutputs()
		s.WithUser()
	})

	// Limit
	query = query.Order(ent.Desc(generationoutput.FieldCreatedAt)).Limit(per_page + 1)

	res, err := query.All(r.Ctx)

	if err != nil {
		log.Errorf("Error retrieving gallery data: %v", err)
		return nil, nil, err
	}

	var nextCursor *time.Time
	if len(res) > per_page {
		res = res[:len(res)-1]
		nextCursor = &res[len(res)-1].CreatedAt
	}

	galleryData := make([]GalleryData, len(res))
	for i, output := range res {
		data := GalleryData{
			ID:             output.ID,
			ImageURL:       utils.GetEnv().GetURLFromImagePath(output.ImagePath),
			CreatedAt:      output.CreatedAt,
			UpdatedAt:      output.UpdatedAt,
			Width:          output.Edges.Generations.Width,
			Height:         output.Edges.Generations.Height,
			InferenceSteps: output.Edges.Generations.InferenceSteps,
			GuidanceScale:  output.Edges.Generations.GuidanceScale,
			Seed:           output.Edges.Generations.Seed,
			ModelID:        output.Edges.Generations.ModelID,
			SchedulerID:    output.Edges.Generations.SchedulerID,
			PromptText:     output.Edges.Generations.Edges.Prompt.Text,
			PromptID:       output.Edges.Generations.Edges.Prompt.ID,
			UserID:         &output.Edges.Generations.UserID,
			User: &UserType{
				Username:   output.Edges.Generations.Edges.User.Username,
				Identifier: utils.Sha256(output.Edges.Generations.Edges.User.ID.String()),
			},
			LikeCount: output.LikeCount,
			IsLiked:   utils.ToPtr(len(output.Edges.GenerationOutputLikes) > 0),
		}
		if output.UpscaledImagePath != nil {
			data.UpscaledImageURL = utils.GetEnv().GetURLFromImagePath(*output.UpscaledImagePath)
		}
		if output.Edges.Generations.Edges.NegativePrompt != nil {
			data.NegativePromptText = output.Edges.Generations.Edges.NegativePrompt.Text
			data.NegativePromptID = &output.Edges.Generations.Edges.NegativePrompt.ID
		}
		galleryData[i] = data
	}

	return galleryData, nextCursor, nil
}

// Retrieves data in gallery format given  output IDs
func (r *Repository) RetrieveGalleryDataWithOutputIDs(outputIDs []uuid.UUID, callingUserId *uuid.UUID, allIsPublic bool) ([]GalleryData, error) {
	q := r.DB.GenerationOutput.Query().Where(generationoutput.IDIn(outputIDs...))
	if allIsPublic {
		q = q.Where(generationoutput.IsPublic(true))
	} else {
		q = q.Where(generationoutput.GalleryStatusEQ(generationoutput.GalleryStatusApproved))
	}
	if callingUserId != nil {
		q = q.WithGenerationOutputLikes(func(gql *ent.GenerationOutputLikeQuery) {
			gql.Where(generationoutputlike.LikedByUserID(*callingUserId))
		})
	}
	res, err := q.
		WithGenerations(func(gq *ent.GenerationQuery) {
			gq.WithPrompt()
			gq.WithNegativePrompt()
			gq.WithUser()
		},
		).All(r.Ctx)
	if err != nil {
		return nil, err
	}

	galleryData := make([]GalleryData, len(res))
	for i, output := range res {
		data := GalleryData{
			ID:             output.ID,
			ImageURL:       utils.GetEnv().GetURLFromImagePath(output.ImagePath),
			CreatedAt:      output.CreatedAt,
			UpdatedAt:      output.UpdatedAt,
			Width:          output.Edges.Generations.Width,
			Height:         output.Edges.Generations.Height,
			InferenceSteps: output.Edges.Generations.InferenceSteps,
			GuidanceScale:  output.Edges.Generations.GuidanceScale,
			Seed:           output.Edges.Generations.Seed,
			ModelID:        output.Edges.Generations.ModelID,
			SchedulerID:    output.Edges.Generations.SchedulerID,
			PromptText:     output.Edges.Generations.Edges.Prompt.Text,
			PromptID:       output.Edges.Generations.Edges.Prompt.ID,
			UserID:         &output.Edges.Generations.UserID,
			User: &UserType{
				Username:   output.Edges.Generations.Edges.User.Username,
				Identifier: utils.Sha256(output.Edges.Generations.Edges.User.ID.String()),
			},
			LikeCount: output.LikeCount,
			IsLiked:   utils.ToPtr(len(output.Edges.GenerationOutputLikes) > 0),
			IsPublic:  output.IsPublic,
		}
		if output.UpscaledImagePath != nil {
			data.UpscaledImageURL = utils.GetEnv().GetURLFromImagePath(*output.UpscaledImagePath)
		}
		if output.Edges.Generations.Edges.NegativePrompt != nil {
			data.NegativePromptText = output.Edges.Generations.Edges.NegativePrompt.Text
			data.NegativePromptID = &output.Edges.Generations.Edges.NegativePrompt.ID
		}
		galleryData[i] = data
	}
	return galleryData, nil
}

type GalleryData struct {
	ID                 uuid.UUID  `json:"id,omitempty" sql:"id"`
	ImageURL           string     `json:"image_url"`
	UpscaledImageURL   string     `json:"upscaled_image_url,omitempty"`
	CreatedAt          time.Time  `json:"created_at" sql:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" sql:"updated_at"`
	Width              int32      `json:"width" sql:"generation_width"`
	Height             int32      `json:"height" sql:"generation_height"`
	InferenceSteps     int32      `json:"inference_steps" sql:"generation_inference_steps"`
	GuidanceScale      float32    `json:"guidance_scale" sql:"generation_guidance_scale"`
	Seed               int        `json:"seed,omitempty" sql:"generation_seed"`
	ModelID            uuid.UUID  `json:"model_id" sql:"model_id"`
	SchedulerID        uuid.UUID  `json:"scheduler_id" sql:"scheduler_id"`
	PromptText         string     `json:"prompt_text" sql:"prompt_text"`
	PromptID           uuid.UUID  `json:"prompt_id" sql:"prompt_id"`
	NegativePromptText string     `json:"negative_prompt_text,omitempty" sql:"negative_prompt_text"`
	NegativePromptID   *uuid.UUID `json:"negative_prompt_id,omitempty" sql:"negative_prompt_id"`
	UserID             *uuid.UUID `json:"user_id,omitempty" sql:"user_id"`
	Score              *float32   `json:"score,omitempty" sql:"score"`
	Username           *string    `json:"username,omitempty" sql:"username"`
	User               *UserType  `json:"user,omitempty" sql:"user"`
	PromptStrength     *float32   `json:"prompt_strength,omitempty" sql:"prompt_strength"`
	WasAutoSubmitted   bool       `json:"was_auto_submitted" sql:"was_auto_submitted"`
	IsPublic           bool       `json:"is_public" sql:"is_public"`
	LikeCount          int        `json:"like_count" sql:"like_count"`
	LikeCountTrending  *int       `json:"like_count_trending,omitempty" sql:"like_count_trending"`
	IsLiked            *bool      `json:"is_liked,omitempty" sql:"liked_by_user"`
}
