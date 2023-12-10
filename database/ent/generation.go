// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/apitoken"
	"github.com/stablecog/sc-go/database/ent/deviceinfo"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationmodel"
	"github.com/stablecog/sc-go/database/ent/negativeprompt"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/database/ent/scheduler"
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stablecog/sc-go/database/enttypes"
)

// Generation is the model entity for the Generation schema.
type Generation struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// Width holds the value of the "width" field.
	Width int32 `json:"width,omitempty"`
	// Height holds the value of the "height" field.
	Height int32 `json:"height,omitempty"`
	// InferenceSteps holds the value of the "inference_steps" field.
	InferenceSteps int32 `json:"inference_steps,omitempty"`
	// GuidanceScale holds the value of the "guidance_scale" field.
	GuidanceScale float32 `json:"guidance_scale,omitempty"`
	// NumOutputs holds the value of the "num_outputs" field.
	NumOutputs int32 `json:"num_outputs,omitempty"`
	// NsfwCount holds the value of the "nsfw_count" field.
	NsfwCount int32 `json:"nsfw_count,omitempty"`
	// Seed holds the value of the "seed" field.
	Seed int `json:"seed,omitempty"`
	// Status holds the value of the "status" field.
	Status generation.Status `json:"status,omitempty"`
	// FailureReason holds the value of the "failure_reason" field.
	FailureReason *string `json:"failure_reason,omitempty"`
	// CountryCode holds the value of the "country_code" field.
	CountryCode *string `json:"country_code,omitempty"`
	// InitImageURL holds the value of the "init_image_url" field.
	InitImageURL *string `json:"init_image_url,omitempty"`
	// MaskImageURL holds the value of the "mask_image_url" field.
	MaskImageURL *string `json:"mask_image_url,omitempty"`
	// PromptStrength holds the value of the "prompt_strength" field.
	PromptStrength *float32 `json:"prompt_strength,omitempty"`
	// WasAutoSubmitted holds the value of the "was_auto_submitted" field.
	WasAutoSubmitted bool `json:"was_auto_submitted,omitempty"`
	// StripeProductID holds the value of the "stripe_product_id" field.
	StripeProductID *string `json:"stripe_product_id,omitempty"`
	// SourceType holds the value of the "source_type" field.
	SourceType enttypes.SourceType `json:"source_type,omitempty"`
	// PromptID holds the value of the "prompt_id" field.
	PromptID *uuid.UUID `json:"prompt_id,omitempty"`
	// NegativePromptID holds the value of the "negative_prompt_id" field.
	NegativePromptID *uuid.UUID `json:"negative_prompt_id,omitempty"`
	// ModelID holds the value of the "model_id" field.
	ModelID uuid.UUID `json:"model_id,omitempty"`
	// SchedulerID holds the value of the "scheduler_id" field.
	SchedulerID uuid.UUID `json:"scheduler_id,omitempty"`
	// UserID holds the value of the "user_id" field.
	UserID uuid.UUID `json:"user_id,omitempty"`
	// DeviceInfoID holds the value of the "device_info_id" field.
	DeviceInfoID uuid.UUID `json:"device_info_id,omitempty"`
	// APITokenID holds the value of the "api_token_id" field.
	APITokenID *uuid.UUID `json:"api_token_id,omitempty"`
	// StartedAt holds the value of the "started_at" field.
	StartedAt *time.Time `json:"started_at,omitempty"`
	// CompletedAt holds the value of the "completed_at" field.
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the GenerationQuery when eager-loading is set.
	Edges GenerationEdges `json:"edges"`
}

