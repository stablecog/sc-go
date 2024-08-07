// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/apitoken"
	"github.com/stablecog/sc-go/database/ent/deviceinfo"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/database/ent/user"
	"github.com/stablecog/sc-go/database/ent/voiceover"
	"github.com/stablecog/sc-go/database/ent/voiceovermodel"
	"github.com/stablecog/sc-go/database/ent/voiceoverspeaker"
	"github.com/stablecog/sc-go/database/enttypes"
)

// Voiceover is the model entity for the Voiceover schema.
type Voiceover struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// CountryCode holds the value of the "country_code" field.
	CountryCode *string `json:"country_code,omitempty"`
	// Status holds the value of the "status" field.
	Status voiceover.Status `json:"status,omitempty"`
	// FailureReason holds the value of the "failure_reason" field.
	FailureReason *string `json:"failure_reason,omitempty"`
	// StripeProductID holds the value of the "stripe_product_id" field.
	StripeProductID *string `json:"stripe_product_id,omitempty"`
	// Temperature holds the value of the "temperature" field.
	Temperature float32 `json:"temperature,omitempty"`
	// Seed holds the value of the "seed" field.
	Seed int `json:"seed,omitempty"`
	// WasAutoSubmitted holds the value of the "was_auto_submitted" field.
	WasAutoSubmitted bool `json:"was_auto_submitted,omitempty"`
	// DenoiseAudio holds the value of the "denoise_audio" field.
	DenoiseAudio bool `json:"denoise_audio,omitempty"`
	// RemoveSilence holds the value of the "remove_silence" field.
	RemoveSilence bool `json:"remove_silence,omitempty"`
	// Cost holds the value of the "cost" field.
	Cost int32 `json:"cost,omitempty"`
	// SourceType holds the value of the "source_type" field.
	SourceType enttypes.SourceType `json:"source_type,omitempty"`
	// PromptID holds the value of the "prompt_id" field.
	PromptID *uuid.UUID `json:"prompt_id,omitempty"`
	// UserID holds the value of the "user_id" field.
	UserID uuid.UUID `json:"user_id,omitempty"`
	// DeviceInfoID holds the value of the "device_info_id" field.
	DeviceInfoID uuid.UUID `json:"device_info_id,omitempty"`
	// ModelID holds the value of the "model_id" field.
	ModelID uuid.UUID `json:"model_id,omitempty"`
	// SpeakerID holds the value of the "speaker_id" field.
	SpeakerID uuid.UUID `json:"speaker_id,omitempty"`
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
	// The values are being populated by the VoiceoverQuery when eager-loading is set.
	Edges        VoiceoverEdges `json:"edges"`
	selectValues sql.SelectValues
}

