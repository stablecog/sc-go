// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/go-apps/database/ent/deviceinfo"
	"github.com/stablecog/go-apps/database/ent/generation"
	"github.com/stablecog/go-apps/database/ent/generationmodel"
	"github.com/stablecog/go-apps/database/ent/negativeprompt"
	"github.com/stablecog/go-apps/database/ent/prompt"
	"github.com/stablecog/go-apps/database/ent/scheduler"
	"github.com/stablecog/go-apps/database/ent/user"
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
	// NumInterferenceSteps holds the value of the "num_interference_steps" field.
	NumInterferenceSteps int32 `json:"num_interference_steps,omitempty"`
	// GuidanceScale holds the value of the "guidance_scale" field.
	GuidanceScale float32 `json:"guidance_scale,omitempty"`
	// Seed holds the value of the "seed" field.
	Seed int `json:"seed,omitempty"`
	// Status holds the value of the "status" field.
	Status generation.Status `json:"status,omitempty"`
	// FailureReason holds the value of the "failure_reason" field.
	FailureReason *string `json:"failure_reason,omitempty"`
	// CountryCode holds the value of the "country_code" field.
	CountryCode string `json:"country_code,omitempty"`
	// IsSubmittedToGallery holds the value of the "is_submitted_to_gallery" field.
	IsSubmittedToGallery bool `json:"is_submitted_to_gallery,omitempty"`
	// IsPublic holds the value of the "is_public" field.
	IsPublic bool `json:"is_public,omitempty"`
	// InitImageURL holds the value of the "init_image_url" field.
	InitImageURL *string `json:"init_image_url,omitempty"`
	// PromptID holds the value of the "prompt_id" field.
	PromptID uuid.UUID `json:"prompt_id,omitempty"`
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
	// Schedulers holds the value of the schedulers edge.
	Schedulers *Scheduler `json:"schedulers,omitempty"`
	// Prompts holds the value of the prompts edge.
	Prompts *Prompt `json:"prompts,omitempty"`
	// NegativePrompts holds the value of the negative_prompts edge.
	NegativePrompts *NegativePrompt `json:"negative_prompts,omitempty"`
	// GenerationModels holds the value of the generation_models edge.
	GenerationModels *GenerationModel `json:"generation_models,omitempty"`
	// Users holds the value of the users edge.
	Users *User `json:"users,omitempty"`
	// GenerationOutputs holds the value of the generation_outputs edge.
	GenerationOutputs []*GenerationOutput `json:"generation_outputs,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [7]bool
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

// SchedulersOrErr returns the Schedulers value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationEdges) SchedulersOrErr() (*Scheduler, error) {
	if e.loadedTypes[1] {
		if e.Schedulers == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: scheduler.Label}
		}
		return e.Schedulers, nil
	}
	return nil, &NotLoadedError{edge: "schedulers"}
}

// PromptsOrErr returns the Prompts value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationEdges) PromptsOrErr() (*Prompt, error) {
	if e.loadedTypes[2] {
		if e.Prompts == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: prompt.Label}
		}
		return e.Prompts, nil
	}
	return nil, &NotLoadedError{edge: "prompts"}
}

// NegativePromptsOrErr returns the NegativePrompts value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationEdges) NegativePromptsOrErr() (*NegativePrompt, error) {
	if e.loadedTypes[3] {
		if e.NegativePrompts == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: negativeprompt.Label}
		}
		return e.NegativePrompts, nil
	}
	return nil, &NotLoadedError{edge: "negative_prompts"}
}

// GenerationModelsOrErr returns the GenerationModels value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationEdges) GenerationModelsOrErr() (*GenerationModel, error) {
	if e.loadedTypes[4] {
		if e.GenerationModels == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: generationmodel.Label}
		}
		return e.GenerationModels, nil
	}
	return nil, &NotLoadedError{edge: "generation_models"}
}

// UsersOrErr returns the Users value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationEdges) UsersOrErr() (*User, error) {
	if e.loadedTypes[5] {
		if e.Users == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: user.Label}
		}
		return e.Users, nil
	}
	return nil, &NotLoadedError{edge: "users"}
}

// GenerationOutputsOrErr returns the GenerationOutputs value or an error if the edge
// was not loaded in eager-loading.
func (e GenerationEdges) GenerationOutputsOrErr() ([]*GenerationOutput, error) {
	if e.loadedTypes[6] {
		return e.GenerationOutputs, nil
	}
	return nil, &NotLoadedError{edge: "generation_outputs"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Generation) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case generation.FieldNegativePromptID:
			values[i] = &sql.NullScanner{S: new(uuid.UUID)}
		case generation.FieldIsSubmittedToGallery, generation.FieldIsPublic:
			values[i] = new(sql.NullBool)
		case generation.FieldGuidanceScale:
			values[i] = new(sql.NullFloat64)
		case generation.FieldWidth, generation.FieldHeight, generation.FieldNumInterferenceSteps, generation.FieldSeed:
			values[i] = new(sql.NullInt64)
		case generation.FieldStatus, generation.FieldFailureReason, generation.FieldCountryCode, generation.FieldInitImageURL:
			values[i] = new(sql.NullString)
		case generation.FieldStartedAt, generation.FieldCompletedAt, generation.FieldCreatedAt, generation.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case generation.FieldID, generation.FieldPromptID, generation.FieldModelID, generation.FieldSchedulerID, generation.FieldUserID, generation.FieldDeviceInfoID:
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
		case generation.FieldNumInterferenceSteps:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field num_interference_steps", values[i])
			} else if value.Valid {
				ge.NumInterferenceSteps = int32(value.Int64)
			}
		case generation.FieldGuidanceScale:
			if value, ok := values[i].(*sql.NullFloat64); !ok {
				return fmt.Errorf("unexpected type %T for field guidance_scale", values[i])
			} else if value.Valid {
				ge.GuidanceScale = float32(value.Float64)
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
				ge.CountryCode = value.String
			}
		case generation.FieldIsSubmittedToGallery:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field is_submitted_to_gallery", values[i])
			} else if value.Valid {
				ge.IsSubmittedToGallery = value.Bool
			}
		case generation.FieldIsPublic:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field is_public", values[i])
			} else if value.Valid {
				ge.IsPublic = value.Bool
			}
		case generation.FieldInitImageURL:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field init_image_url", values[i])
			} else if value.Valid {
				ge.InitImageURL = new(string)
				*ge.InitImageURL = value.String
			}
		case generation.FieldPromptID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field prompt_id", values[i])
			} else if value != nil {
				ge.PromptID = *value
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

// QuerySchedulers queries the "schedulers" edge of the Generation entity.
func (ge *Generation) QuerySchedulers() *SchedulerQuery {
	return NewGenerationClient(ge.config).QuerySchedulers(ge)
}

// QueryPrompts queries the "prompts" edge of the Generation entity.
func (ge *Generation) QueryPrompts() *PromptQuery {
	return NewGenerationClient(ge.config).QueryPrompts(ge)
}

// QueryNegativePrompts queries the "negative_prompts" edge of the Generation entity.
func (ge *Generation) QueryNegativePrompts() *NegativePromptQuery {
	return NewGenerationClient(ge.config).QueryNegativePrompts(ge)
}

// QueryGenerationModels queries the "generation_models" edge of the Generation entity.
func (ge *Generation) QueryGenerationModels() *GenerationModelQuery {
	return NewGenerationClient(ge.config).QueryGenerationModels(ge)
}

// QueryUsers queries the "users" edge of the Generation entity.
func (ge *Generation) QueryUsers() *UserQuery {
	return NewGenerationClient(ge.config).QueryUsers(ge)
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
	builder.WriteString("num_interference_steps=")
	builder.WriteString(fmt.Sprintf("%v", ge.NumInterferenceSteps))
	builder.WriteString(", ")
	builder.WriteString("guidance_scale=")
	builder.WriteString(fmt.Sprintf("%v", ge.GuidanceScale))
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
	builder.WriteString("country_code=")
	builder.WriteString(ge.CountryCode)
	builder.WriteString(", ")
	builder.WriteString("is_submitted_to_gallery=")
	builder.WriteString(fmt.Sprintf("%v", ge.IsSubmittedToGallery))
	builder.WriteString(", ")
	builder.WriteString("is_public=")
	builder.WriteString(fmt.Sprintf("%v", ge.IsPublic))
	builder.WriteString(", ")
	if v := ge.InitImageURL; v != nil {
		builder.WriteString("init_image_url=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	builder.WriteString("prompt_id=")
	builder.WriteString(fmt.Sprintf("%v", ge.PromptID))
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