// GenerationEdges holds the relations/edges for other nodes in the graph.
type GenerationEdges struct {
	// DeviceInfo holds the value of the device_info edge.
	DeviceInfo *DeviceInfo `json:"device_info,omitempty"`
	// Scheduler holds the value of the scheduler edge.
	Scheduler *Scheduler `json:"scheduler,omitempty"`
	// Prompt holds the value of the prompt edge.
	Prompt *Prompt `json:"prompt,omitempty"`
	// NegativePrompt holds the value of the negative_prompt edge.
	NegativePrompt *NegativePrompt `json:"negative_prompt,omitempty"`
	// GenerationModel holds the value of the generation_model edge.
	GenerationModel *GenerationModel `json:"generation_model,omitempty"`
	// User holds the value of the user edge.
	User *User `json:"user,omitempty"`
	// APITokens holds the value of the api_tokens edge.
	APITokens *ApiToken `json:"api_tokens,omitempty"`
	// GenerationOutputs holds the value of the generation_outputs edge.
	GenerationOutputs []*GenerationOutput `json:"generation_outputs,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [8]bool
}

// DeviceInfoOrErr returns the DeviceInfo value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationEdges) DeviceInfoOrErr() (*DeviceInfo, error) {
	if e.loadedTypes[0] {
		if e.DeviceInfo == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: deviceinfo.Label}
		}
		return e.DeviceInfo, nil
	}
	return nil, &NotLoadedError{edge: "device_info"}
}

// SchedulerOrErr returns the Scheduler value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationEdges) SchedulerOrErr() (*Scheduler, error) {
	if e.loadedTypes[1] {
		if e.Scheduler == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: scheduler.Label}
		}
		return e.Scheduler, nil
	}
	return nil, &NotLoadedError{edge: "scheduler"}
}

// PromptOrErr returns the Prompt value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationEdges) PromptOrErr() (*Prompt, error) {
	if e.loadedTypes[2] {
		if e.Prompt == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: prompt.Label}
		}
		return e.Prompt, nil
	}
	return nil, &NotLoadedError{edge: "prompt"}
}

// NegativePromptOrErr returns the NegativePrompt value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationEdges) NegativePromptOrErr() (*NegativePrompt, error) {
	if e.loadedTypes[3] {
		if e.NegativePrompt == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: negativeprompt.Label}
		}
		return e.NegativePrompt, nil
	}
	return nil, &NotLoadedError{edge: "negative_prompt"}
}

// GenerationModelOrErr returns the GenerationModel value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationEdges) GenerationModelOrErr() (*GenerationModel, error) {
	if e.loadedTypes[4] {
		if e.GenerationModel == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: generationmodel.Label}
		}
		return e.GenerationModel, nil
	}
	return nil, &NotLoadedError{edge: "generation_model"}
}

// UserOrErr returns the User value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationEdges) UserOrErr() (*User, error) {
	if e.loadedTypes[5] {
		if e.User == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: user.Label}
		}
		return e.User, nil
	}
	return nil, &NotLoadedError{edge: "user"}
}

// APITokensOrErr returns the APITokens value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationEdges) APITokensOrErr() (*ApiToken, error) {
	if e.loadedTypes[6] {
		if e.APITokens == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: apitoken.Label}
		}
		return e.APITokens, nil
	}
	return nil, &NotLoadedError{edge: "api_tokens"}
}

// GenerationOutputsOrErr returns the GenerationOutputs value or an error if the edge
// was not loaded in eager-loading.
func (e GenerationEdges) GenerationOutputsOrErr() ([]*GenerationOutput, error) {
	if e.loadedTypes[7] {
		return e.GenerationOutputs, nil
	}
	return nil, &NotLoadedError{edge: "generation_outputs"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Generation) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case generation.FieldPromptID, generation.FieldNegativePromptID, generation.FieldAPITokenID:
			values[i] = &sql.NullScanner{S: new(uuid.UUID)}
		case generation.FieldWasAutoSubmitted:
			values[i] = new(sql.NullBool)
		case generation.FieldGuidanceScale, generation.FieldPromptStrength:
			values[i] = new(sql.NullFloat64)
		case generation.FieldWidth, generation.FieldHeight, generation.FieldInferenceSteps, generation.FieldNumOutputs, generation.FieldNsfwCount, generation.FieldSeed:
			values[i] = new(sql.NullInt64)
		case generation.FieldStatus, generation.FieldFailureReason, generation.FieldCountryCode, generation.FieldInitImageURL, generation.FieldMaskImageURL, generation.FieldStripeProductID, generation.FieldSourceType:
			values[i] = new(sql.NullString)
		case generation.FieldStartedAt, generation.FieldCompletedAt, generation.FieldCreatedAt, generation.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case generation.FieldID, generation.FieldModelID, generation.FieldSchedulerID, generation.FieldUserID, generation.FieldDeviceInfoID:
			values[i] = new(uuid.UUID)
		default:
			return nil, fmt.Errorf("unexpected column %q for type Generation", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Generation fields.
func (ge *Generation) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case generation.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				ge.ID = *value
			}
		case generation.FieldWidth:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field width", values[i])
			} else if value.Valid {
				ge.Width = int32(value.Int64)
			}
		case generation.FieldHeight:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field height", values[i])
			} else if value.Valid {
				ge.Height = int32(value.Int64)
			}
		case generation.FieldInferenceSteps:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field inference_steps", values[i])
			} else if value.Valid {
				ge.InferenceSteps = int32(value.Int64)
			}
		case generation.FieldGuidanceScale:
			if value, ok := values[i].(*sql.NullFloat64); !ok {
				return fmt.Errorf("unexpected type %T for field guidance_scale", values[i])
			} else if value.Valid {
				ge.GuidanceScale = float32(value.Float64)
			}
		case generation.FieldNumOutputs:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field num_outputs", values[i])
			} else if value.Valid {
				ge.NumOutputs = int32(value.Int64)
			}
		case generation.FieldNsfwCount:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field nsfw_count", values[i])
			} else if value.Valid {
				ge.NsfwCount = int32(value.Int64)
			}
		case generation.FieldSeed:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field seed", values[i])
			} else if value.Valid {
				ge.Seed = int(value.Int64)
			}
		case generation.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				ge.Status = generation.Status(value.String)
			}
		case generation.FieldFailureReason:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field failure_reason", values[i])
			} else if value.Valid {
				ge.FailureReason = new(string)
				*ge.FailureReason = value.String
			}
		case generation.FieldCountryCode:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field country_code", values[i])
			} else if value.Valid {
				ge.CountryCode = new(string)
				*ge.CountryCode = value.String
			}
		case generation.FieldInitImageURL:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field init_image_url", values[i])
			} else if value.Valid {
				ge.InitImageURL = new(string)
				*ge.InitImageURL = value.String
			}
		case generation.FieldMaskImageURL:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field mask_image_url", values[i])
			} else if value.Valid {
				ge.MaskImageURL = new(string)
				*ge.MaskImageURL = value.String
			}
		case generation.FieldPromptStrength:
			if value, ok := values[i].(*sql.NullFloat64); !ok {
				return fmt.Errorf("unexpected type %T for field prompt_strength", values[i])
			} else if value.Valid {
				ge.PromptStrength = new(float32)
				*ge.PromptStrength = float32(value.Float64)
			}
		case generation.FieldWasAutoSubmitted:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field was_auto_submitted", values[i])
			} else if value.Valid {
				ge.WasAutoSubmitted = value.Bool
			}
		case generation.FieldStripeProductID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field stripe_product_id", values[i])
			} else if value.Valid {
				ge.StripeProductID = new(string)
				*ge.StripeProductID = value.String
			}
		case generation.FieldSourceType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field source_type", values[i])
			} else if value.Valid {
				ge.SourceType = enttypes.SourceType(value.String)
			}
		case generation.FieldPromptID:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field prompt_id", values[i])
			} else if value.Valid {
				ge.PromptID = new(uuid.UUID)
				*ge.PromptID = *value.S.(*uuid.UUID)
			}
		case generation.FieldNegativePromptID:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field negative_prompt_id", values[i])
			} else if value.Valid {
				ge.NegativePromptID = new(uuid.UUID)
				*ge.NegativePromptID = *value.S.(*uuid.UUID)
			}
		case generation.FieldModelID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field model_id", values[i])
			} else if value != nil {
				ge.ModelID = *value
			}
		case generation.FieldSchedulerID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field scheduler_id", values[i])
			} else if value != nil {
				ge.SchedulerID = *value
			}
		case generation.FieldUserID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field user_id", values[i])
			} else if value != nil {
				ge.UserID = *value
			}
		case generation.FieldDeviceInfoID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field device_info_id", values[i])
			} else if value != nil {
				ge.DeviceInfoID = *value
			}
		case generation.FieldAPITokenID:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field api_token_id", values[i])
			} else if value.Valid {
				ge.APITokenID = new(uuid.UUID)
				*ge.APITokenID = *value.S.(*uuid.UUID)
			}
		case generation.FieldStartedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field started_at", values[i])
			} else if value.Valid {
				ge.StartedAt = new(time.Time)
				*ge.StartedAt = value.Time
			}
		case generation.FieldCompletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field completed_at", values[i])
			} else if value.Valid {
				ge.CompletedAt = new(time.Time)
				*ge.CompletedAt = value.Time
			}
		case generation.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				ge.CreatedAt = value.Time
			}
		case generation.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				ge.UpdatedAt = value.Time
			}
		}
	}
	return nil
}

// QueryDeviceInfo queries the "device_info" edge of the Generation entity.
func (ge *Generation) QueryDeviceInfo() *DeviceInfoQuery {
	return NewGenerationClient(ge.config).QueryDeviceInfo(ge)
}

// QueryScheduler queries the "scheduler" edge of the Generation entity.
func (ge *Generation) QueryScheduler() *SchedulerQuery {
	return NewGenerationClient(ge.config).QueryScheduler(ge)
}

// QueryPrompt queries the "prompt" edge of the Generation entity.
func (ge *Generation) QueryPrompt() *PromptQuery {
	return NewGenerationClient(ge.config).QueryPrompt(ge)
}

// QueryNegativePrompt queries the "negative_prompt" edge of the Generation entity.
func (ge *Generation) QueryNegativePrompt() *NegativePromptQuery {
	return NewGenerationClient(ge.config).QueryNegativePrompt(ge)
}

// QueryGenerationModel queries the "generation_model" edge of the Generation entity.
func (ge *Generation) QueryGenerationModel() *GenerationModelQuery {
	return NewGenerationClient(ge.config).QueryGenerationModel(ge)
}

// QueryUser queries the "user" edge of the Generation entity.
func (ge *Generation) QueryUser() *UserQuery {
	return NewGenerationClient(ge.config).QueryUser(ge)
}

// QueryAPITokens queries the "api_tokens" edge of the Generation entity.
func (ge *Generation) QueryAPITokens() *ApiTokenQuery {
	return NewGenerationClient(ge.config).QueryAPITokens(ge)
}

// QueryGenerationOutputs queries the "generation_outputs" edge of the Generation entity.
func (ge *Generation) QueryGenerationOutputs() *GenerationOutputQuery {
	return NewGenerationClient(ge.config).QueryGenerationOutputs(ge)
}

// Update returns a builder for updating this Generation.
// Note that you need to call Generation.Unwrap() before calling this method if this Generation
// was returned from a transaction, and the transaction was committed or rolled back.
func (ge *Generation) Update() *GenerationUpdateOne {
	return NewGenerationClient(ge.config).UpdateOne(ge)
}

// Unwrap unwraps the Generation entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ge *Generation) Unwrap() *Generation {
	_tx, ok := ge.config.driver.(*txDriver)
	if !ok {
		panic("ent: Generation is not a transactional entity")
	}
	ge.config.driver = _tx.drv
	return ge
}

// String implements the fmt.Stringer.
func (ge *Generation) String() string {
	var builder strings.Builder
	builder.WriteString("Generation(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ge.ID))
	builder.WriteString("width=")
	builder.WriteString(fmt.Sprintf("%v", ge.Width))
	builder.WriteString(", ")
	builder.WriteString("height=")
	builder.WriteString(fmt.Sprintf("%v", ge.Height))
	builder.WriteString(", ")
	builder.WriteString("inference_steps=")
	builder.WriteString(fmt.Sprintf("%v", ge.InferenceSteps))
	builder.WriteString(", ")
	builder.WriteString("guidance_scale=")
	builder.WriteString(fmt.Sprintf("%v", ge.GuidanceScale))
	builder.WriteString(", ")
	builder.WriteString("num_outputs=")
	builder.WriteString(fmt.Sprintf("%v", ge.NumOutputs))
	builder.WriteString(", ")
	builder.WriteString("nsfw_count=")
	builder.WriteString(fmt.Sprintf("%v", ge.NsfwCount))
	builder.WriteString(", ")
	builder.WriteString("seed=")
	builder.WriteString(fmt.Sprintf("%v", ge.Seed))
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", ge.Status))
	builder.WriteString(", ")
	if v := ge.FailureReason; v != nil {
		builder.WriteString("failure_reason=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	if v := ge.CountryCode; v != nil {
		builder.WriteString("country_code=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	if v := ge.InitImageURL; v != nil {
		builder.WriteString("init_image_url=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	if v := ge.MaskImageURL; v != nil {
		builder.WriteString("mask_image_url=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	if v := ge.PromptStrength; v != nil {
		builder.WriteString("prompt_strength=")
		builder.WriteString(fmt.Sprintf("%v", *v))
	}
	builder.WriteString(", ")
	builder.WriteString("was_auto_submitted=")
	builder.WriteString(fmt.Sprintf("%v", ge.WasAutoSubmitted))
	builder.WriteString(", ")
	if v := ge.StripeProductID; v != nil {
		builder.WriteString("stripe_product_id=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	builder.WriteString("source_type=")
	builder.WriteString(fmt.Sprintf("%v", ge.SourceType))
	builder.WriteString(", ")
	if v := ge.PromptID; v != nil {
		builder.WriteString("prompt_id=")
		builder.WriteString(fmt.Sprintf("%v", *v))
	}
	builder.WriteString(", ")
	if v := ge.NegativePromptID; v != nil {
		builder.WriteString("negative_prompt_id=")
		builder.WriteString(fmt.Sprintf("%v", *v))
	}
	builder.WriteString(", ")
	builder.WriteString("model_id=")
	builder.WriteString(fmt.Sprintf("%v", ge.ModelID))
	builder.WriteString(", ")
	builder.WriteString("scheduler_id=")
	builder.WriteString(fmt.Sprintf("%v", ge.SchedulerID))
	builder.WriteString(", ")
	builder.WriteString("user_id=")
	builder.WriteString(fmt.Sprintf("%v", ge.UserID))
	builder.WriteString(", ")
	builder.WriteString("device_info_id=")
	builder.WriteString(fmt.Sprintf("%v", ge.DeviceInfoID))
	builder.WriteString(", ")
	if v := ge.APITokenID; v != nil {
		builder.WriteString("api_token_id=")
		builder.WriteString(fmt.Sprintf("%v", *v))
	}
	builder.WriteString(", ")
	if v := ge.StartedAt; v != nil {
		builder.WriteString("started_at=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteString(", ")
	if v := ge.CompletedAt; v != nil {
		builder.WriteString("completed_at=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(ge.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(ge.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// Generations is a parsable slice of Generation.
type Generations []*Generation

func (ge Generations) config(cfg config) {
	for _i := range ge {
		ge[_i].config = cfg
	}
}
