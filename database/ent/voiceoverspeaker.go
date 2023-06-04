// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/voiceovermodel"
	"github.com/stablecog/sc-go/database/ent/voiceoverspeaker"
)

// VoiceoverSpeaker is the model entity for the VoiceoverSpeaker schema.
type VoiceoverSpeaker struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// NameInWorker holds the value of the "name_in_worker" field.
	NameInWorker string `json:"name_in_worker,omitempty"`
	// IsActive holds the value of the "is_active" field.
	IsActive bool `json:"is_active,omitempty"`
	// IsDefault holds the value of the "is_default" field.
	IsDefault bool `json:"is_default,omitempty"`
	// IsHidden holds the value of the "is_hidden" field.
	IsHidden bool `json:"is_hidden,omitempty"`
	// ModelID holds the value of the "model_id" field.
	ModelID uuid.UUID `json:"model_id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the VoiceoverSpeakerQuery when eager-loading is set.
	Edges VoiceoverSpeakerEdges `json:"edges"`
}

// VoiceoverSpeakerEdges holds the relations/edges for other nodes in the graph.
type VoiceoverSpeakerEdges struct {
	// Voiceovers holds the value of the voiceovers edge.
	Voiceovers []*Voiceover `json:"voiceovers,omitempty"`
	// VoiceoverModels holds the value of the voiceover_models edge.
	VoiceoverModels *VoiceoverModel `json:"voiceover_models,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// VoiceoversOrErr returns the Voiceovers value or an error if the edge
// was not loaded in eager-loading.
func (e VoiceoverSpeakerEdges) VoiceoversOrErr() ([]*Voiceover, error) {
	if e.loadedTypes[0] {
		return e.Voiceovers, nil
	}
	return nil, &NotLoadedError{edge: "voiceovers"}
}

// VoiceoverModelsOrErr returns the VoiceoverModels value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e VoiceoverSpeakerEdges) VoiceoverModelsOrErr() (*VoiceoverModel, error) {
	if e.loadedTypes[1] {
		if e.VoiceoverModels == nil {
			// Edge was loaded but was not found.
			return nil, &NotFoundError{label: voiceovermodel.Label}
		}
		return e.VoiceoverModels, nil
	}
	return nil, &NotLoadedError{edge: "voiceover_models"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*VoiceoverSpeaker) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case voiceoverspeaker.FieldIsActive, voiceoverspeaker.FieldIsDefault, voiceoverspeaker.FieldIsHidden:
			values[i] = new(sql.NullBool)
		case voiceoverspeaker.FieldNameInWorker:
			values[i] = new(sql.NullString)
		case voiceoverspeaker.FieldCreatedAt, voiceoverspeaker.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case voiceoverspeaker.FieldID, voiceoverspeaker.FieldModelID:
			values[i] = new(uuid.UUID)
		default:
			return nil, fmt.Errorf("unexpected column %q for type VoiceoverSpeaker", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the VoiceoverSpeaker fields.
func (vs *VoiceoverSpeaker) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case voiceoverspeaker.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				vs.ID = *value
			}
		case voiceoverspeaker.FieldNameInWorker:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name_in_worker", values[i])
			} else if value.Valid {
				vs.NameInWorker = value.String
			}
		case voiceoverspeaker.FieldIsActive:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field is_active", values[i])
			} else if value.Valid {
				vs.IsActive = value.Bool
			}
		case voiceoverspeaker.FieldIsDefault:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field is_default", values[i])
			} else if value.Valid {
				vs.IsDefault = value.Bool
			}
		case voiceoverspeaker.FieldIsHidden:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field is_hidden", values[i])
			} else if value.Valid {
				vs.IsHidden = value.Bool
			}
		case voiceoverspeaker.FieldModelID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field model_id", values[i])
			} else if value != nil {
				vs.ModelID = *value
			}
		case voiceoverspeaker.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				vs.CreatedAt = value.Time
			}
		case voiceoverspeaker.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				vs.UpdatedAt = value.Time
			}
		}
	}
	return nil
}

// QueryVoiceovers queries the "voiceovers" edge of the VoiceoverSpeaker entity.
func (vs *VoiceoverSpeaker) QueryVoiceovers() *VoiceoverQuery {
	return NewVoiceoverSpeakerClient(vs.config).QueryVoiceovers(vs)
}

// QueryVoiceoverModels queries the "voiceover_models" edge of the VoiceoverSpeaker entity.
func (vs *VoiceoverSpeaker) QueryVoiceoverModels() *VoiceoverModelQuery {
	return NewVoiceoverSpeakerClient(vs.config).QueryVoiceoverModels(vs)
}

// Update returns a builder for updating this VoiceoverSpeaker.
// Note that you need to call VoiceoverSpeaker.Unwrap() before calling this method if this VoiceoverSpeaker
// was returned from a transaction, and the transaction was committed or rolled back.
func (vs *VoiceoverSpeaker) Update() *VoiceoverSpeakerUpdateOne {
	return NewVoiceoverSpeakerClient(vs.config).UpdateOne(vs)
}

// Unwrap unwraps the VoiceoverSpeaker entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (vs *VoiceoverSpeaker) Unwrap() *VoiceoverSpeaker {
	_tx, ok := vs.config.driver.(*txDriver)
	if !ok {
		panic("ent: VoiceoverSpeaker is not a transactional entity")
	}
	vs.config.driver = _tx.drv
	return vs
}

// String implements the fmt.Stringer.
func (vs *VoiceoverSpeaker) String() string {
	var builder strings.Builder
	builder.WriteString("VoiceoverSpeaker(")
	builder.WriteString(fmt.Sprintf("id=%v, ", vs.ID))
	builder.WriteString("name_in_worker=")
	builder.WriteString(vs.NameInWorker)
	builder.WriteString(", ")
	builder.WriteString("is_active=")
	builder.WriteString(fmt.Sprintf("%v", vs.IsActive))
	builder.WriteString(", ")
	builder.WriteString("is_default=")
	builder.WriteString(fmt.Sprintf("%v", vs.IsDefault))
	builder.WriteString(", ")
	builder.WriteString("is_hidden=")
	builder.WriteString(fmt.Sprintf("%v", vs.IsHidden))
	builder.WriteString(", ")
	builder.WriteString("model_id=")
	builder.WriteString(fmt.Sprintf("%v", vs.ModelID))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(vs.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(vs.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// VoiceoverSpeakers is a parsable slice of VoiceoverSpeaker.
type VoiceoverSpeakers []*VoiceoverSpeaker

func (vs VoiceoverSpeakers) config(cfg config) {
	for _i := range vs {
		vs[_i].config = cfg
	}
}