// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/tiplog"
	"github.com/stablecog/sc-go/database/ent/user"
)

// TipLog is the model entity for the TipLog schema.
type TipLog struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// Amount holds the value of the "amount" field.
	Amount int32 `json:"amount,omitempty"`
	// TippedToDiscordID holds the value of the "tipped_to_discord_id" field.
	TippedToDiscordID string `json:"tipped_to_discord_id,omitempty"`
	// TippedBy holds the value of the "tipped_by" field.
	TippedBy uuid.UUID `json:"tipped_by,omitempty"`
	// TippedTo holds the value of the "tipped_to" field.
	TippedTo *uuid.UUID `json:"tipped_to,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the TipLogQuery when eager-loading is set.
	Edges        TipLogEdges `json:"edges"`
	selectValues sql.SelectValues
}

// TipLogEdges holds the relations/edges for other nodes in the graph.
type TipLogEdges struct {
	// TipsReceived holds the value of the tips_received edge.
	TipsReceived *User `json:"tips_received,omitempty"`
	// TipsGiven holds the value of the tips_given edge.
	TipsGiven *User `json:"tips_given,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// TipsReceivedOrErr returns the TipsReceived value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e TipLogEdges) TipsReceivedOrErr() (*User, error) {
	if e.TipsReceived != nil {
		return e.TipsReceived, nil
	} else if e.loadedTypes[0] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "tips_received"}
}

// TipsGivenOrErr returns the TipsGiven value or an error if the edge
// was not loaded in eager-loading, or loaded but was not found.
func (e TipLogEdges) TipsGivenOrErr() (*User, error) {
	if e.TipsGiven != nil {
		return e.TipsGiven, nil
	} else if e.loadedTypes[1] {
		return nil, &NotFoundError{label: user.Label}
	}
	return nil, &NotLoadedError{edge: "tips_given"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*TipLog) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case tiplog.FieldTippedTo:
			values[i] = &sql.NullScanner{S: new(uuid.UUID)}
		case tiplog.FieldAmount:
			values[i] = new(sql.NullInt64)
		case tiplog.FieldTippedToDiscordID:
			values[i] = new(sql.NullString)
		case tiplog.FieldCreatedAt, tiplog.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case tiplog.FieldID, tiplog.FieldTippedBy:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the TipLog fields.
func (tl *TipLog) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case tiplog.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				tl.ID = *value
			}
		case tiplog.FieldAmount:
			if value, ok := values[i].(*sql.NullInt64); !ok {
				return fmt.Errorf("unexpected type %T for field amount", values[i])
			} else if value.Valid {
				tl.Amount = int32(value.Int64)
			}
		case tiplog.FieldTippedToDiscordID:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field tipped_to_discord_id", values[i])
			} else if value.Valid {
				tl.TippedToDiscordID = value.String
			}
		case tiplog.FieldTippedBy:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field tipped_by", values[i])
			} else if value != nil {
				tl.TippedBy = *value
			}
		case tiplog.FieldTippedTo:
			if value, ok := values[i].(*sql.NullScanner); !ok {
				return fmt.Errorf("unexpected type %T for field tipped_to", values[i])
			} else if value.Valid {
				tl.TippedTo = new(uuid.UUID)
				*tl.TippedTo = *value.S.(*uuid.UUID)
			}
		case tiplog.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				tl.CreatedAt = value.Time
			}
		case tiplog.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				tl.UpdatedAt = value.Time
			}
		default:
			tl.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the TipLog.
// This includes values selected through modifiers, order, etc.
func (tl *TipLog) Value(name string) (ent.Value, error) {
	return tl.selectValues.Get(name)
}

// QueryTipsReceived queries the "tips_received" edge of the TipLog entity.
func (tl *TipLog) QueryTipsReceived() *UserQuery {
	return NewTipLogClient(tl.config).QueryTipsReceived(tl)
}

// QueryTipsGiven queries the "tips_given" edge of the TipLog entity.
func (tl *TipLog) QueryTipsGiven() *UserQuery {
	return NewTipLogClient(tl.config).QueryTipsGiven(tl)
}

// Update returns a builder for updating this TipLog.
// Note that you need to call TipLog.Unwrap() before calling this method if this TipLog
// was returned from a transaction, and the transaction was committed or rolled back.
func (tl *TipLog) Update() *TipLogUpdateOne {
	return NewTipLogClient(tl.config).UpdateOne(tl)
}

// Unwrap unwraps the TipLog entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (tl *TipLog) Unwrap() *TipLog {
	_tx, ok := tl.config.driver.(*txDriver)
	if !ok {
		panic("ent: TipLog is not a transactional entity")
	}
	tl.config.driver = _tx.drv
	return tl
}

// String implements the fmt.Stringer.
func (tl *TipLog) String() string {
	var builder strings.Builder
	builder.WriteString("TipLog(")
	builder.WriteString(fmt.Sprintf("id=%v, ", tl.ID))
	builder.WriteString("amount=")
	builder.WriteString(fmt.Sprintf("%v", tl.Amount))
	builder.WriteString(", ")
	builder.WriteString("tipped_to_discord_id=")
	builder.WriteString(tl.TippedToDiscordID)
	builder.WriteString(", ")
	builder.WriteString("tipped_by=")
	builder.WriteString(fmt.Sprintf("%v", tl.TippedBy))
	builder.WriteString(", ")
	if v := tl.TippedTo; v != nil {
		builder.WriteString("tipped_to=")
		builder.WriteString(fmt.Sprintf("%v", *v))
	}
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(tl.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(tl.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// TipLogs is a parsable slice of TipLog.
type TipLogs []*TipLog