// VoiceoverEdges holds the relations/edges for other nodes in the graph.
type VoiceoverEdges struct {
	// User holds the value of the user edge.
	User *User `json:"user,omitempty"`
	// Prompt holds the value of the prompt edge.
	Prompt *Prompt `json:"prompt,omitempty"`
	// DeviceInfo holds the value of the device_info edge.
	DeviceInfo *DeviceInfo `json:"device_info,omitempty"`
	// VoiceoverModels holds the value of the voiceover_models edge.
	VoiceoverModels *VoiceoverModel `json:"voiceover_models,omitempty"`
	// VoiceoverSpeakers holds the value of the voiceover_speakers edge.
	VoiceoverSpeakers *VoiceoverSpeaker `json:"voiceover_speakers,omitempty"`
	// APITokens holds the value of the api_tokens edge.
	APITokens *ApiToken `json:"api_tokens,omitempty"`
	// VoiceoverOutputs holds the value of the voiceover_outputs edge.
	VoiceoverOutputs []*VoiceoverOutput `json:"voiceover_outputs,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [7]bool
}

// UserOrErr returns the User value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e VoiceoverEdges) UserOrErr() (*User, error) {
	if e.User != nil {
		return e.User, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "user"}
}

// PromptOrErr returns the Prompt value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e VoiceoverEdges) PromptOrErr() (*Prompt, error) {
	if e.Prompt != nil {
		return e.Prompt, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: prompt.Label}
	}
	return nil, &NotLoadedError{edge: "prompt"}
}

// DeviceInfoOrErr returns the DeviceInfo value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e VoiceoverEdges) DeviceInfoOrErr() (*DeviceInfo, error) {
	if e.DeviceInfo != nil {
		return e.DeviceInfo, nil
	} else if e.loadedTypes[2] {
		return nil, &NotFoundError{label: deviceinfo.Label}
	}
	return nil, &NotLoadedError{edge: "device_info"}
}

// VoiceoverModelsOrErr returns the VoiceoverModels value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e VoiceoverEdges) VoiceoverModelsOrErr() (*VoiceoverModel, error) {
	if e.VoiceoverModels != nil {
		return e.VoiceoverModels, nil
	} else if e.loadedTypes[3] {
		return nil, &NotFoundError{label: voiceovermodel.Label}
	}
	return nil, &NotLoadedError{edge: "voiceover_models"}
}

// VoiceoverSpeakersOrErr returns the VoiceoverSpeakers value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e VoiceoverEdges) VoiceoverSpeakersOrErr() (*VoiceoverSpeaker, error) {
	if e.VoiceoverSpeakers != nil {
		return e.VoiceoverSpeakers, nil
	} else if e.loadedTypes[4] {
		return nil, &NotFoundError{label: voiceoverspeaker.Label}
	}
	return nil, &NotLoadedError{edge: "voiceover_speakers"}
}

// APITokensOrErr returns the APITokens value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e VoiceoverEdges) APITokensOrErr() (*ApiToken, error) {
	if e.APITokens != nil {
		return e.APITokens, nil
	} else if e.loadedTypes[5] {
		return nil, &NotFoundError{label: apitoken.Label}
	}
	return nil, &NotLoadedError{edge: "api_tokens"}
}

// VoiceoverOutputsOrErr returns the VoiceoverOutputs value or an error if the edge
// was not loaded in eager-loading.
func (e VoiceoverEdges) VoiceoverOutputsOrErr() ([]*VoiceoverOutput, error) {
	if e.loadedTypes[6] {
		return e.VoiceoverOutputs, nil
	}
	return nil, &NotLoadedError{edge: "voiceover_outputs"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*Voiceover) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case voiceover.FieldPromptID, voiceover.FieldAPITokenID:
			values[i] = &sql.NullScanner{S: new(uuid.UUID)}
		case voiceover.FieldWasAutoSubmitted, voiceover.FieldDenoiseAudio, voiceover.FieldRemoveSilence:
			values[i] = new(sql.NullBool)
		case voiceover.FieldTemperature:
			values[i] = new(sql.NullFloat64)
		case voiceover.FieldSeed, voiceover.FieldCost:
			values[i] = new(sql.NullInt64)
		case voiceover.FieldCountryCode, voiceover.FieldStatus, voiceover.FieldFailureReason, voiceover.FieldStripeProductID, voiceover.FieldSourceType:
			values[i] = new(sql.NullString)
		case voiceover.FieldStartedAt, voiceover.FieldCompletedAt, voiceover.FieldCreatedAt, voiceover.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case voiceover.FieldID, voiceover.FieldUserID, voiceover.FieldDeviceInfoID, voiceover.FieldModelID, voiceover.FieldSpeakerID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the Voiceover fields.
func (v *Voiceover) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case voiceover.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				v.ID = *value
			}
		case voiceover.FieldCountryCode:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field country_code", values[i])
			} else if value.Valid {
				v.CountryCode = new(string)
				*v.CountryCode = value.String
			}
		case voiceover.FieldStatus:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field status", values[i])
			} else if value.Valid {
				v.Status = voiceover.Status(value.String)
			}
		case voiceover.FieldFailureReason:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field failure_reason", values[i])
			} else if value.Valid {
				v.FailureReason = new(string)
				*v.FailureReason = value.String
			}
		case voiceover.FieldStripeProductID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field stripe_product_id", values[i])
			} else if value.Valid {
				v.StripeProductID = new(string)
				*v.StripeProductID = value.String
			}
		case voiceover.FieldTemperature:
			if value, ok := values[i].(*sql.NullFloat64); !ok {
				return fmt.Errorf("unexpected type %T for field temperature", values[i])
			} else if value.Valid {
				v.Temperature = float32(value.Float64)
			}
		case voiceover.FieldSeed:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field seed", values[i])
			} else if value.Valid {
				v.Seed = int(value.Int64)
			}
		case voiceover.FieldWasAutoSubmitted:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field was_auto_submitted", values[i])
			} else if value.Valid {
				v.WasAutoSubmitted = value.Bool
			}
		case voiceover.FieldDenoiseAudio:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field denoise_audio", values[i])
			} else if value.Valid {
				v.DenoiseAudio = value.Bool
			}
		case voiceover.FieldRemoveSilence:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field remove_silence", values[i])
			} else if value.Valid {
				v.RemoveSilence = value.Bool
			}
		case voiceover.FieldCost:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field cost", values[i])
			} else if value.Valid {
				v.Cost = int32(value.Int64)
			}
		case voiceover.FieldSourceType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field source_type", values[i])
			} else if value.Valid {
				v.SourceType = enttypes.SourceType(value.String)
			}
		case voiceover.FieldPromptID:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field prompt_id", values[i])
			} else if value.Valid {
				v.PromptID = new(uuid.UUID)
				*v.PromptID = *value.S.(*uuid.UUID)
			}
		case voiceover.FieldUserID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field user_id", values[i])
			} else if value != nil {
				v.UserID = *value
			}
		case voiceover.FieldDeviceInfoID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field device_info_id", values[i])
			} else if value != nil {
				v.DeviceInfoID = *value
			}
		case voiceover.FieldModelID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field model_id", values[i])
			} else if value != nil {
				v.ModelID = *value
			}
		case voiceover.FieldSpeakerID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field speaker_id", values[i])
			} else if value != nil {
				v.SpeakerID = *value
			}
		case voiceover.FieldAPITokenID:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field api_token_id", values[i])
			} else if value.Valid {
				v.APITokenID = new(uuid.UUID)
				*v.APITokenID = *value.S.(*uuid.UUID)
			}
		case voiceover.FieldStartedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field started_at", values[i])
			} else if value.Valid {
				v.StartedAt = new(time.Time)
				*v.StartedAt = value.Time
			}
		case voiceover.FieldCompletedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field completed_at", values[i])
			} else if value.Valid {
				v.CompletedAt = new(time.Time)
				*v.CompletedAt = value.Time
			}
		case voiceover.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				v.CreatedAt = value.Time
			}
		case voiceover.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				v.UpdatedAt = value.Time
			}
		default:
			v.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the Voiceover.
// This includes values selected through modifiers, order, etc.
func (v *Voiceover) Value(name string) (ent.Value, error) {
	return v.selectValues.Get(name)
}

// QueryUser queries the "user" edge of the Voiceover entity.
func (v *Voiceover) QueryUser() *UserQuery {
	return NewVoiceoverClient(v.config).QueryUser(v)
}

// QueryPrompt queries the "prompt" edge of the Voiceover entity.
func (v *Voiceover) QueryPrompt() *PromptQuery {
	return NewVoiceoverClient(v.config).QueryPrompt(v)
}

// QueryDeviceInfo queries the "device_info" edge of the Voiceover entity.
func (v *Voiceover) QueryDeviceInfo() *DeviceInfoQuery {
	return NewVoiceoverClient(v.config).QueryDeviceInfo(v)
}

// QueryVoiceoverModels queries the "voiceover_models" edge of the Voiceover entity.
func (v *Voiceover) QueryVoiceoverModels() *VoiceoverModelQuery {
	return NewVoiceoverClient(v.config).QueryVoiceoverModels(v)
}

// QueryVoiceoverSpeakers queries the "voiceover_speakers" edge of the Voiceover entity.
func (v *Voiceover) QueryVoiceoverSpeakers() *VoiceoverSpeakerQuery {
	return NewVoiceoverClient(v.config).QueryVoiceoverSpeakers(v)
}

// QueryAPITokens queries the "api_tokens" edge of the Voiceover entity.
func (v *Voiceover) QueryAPITokens() *ApiTokenQuery {
	return NewVoiceoverClient(v.config).QueryAPITokens(v)
}

// QueryVoiceoverOutputs queries the "voiceover_outputs" edge of the Voiceover entity.
func (v *Voiceover) QueryVoiceoverOutputs() *VoiceoverOutputQuery {
	return NewVoiceoverClient(v.config).QueryVoiceoverOutputs(v)
}

// Update returns a builder for updating this Voiceover.
// Note that you need to call Voiceover.Unwrap() before calling this method if this Voiceover
// was returned from a transaction, and the transaction was committed or rolled back.
func (v *Voiceover) Update() *VoiceoverUpdateOne {
	return NewVoiceoverClient(v.config).UpdateOne(v)
}

// Unwrap unwraps the Voiceover entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (v *Voiceover) Unwrap() *Voiceover {
	_tx, ok := v.config.driver.(*txDriver)
	if !ok {
		panic("ent: Voiceover is not a transactional entity")
	}
	v.config.driver = _tx.drv
	return v
}

// String implements the fmt.Stringer.
func (v *Voiceover) String() string {
	var builder strings.Builder
	builder.WriteString("Voiceover(")
	builder.WriteString(fmt.Sprintf("id=%v, ", v.ID))
	if v := v.CountryCode; v != nil {
		builder.WriteString("country_code=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	builder.WriteString("status=")
	builder.WriteString(fmt.Sprintf("%v", v.Status))
	builder.WriteString(", ")
	if v := v.FailureReason; v != nil {
		builder.WriteString("failure_reason=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	if v := v.StripeProductID; v != nil {
		builder.WriteString("stripe_product_id=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	builder.WriteString("temperature=")
	builder.WriteString(fmt.Sprintf("%v", v.Temperature))
	builder.WriteString(", ")
	builder.WriteString("seed=")
	builder.WriteString(fmt.Sprintf("%v", v.Seed))
	builder.WriteString(", ")
	builder.WriteString("was_auto_submitted=")
	builder.WriteString(fmt.Sprintf("%v", v.WasAutoSubmitted))
	builder.WriteString(", ")
	builder.WriteString("denoise_audio=")
	builder.WriteString(fmt.Sprintf("%v", v.DenoiseAudio))
	builder.WriteString(", ")
	builder.WriteString("remove_silence=")
	builder.WriteString(fmt.Sprintf("%v", v.RemoveSilence))
	builder.WriteString(", ")
	builder.WriteString("cost=")
	builder.WriteString(fmt.Sprintf("%v", v.Cost))
	builder.WriteString(", ")
	builder.WriteString("source_type=")
	builder.WriteString(fmt.Sprintf("%v", v.SourceType))
	builder.WriteString(", ")
	if v := v.PromptID; v != nil {
		builder.WriteString("prompt_id=")
		builder.WriteString(fmt.Sprintf("%v", *v))
	}
	builder.WriteString(", ")
	builder.WriteString("user_id=")
	builder.WriteString(fmt.Sprintf("%v", v.UserID))
	builder.WriteString(", ")
	builder.WriteString("device_info_id=")
	builder.WriteString(fmt.Sprintf("%v", v.DeviceInfoID))
	builder.WriteString(", ")
	builder.WriteString("model_id=")
	builder.WriteString(fmt.Sprintf("%v", v.ModelID))
	builder.WriteString(", ")
	builder.WriteString("speaker_id=")
	builder.WriteString(fmt.Sprintf("%v", v.SpeakerID))
	builder.WriteString(", ")
	if v := v.APITokenID; v != nil {
		builder.WriteString("api_token_id=")
		builder.WriteString(fmt.Sprintf("%v", *v))
	}
	builder.WriteString(", ")
	if v := v.StartedAt; v != nil {
		builder.WriteString("started_at=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteString(", ")
	if v := v.CompletedAt; v != nil {
		builder.WriteString("completed_at=")
		builder.WriteString(v.Format(time.ANSIC))
	}
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(v.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(v.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// Voiceovers is a parsable slice of Voiceover.
type Voiceovers []*Voiceover
