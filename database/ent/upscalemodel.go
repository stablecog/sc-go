// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/upscalemodel"
)

// UpscaleModel is the model entity for the UpscaleModel schema.
type UpscaleModel struct {
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
	// RunpodEndpoint holds the value of the "runpod_endpoint" field.
	RunpodEndpoint *string `json:"runpod_endpoint,omitempty"`
	// RunpodActive holds the value of the "runpod_active" field.
	RunpodActive bool `json:"runpod_active,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the UpscaleModelQuery when eager-loading is set.
	Edges        UpscaleModelEdges `json:"edges"`
	selectValues sql.SelectValues
}

// UpscaleModelEdges holds the relations/edges for other nodes in the graph.
type UpscaleModelEdges struct {
	// Upscales holds the value of the upscales edge.
	Upscales []*Upscale `json:"upscales,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [1]bool
}

// UpscalesOrErr returns the Upscales value or an error if the edge
// was not loaded in eager-loading.
func (e UpscaleModelEdges) UpscalesOrErr() ([]*Upscale, error) {
	if e.loadedTypes[0] {
		return e.Upscales, nil
	}
	return nil, &NotLoadedError{edge: "upscales"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*UpscaleModel) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case upscalemodel.FieldIsActive, upscalemodel.FieldIsDefault, upscalemodel.FieldIsHidden, upscalemodel.FieldRunpodActive:
			values[i] = new(sql.NullBool)
		case upscalemodel.FieldNameInWorker, upscalemodel.FieldRunpodEndpoint:
			values[i] = new(sql.NullString)
		case upscalemodel.FieldCreatedAt, upscalemodel.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case upscalemodel.FieldID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the UpscaleModel fields.
func (um *UpscaleModel) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case upscalemodel.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				um.ID = *value
			}
		case upscalemodel.FieldNameInWorker:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field name_in_worker", values[i])
			} else if value.Valid {
				um.NameInWorker = value.String
			}
		case upscalemodel.FieldIsActive:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field is_active", values[i])
			} else if value.Valid {
				um.IsActive = value.Bool
			}
		case upscalemodel.FieldIsDefault:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field is_default", values[i])
			} else if value.Valid {
				um.IsDefault = value.Bool
			}
		case upscalemodel.FieldIsHidden:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field is_hidden", values[i])
			} else if value.Valid {
				um.IsHidden = value.Bool
			}
		case upscalemodel.FieldRunpodEndpoint:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field runpod_endpoint", values[i])
			} else if value.Valid {
				um.RunpodEndpoint = new(string)
				*um.RunpodEndpoint = value.String
			}
		case upscalemodel.FieldRunpodActive:
			if value, ok := values[i].(*sql.NullBool); !ok {
				return fmt.Errorf("unexpected type %T for field runpod_active", values[i])
			} else if value.Valid {
				um.RunpodActive = value.Bool
			}
		case upscalemodel.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				um.CreatedAt = value.Time
			}
		case upscalemodel.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				um.UpdatedAt = value.Time
			}
		default:
			um.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the UpscaleModel.
// This includes values selected through modifiers, order, etc.
func (um *UpscaleModel) Value(name string) (ent.Value, error) {
	return um.selectValues.Get(name)
}

// QueryUpscales queries the "upscales" edge of the UpscaleModel entity.
func (um *UpscaleModel) QueryUpscales() *UpscaleQuery {
	return NewUpscaleModelClient(um.config).QueryUpscales(um)
}

// Update returns a builder for updating this UpscaleModel.
// Note that you need to call UpscaleModel.Unwrap() before calling this method if this UpscaleModel
// was returned from a transaction, and the transaction was committed or rolled back.
func (um *UpscaleModel) Update() *UpscaleModelUpdateOne {
	return NewUpscaleModelClient(um.config).UpdateOne(um)
}

// Unwrap unwraps the UpscaleModel entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (um *UpscaleModel) Unwrap() *UpscaleModel {
	_tx, ok := um.config.driver.(*txDriver)
	if !ok {
		panic("ent: UpscaleModel is not a transactional entity")
	}
	um.config.driver = _tx.drv
	return um
}

// String implements the fmt.Stringer.
func (um *UpscaleModel) String() string {
	var builder strings.Builder
	builder.WriteString("UpscaleModel(")
	builder.WriteString(fmt.Sprintf("id=%v, ", um.ID))
	builder.WriteString("name_in_worker=")
	builder.WriteString(um.NameInWorker)
	builder.WriteString(", ")
	builder.WriteString("is_active=")
	builder.WriteString(fmt.Sprintf("%v", um.IsActive))
	builder.WriteString(", ")
	builder.WriteString("is_default=")
	builder.WriteString(fmt.Sprintf("%v", um.IsDefault))
	builder.WriteString(", ")
	builder.WriteString("is_hidden=")
	builder.WriteString(fmt.Sprintf("%v", um.IsHidden))
	builder.WriteString(", ")
	if v := um.RunpodEndpoint; v != nil {
		builder.WriteString("runpod_endpoint=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	builder.WriteString("runpod_active=")
	builder.WriteString(fmt.Sprintf("%v", um.RunpodActive))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(um.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(um.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// UpscaleModels is a parsable slice of UpscaleModel.
type UpscaleModels []*UpscaleModel
