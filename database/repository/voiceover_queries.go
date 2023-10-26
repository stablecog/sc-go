package repository

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/database/ent/voiceover"
	"github.com/stablecog/sc-go/database/ent/voiceoveroutput"
	"github.com/stablecog/sc-go/database/ent/voiceoverspeaker"
	"github.com/stablecog/sc-go/log"
	"github.com/stablecog/sc-go/server/requests"
	"github.com/stablecog/sc-go/utils"
)

func (r *Repository) GetAllVoiceoverModels() ([]*ent.VoiceoverModel, error) {
	models, err := r.DB.VoiceoverModel.Query().All(r.Ctx)
	if err != nil {
		return nil, err
	}

	return models, nil
}

func (r *Repository) GetAllVoiceoverSpeakers() ([]*ent.VoiceoverSpeaker, error) {
	speakers, err := r.DB.VoiceoverSpeaker.Query().All(r.Ctx)
	if err != nil {
		return nil, err
	}

	return speakers, nil
}

func (r *Repository) GetVoiceover(id uuid.UUID) (*ent.Voiceover, error) {
	return r.DB.Voiceover.Query().Where(voiceover.ID(id)).Only(r.Ctx)
}

func (r *Repository) GetVoiceoversQueuedOrStarted() ([]*ent.Voiceover, error) {
	// Get voiceovers that are started/queued and older than 5 minutes
	return r.DB.Voiceover.Query().
		Where(
			voiceover.StatusIn(
				voiceover.StatusQueued,
				voiceover.StatusStarted,
			),
			voiceover.CreatedAtLT(time.Now().Add(-5*time.Minute)),
		).
		Order(ent.Desc(voiceover.FieldCreatedAt)).
		Limit(100).
		All(r.Ctx)
}

// Apply all filters to root ent query
func (r *Repository) ApplyUserVoiceoverFilters(query *ent.VoiceoverQuery, filters *requests.QueryVoiceoverFilters, omitEdges bool) *ent.VoiceoverQuery {
	resQuery := query
	if filters != nil {

		if !omitEdges {
			if filters.IsFavorited != nil {
				resQuery = resQuery.Where(func(s *sql.Selector) {
					s.Where(sql.EQ(voiceoveroutput.FieldIsFavorited, *filters.IsFavorited))
				})
			}
		}

		// prompt ID
		if filters.PromptID != nil {
			resQuery = resQuery.Where(voiceover.PromptIDEQ(*filters.PromptID))
		}

		// Start dt
		if filters.StartDt != nil {
			resQuery = resQuery.Where(voiceover.CreatedAtGTE(*filters.StartDt))
		}

		// End dt
		if filters.EndDt != nil {
			resQuery = resQuery.Where(voiceover.CreatedAtLTE(*filters.EndDt))
		}

		if filters.WasAutoSubmitted != nil {
			resQuery = resQuery.Where(voiceover.WasAutoSubmittedEQ(*filters.WasAutoSubmitted))
		}
	}
	return resQuery
}

// Gets the count of voiceovers with outputs user has with filters
func (r *Repository) GetVoiceoverCount(filters *requests.QueryVoiceoverFilters) (int, error) {
	var query *ent.VoiceoverQuery

	query = r.DB.Voiceover.Query().
		Where(voiceover.StatusEQ(voiceover.StatusSucceeded))
	if filters.UserID != nil {
		query = query.Where(voiceover.UserID(*filters.UserID))
	}

	// Exclude deleted at always
	query = query.Where(func(s *sql.Selector) {
		s.Where(sql.IsNull("deleted_at"))
	})

	// Apply filters
	query = r.ApplyUserVoiceoverFilters(query, filters, false)

	// Join other data
	var res []UserGenCount
	err := query.Modify(func(s *sql.Selector) {
		pt := sql.Table(prompt.Table)
		vot := sql.Table(voiceoveroutput.Table)
		s.LeftJoin(pt).On(
			s.C(voiceover.FieldPromptID), pt.C(prompt.FieldID),
		).LeftJoin(vot).On(
			s.C(voiceover.FieldID), vot.C(voiceoveroutput.FieldVoiceoverID),
		).Select(sql.As(sql.Count("*"), "total"))
	}).Scan(r.Ctx, &res)
	if err != nil {
		return 0, err
	} else if len(res) == 0 {
		return 0, nil
	}
	return res[0].Total, nil
}

