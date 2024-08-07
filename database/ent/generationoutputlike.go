// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/ent/generationoutputlike"
	"github.com/stablecog/sc-go/database/ent/user"
)

// GenerationOutputLike is the model entity for the GenerationOutputLike schema.
type GenerationOutputLike struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// OutputID holds the value of the "output_id" field.
	OutputID uuid.UUID `json:"output_id,omitempty"`
	// LikedByUserID holds the value of the "liked_by_user_id" field.
	LikedByUserID uuid.UUID `json:"liked_by_user_id,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the GenerationOutputLikeQuery when eager-loading is set.
	Edges        GenerationOutputLikeEdges `json:"edges"`
	selectValues sql.SelectValues
}

// GenerationOutputLikeEdges holds the relations/edges for other nodes in the graph.
type GenerationOutputLikeEdges struct {
	// GenerationOutputs holds the value of the generation_outputs edge.
	GenerationOutputs *GenerationOutput `json:"generation_outputs,omitempty"`
	// Users holds the value of the users edge.
	Users *User `json:"users,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// GenerationOutputsOrErr returns the GenerationOutputs value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationOutputLikeEdges) GenerationOutputsOrErr() (*GenerationOutput, error) {
	if e.GenerationOutputs != nil {
		return e.GenerationOutputs, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: generationoutput.Label}
	}
	return nil, &NotLoadedError{edge: "generation_outputs"}
}

// UsersOrErr returns the Users value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e GenerationOutputLikeEdges) UsersOrErr() (*User, error) {
	if e.Users != nil {
		return e.Users, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "users"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*GenerationOutputLike) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case generationoutputlike.FieldCreatedAt:
			values[i] = new(sql.NullTime)
		case generationoutputlike.FieldID, generationoutputlike.FieldOutputID, generationoutputlike.FieldLikedByUserID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the GenerationOutputLike fields.
func (gol *GenerationOutputLike) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case generationoutputlike.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				gol.ID = *value
			}
		case generationoutputlike.FieldOutputID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field output_id", values[i])
			} else if value != nil {
				gol.OutputID = *value
			}
		case generationoutputlike.FieldLikedByUserID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field liked_by_user_id", values[i])
			} else if value != nil {
				gol.LikedByUserID = *value
			}
		case generationoutputlike.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				gol.CreatedAt = value.Time
			}
		default:
			gol.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the GenerationOutputLike.
// This includes values selected through modifiers, order, etc.
func (gol *GenerationOutputLike) Value(name string) (ent.Value, error) {
	return gol.selectValues.Get(name)
}

// QueryGenerationOutputs queries the "generation_outputs" edge of the GenerationOutputLike entity.
func (gol *GenerationOutputLike) QueryGenerationOutputs() *GenerationOutputQuery {
	return NewGenerationOutputLikeClient(gol.config).QueryGenerationOutputs(gol)
}

// QueryUsers queries the "users" edge of the GenerationOutputLike entity.
func (gol *GenerationOutputLike) QueryUsers() *UserQuery {
	return NewGenerationOutputLikeClient(gol.config).QueryUsers(gol)
}

// Update returns a builder for updating this GenerationOutputLike.
// Note that you need to call GenerationOutputLike.Unwrap() before calling this method if this GenerationOutputLike
// was returned from a transaction, and the transaction was committed or rolled back.
func (gol *GenerationOutputLike) Update() *GenerationOutputLikeUpdateOne {
	return NewGenerationOutputLikeClient(gol.config).UpdateOne(gol)
}

// Unwrap unwraps the GenerationOutputLike entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (gol *GenerationOutputLike) Unwrap() *GenerationOutputLike {
	_tx, ok := gol.config.driver.(*txDriver)
	if !ok {
		panic("ent: GenerationOutputLike is not a transactional entity")
	}
	gol.config.driver = _tx.drv
	return gol
}

// String implements the fmt.Stringer.
func (gol *GenerationOutputLike) String() string {
	var builder strings.Builder
	builder.WriteString("GenerationOutputLike(")
	builder.WriteString(fmt.Sprintf("id=%v, ", gol.ID))
	builder.WriteString("output_id=")
	builder.WriteString(fmt.Sprintf("%v", gol.OutputID))
	builder.WriteString(", ")
	builder.WriteString("liked_by_user_id=")
	builder.WriteString(fmt.Sprintf("%v", gol.LikedByUserID))
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(gol.CreatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// GenerationOutputLikes is a parsable slice of GenerationOutputLike.
type GenerationOutputLikes []*GenerationOutputLike
