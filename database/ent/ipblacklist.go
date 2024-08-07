// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/ipblacklist"
)

// IPBlackList is the model entity for the IPBlackList schema.
type IPBlackList struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// IP holds the value of the "ip" field.
	IP string `json:"ip,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
	selectValues sql.SelectValues
}

// scanValues returns the types for scanning values from sql.Rows.
func (*IPBlackList) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case ipblacklist.FieldIP:
			values[i] = new(sql.NullString)
		case ipblacklist.FieldCreatedAt, ipblacklist.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case ipblacklist.FieldID:
			values[i] = new(uuid.UUID)
		default:
			values[i] = new(sql.UnknownType)
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the IPBlackList fields.
func (ibl *IPBlackList) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case ipblacklist.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				ibl.ID = *value
			}
		case ipblacklist.FieldIP:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field ip", values[i])
			} else if value.Valid {
				ibl.IP = value.String
			}
		case ipblacklist.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				ibl.CreatedAt = value.Time
			}
		case ipblacklist.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				ibl.UpdatedAt = value.Time
			}
		default:
			ibl.selectValues.Set(columns[i], values[i])
		}
	}
	return nil
}

// Value returns the ent.Value that was dynamically selected and assigned to the IPBlackList.
// This includes values selected through modifiers, order, etc.
func (ibl *IPBlackList) Value(name string) (ent.Value, error) {
	return ibl.selectValues.Get(name)
}

// Update returns a builder for updating this IPBlackList.
// Note that you need to call IPBlackList.Unwrap() before calling this method if this IPBlackList
// was returned from a transaction, and the transaction was committed or rolled back.
func (ibl *IPBlackList) Update() *IPBlackListUpdateOne {
	return NewIPBlackListClient(ibl.config).UpdateOne(ibl)
}

// Unwrap unwraps the IPBlackList entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (ibl *IPBlackList) Unwrap() *IPBlackList {
	_tx, ok := ibl.config.driver.(*txDriver)
	if !ok {
		panic("ent: IPBlackList is not a transactional entity")
	}
	ibl.config.driver = _tx.drv
	return ibl
}

// String implements the fmt.Stringer.
func (ibl *IPBlackList) String() string {
	var builder strings.Builder
	builder.WriteString("IPBlackList(")
	builder.WriteString(fmt.Sprintf("id=%v, ", ibl.ID))
	builder.WriteString("ip=")
	builder.WriteString(ibl.IP)
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(ibl.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(ibl.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// IPBlackLists is a parsable slice of IPBlackList.
type IPBlackLists []*IPBlackList