// Get user voiceovers from the database using page options
// Cursor actually represents created_at, we paginate using this for performance reasons
// If present, we will get results after the cursor (anything before, represents previous pages)
// ! using ent .With... doesn't use joins, so we construct our own query to make it more efficient
func (r *Repository) QueryVoiceovers(per_page int, cursor *time.Time, filters *requests.QueryVoiceoverFilters) (*VoiceoverQueryWithOutputsMeta, error) {
	// Base fields to select in our query
	selectFields := []string{
		voiceover.FieldID,
		voiceover.FieldSeed,
		voiceover.FieldTemperature,
		voiceover.FieldStatus,
		voiceover.FieldSpeakerID,
		voiceover.FieldModelID,
		voiceover.FieldPromptID,
		voiceover.FieldCreatedAt,
		voiceover.FieldUpdatedAt,
		voiceover.FieldStartedAt,
		voiceover.FieldCompletedAt,
		voiceover.FieldWasAutoSubmitted,
		voiceover.FieldDenoiseAudio,
		voiceover.FieldRemoveSilence,
	}
	var query *ent.VoiceoverQuery
	var vQueryResult []VoiceoverQueryWithOutputsResult

	// Figure out order bys
	var orderByVoiceover []string
	var orderByOutput []string
	if filters == nil || (filters != nil && filters.OrderBy == requests.OrderByCreatedAt) {
		orderByVoiceover = []string{voiceover.FieldCreatedAt}
		orderByOutput = []string{voiceoveroutput.FieldCreatedAt}
	} else {
		orderByVoiceover = []string{voiceover.FieldCreatedAt, voiceover.FieldUpdatedAt}
		orderByOutput = []string{voiceoveroutput.FieldCreatedAt, voiceoveroutput.FieldUpdatedAt}
	}

	query = r.DB.Voiceover.Query().Select(selectFields...).
		Where(voiceover.StatusEQ(voiceover.StatusSucceeded))
	if filters.UserID != nil {
		query = query.Where(voiceover.UserID(*filters.UserID))
	}
	if cursor != nil {
		query = query.Where(voiceover.CreatedAtLT(*cursor))
	}

	// Exclude deleted at always
	query = query.Where(func(s *sql.Selector) {
		s.Where(sql.IsNull("deleted_at"))
	})

	// Apply filters
	query = r.ApplyUserVoiceoverFilters(query, filters, false)

	// Limits is + 1 so we can check if there are more pages
	query = query.Limit(per_page + 1)

	// Join other data
	err := query.Modify(func(s *sql.Selector) {
		vt := sql.Table(voiceover.Table)
		pt := sql.Table(prompt.Table)
		vot := sql.Table(voiceoveroutput.Table)
		st := sql.Table(voiceoverspeaker.Table)
		s.LeftJoin(pt).On(
			s.C(voiceover.FieldPromptID), pt.C(prompt.FieldID),
		).LeftJoin(vot).On(
			s.C(voiceover.FieldID), vot.C(voiceoveroutput.FieldVoiceoverID),
		).LeftJoin(st).On(
			s.C(voiceover.FieldSpeakerID), st.C(voiceoverspeaker.FieldID),
		).AppendSelect(sql.As(pt.C(prompt.FieldText), "prompt_text"), sql.As(vot.C(voiceoveroutput.FieldID), "output_id"), sql.As(vot.C(voiceoveroutput.FieldAudioPath), "audio_path"), sql.As(vot.C(voiceoveroutput.FieldVideoPath), "video_path"), sql.As(vot.C(voiceoveroutput.FieldAudioArray), "audio_array"), sql.As(vot.C(voiceoveroutput.FieldDeletedAt), "deleted_at"), sql.As(vot.C(voiceoveroutput.FieldIsFavorited), "is_favorited"), sql.As(vot.C(voiceoveroutput.FieldAudioDuration), "audio_duration"), sql.As(st.C(voiceoverspeaker.FieldNameInWorker), "name_in_worker"), sql.As(st.C(voiceoverspeaker.FieldLocale), "locale")).
			GroupBy(s.C(voiceover.FieldID), pt.C(prompt.FieldText),
				vot.C(voiceoveroutput.FieldID), vot.C(voiceoveroutput.FieldAudioPath), vot.C(voiceoveroutput.FieldVideoPath), vot.C(voiceoveroutput.FieldAudioArray), vot.C(voiceoveroutput.FieldAudioDuration),
				st.C(voiceoverspeaker.FieldNameInWorker), st.C(voiceoverspeaker.FieldLocale))
		orderDir := "asc"
		if filters == nil || (filters != nil && filters.Order == requests.SortOrderDescending) {
			orderDir = "desc"
		}
		var orderByVoiceover2 []string
		var orderByOutput2 []string
		for _, o := range orderByVoiceover {
			if orderDir == "desc" {
				orderByVoiceover2 = append(orderByVoiceover2, sql.Desc(vt.C(o)))
			} else {
				orderByVoiceover2 = append(orderByVoiceover2, sql.Asc(vt.C(o)))
			}
		}
		for _, o := range orderByOutput {
			if orderDir == "desc" {
				orderByOutput2 = append(orderByOutput2, sql.Desc(vot.C(o)))
			} else {
				orderByOutput2 = append(orderByOutput2, sql.Asc(vot.C(o)))
			}
		}
		// Order by voiceover, then output
		orderByCombined := append(orderByVoiceover2, orderByOutput2...)
		s.OrderBy(orderByCombined...)
	}).Scan(r.Ctx, &vQueryResult)

	if err != nil {
		log.Error("Error getting user voiceovers", "err", err)
		return nil, err
	}

	if len(vQueryResult) == 0 {
		meta := &VoiceoverQueryWithOutputsMeta{
			Outputs: []VoiceoverQueryWithOutputsResultFormatted{},
		}
		// Only give total if we have no cursor
		if cursor == nil {
			zero := 0
			meta.Total = &zero
		}
		return meta, nil
	}

	meta := &VoiceoverQueryWithOutputsMeta{}
	if len(vQueryResult) > per_page {
		// Remove last item
		vQueryResult = vQueryResult[:len(vQueryResult)-1]
		meta.Next = &vQueryResult[len(vQueryResult)-1].CreatedAt
	}

	// Get real URLs for each
	for i, v := range vQueryResult {
		if v.AudioFileUrl != "" {
			parsed := utils.GetEnv().GetURLFromAudioFilePath(v.AudioFileUrl)
			vQueryResult[i].AudioFileUrl = parsed
		}
		if v.VideoFileUrl != "" {
			parsed := utils.GetEnv().GetURLFromAudioFilePath(v.VideoFileUrl)
			vQueryResult[i].VideoFileUrl = parsed
		}
	}

	// Format to VoiceoverQueryWithOutputsResultFormatted
	voiceoverOutputMap := make(map[uuid.UUID][]GenerationUpscaleOutput)
	for _, v := range vQueryResult {
		if v.OutputID == nil {
			log.Warn("Output ID is nil for voiceover, cannot include in result", "id", v.ID)
			continue
		}
		vOutput := GenerationUpscaleOutput{
			ID:               *v.OutputID,
			AudioFileUrl:     v.AudioFileUrl,
			VideoFileUrl:     v.VideoFileUrl,
			AudioArray:       v.AudioArray,
			WasAutoSubmitted: v.WasAutoSubmitted,
			IsFavorited:      v.IsFavorited,
			AudioDuration:    utils.ToPtr(v.AudioDuration),
		}
		var speaker VoiceoverSpeaker
		if v.SpeakerID != nil {
			speaker = VoiceoverSpeaker{
				ID:     *v.SpeakerID,
				Name:   v.NameInWorker,
				Locale: v.Locale,
			}
		}
		output := VoiceoverQueryWithOutputsResultFormatted{
			GenerationUpscaleOutput: vOutput,
			Voiceover: VoiceoverQueryWithOutputsData{
				ID:          v.ID,
				Seed:        v.Seed,
				Status:      v.Status,
				Temperature: v.Temperature,
				ModelID:     v.ModelID,
				CreatedAt:   v.CreatedAt,
				UpdatedAt:   v.UpdatedAt,
				StartedAt:   v.StartedAt,
				CompletedAt: v.CompletedAt,
				Prompt: PromptType{
					Text: v.PromptText,
					ID:   *v.PromptID,
				},
				IsFavorited:   vOutput.IsFavorited,
				Speaker:       &speaker,
				DenoiseAudio:  v.DenoiseAudio,
				RemoveSilence: v.RemoveSilence,
			},
		}
		voiceoverOutputMap[v.ID] = append(voiceoverOutputMap[v.ID], vOutput)
		meta.Outputs = append(meta.Outputs, output)
	}
	// Now loop through and add outputs to each voiceover
	for i, g := range meta.Outputs {
		meta.Outputs[i].Voiceover.Outputs = voiceoverOutputMap[g.Voiceover.ID]
	}

	if cursor == nil {
		total, err := r.GetVoiceoverCount(filters)
		if err != nil {
			log.Error("Error getting user voiceover count", "err", err)
			return nil, err
		}
		meta.Total = &total
	}

	return meta, err
}

func (r *Repository) GetVoiceverSpeakersWithName(limit int) ([]*ent.VoiceoverSpeaker, error) {
	// Only get english for now
	return r.DB.VoiceoverSpeaker.Query().Where(voiceoverspeaker.NameNotNil(), voiceoverspeaker.IsActiveEQ(true), voiceoverspeaker.LocaleEQ("en")).Limit(limit).Order(ent.Desc(voiceoverspeaker.FieldIsDefault)).All(r.Ctx)
}

type VoiceoverQueryWithOutputsResult struct {
	OutputID      *uuid.UUID `json:"output_id,omitempty" sql:"output_id"`
	AudioFileUrl  string     `json:"audio_file_url,omitempty" sql:"audio_path"`
	VideoFileUrl  string     `json:"video_file_url,omitempty" sql:"video_path"`
	AudioArray    []float64  `json:"audio_array,omitempty" sql:"audio_array"`
	AudioDuration float32    `json:"audio_duration" sql:"audio_duration"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" sql:"deleted_at"`
	VoiceoverQueryWithOutputsData
}

type VoiceoverSpeaker struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Locale string    `json:"locale"`
}

type VoiceoverQueryWithOutputsData struct {
	ID               uuid.UUID                 `json:"id" sql:"id"`
	Seed             int                       `json:"seed" sql:"seed"`
	Temperature      float32                   `json:"temperature" sql:"temperature"`
	Status           string                    `json:"status" sql:"status"`
	ModelID          uuid.UUID                 `json:"model_id" sql:"model_id"`
	PromptID         *uuid.UUID                `json:"prompt_id,omitempty" sql:"prompt_id"`
	CreatedAt        time.Time                 `json:"created_at" sql:"created_at"`
	UpdatedAt        time.Time                 `json:"updated_at" sql:"updated_at"`
	StartedAt        *time.Time                `json:"started_at,omitempty" sql:"started_at"`
	CompletedAt      *time.Time                `json:"completed_at,omitempty" sql:"completed_at"`
	PromptText       string                    `json:"prompt_text,omitempty" sql:"prompt_text"`
	IsFavorited      bool                      `json:"is_favorited" sql:"is_favorited"`
	Outputs          []GenerationUpscaleOutput `json:"outputs"`
	Prompt           PromptType                `json:"prompt"`
	WasAutoSubmitted bool                      `json:"was_auto_submitted" sql:"was_auto_submitted"`
	DenoiseAudio     bool                      `json:"denoise_audio" sql:"denoise_audio"`
	RemoveSilence    bool                      `json:"remove_silence" sql:"remove_silence"`
	Speaker          *VoiceoverSpeaker         `json:"speaker,omitempty"`
	// For speaker object
	SpeakerID    *uuid.UUID `json:"speaker_id,omitempty" sql:"speaker_id"`
	NameInWorker string     `json:"name_in_worker,omitempty" sql:"name_in_worker"`
	Locale       string     `json:"locale,omitempty" sql:"locale"`
}

type VoiceoverQueryWithOutputsResultFormatted struct {
	GenerationUpscaleOutput
	Voiceover VoiceoverQueryWithOutputsData `json:"voiceover"`
}

// Paginated meta for querying voiceovers
type VoiceoverQueryWithOutputsMeta struct {
	Total   *int                                       `json:"total_count,omitempty"`
	Outputs []VoiceoverQueryWithOutputsResultFormatted `json:"outputs"`
	Next    *time.Time                                 `json:"next,omitempty"`
}
