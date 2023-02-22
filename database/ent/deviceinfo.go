// Code generated by ent, DO NOT EDIT.

package ent

import (
	"fmt"
	"strings"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/deviceinfo"
)

// DeviceInfo is the model entity for the DeviceInfo schema.
type DeviceInfo struct {
	config `json:"-"`
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// Type holds the value of the "type" field.
	Type *string `json:"type,omitempty"`
	// Os holds the value of the "os" field.
	Os *string `json:"os,omitempty"`
	// Browser holds the value of the "browser" field.
	Browser *string `json:"browser,omitempty"`
	// CreatedAt holds the value of the "created_at" field.
	CreatedAt time.Time `json:"created_at,omitempty"`
	// UpdatedAt holds the value of the "updated_at" field.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the DeviceInfoQuery when eager-loading is set.
	Edges DeviceInfoEdges `json:"edges"`
}

// DeviceInfoEdges holds the relations/edges for other nodes in the graph.
type DeviceInfoEdges struct {
	// Generations holds the value of the generations edge.
	Generations []*Generation `json:"generations,omitempty"`
	// Upscales holds the value of the upscales edge.
	Upscales []*Upscale `json:"upscales,omitempty"`
	// loadedTypes holds the information for reporting if a
	// type was loaded (or requested) in eager-loading or not.
	loadedTypes [2]bool
}

// GenerationsOrErr returns the Generations value or an error if the edge
// was not loaded in eager-loading.
func (e DeviceInfoEdges) GenerationsOrErr() ([]*Generation, error) {
	if e.loadedTypes[0] {
		return e.Generations, nil
	}
	return nil, &NotLoadedError{edge: "generations"}
}

// UpscalesOrErr returns the Upscales value or an error if the edge
// was not loaded in eager-loading.
func (e DeviceInfoEdges) UpscalesOrErr() ([]*Upscale, error) {
	if e.loadedTypes[1] {
		return e.Upscales, nil
	}
	return nil, &NotLoadedError{edge: "upscales"}
}

// scanValues returns the types for scanning values from sql.Rows.
func (*DeviceInfo) scanValues(columns []string) ([]any, error) {
	values := make([]any, len(columns))
	for i := range columns {
		switch columns[i] {
		case deviceinfo.FieldType, deviceinfo.FieldOs, deviceinfo.FieldBrowser:
			values[i] = new(sql.NullString)
		case deviceinfo.FieldCreatedAt, deviceinfo.FieldUpdatedAt:
			values[i] = new(sql.NullTime)
		case deviceinfo.FieldID:
			values[i] = new(uuid.UUID)
		default:
			return nil, fmt.Errorf("unexpected column %q for type DeviceInfo", columns[i])
		}
	}
	return values, nil
}

// assignValues assigns the values that were returned from sql.Rows (after scanning)
// to the DeviceInfo fields.
func (di *DeviceInfo) assignValues(columns []string, values []any) error {
	if m, n := len(values), len(columns); m < n {
		return fmt.Errorf("mismatch number of scan values: %d != %d", m, n)
	}
	for i := range columns {
		switch columns[i] {
		case deviceinfo.FieldID:
			if value, ok := values[i].(*uuid.UUID); !ok {
				return fmt.Errorf("unexpected type %T for field id", values[i])
			} else if value != nil {
				di.ID = *value
			}
		case deviceinfo.FieldType:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field type", values[i])
			} else if value.Valid {
				di.Type = new(string)
				*di.Type = value.String
			}
		case deviceinfo.FieldOs:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field os", values[i])
			} else if value.Valid {
				di.Os = new(string)
				*di.Os = value.String
			}
		case deviceinfo.FieldBrowser:
			if value, ok := values[i].(*sql.NullString); !ok {
				return fmt.Errorf("unexpected type %T for field browser", values[i])
			} else if value.Valid {
				di.Browser = new(string)
				*di.Browser = value.String
			}
		case deviceinfo.FieldCreatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field created_at", values[i])
			} else if value.Valid {
				di.CreatedAt = value.Time
			}
		case deviceinfo.FieldUpdatedAt:
			if value, ok := values[i].(*sql.NullTime); !ok {
				return fmt.Errorf("unexpected type %T for field updated_at", values[i])
			} else if value.Valid {
				di.UpdatedAt = value.Time
			}
		}
	}
	return nil
}

// QueryGenerations queries the "generations" edge of the DeviceInfo entity.
func (di *DeviceInfo) QueryGenerations() *GenerationQuery {
	return NewDeviceInfoClient(di.config).QueryGenerations(di)
}

// QueryUpscales queries the "upscales" edge of the DeviceInfo entity.
func (di *DeviceInfo) QueryUpscales() *UpscaleQuery {
	return NewDeviceInfoClient(di.config).QueryUpscales(di)
}

// Update returns a builder for updating this DeviceInfo.
// Note that you need to call DeviceInfo.Unwrap() before calling this method if this DeviceInfo
// was returned from a transaction, and the transaction was committed or rolled back.
func (di *DeviceInfo) Update() *DeviceInfoUpdateOne {
	return NewDeviceInfoClient(di.config).UpdateOne(di)
}

// Unwrap unwraps the DeviceInfo entity that was returned from a transaction after it was closed,
// so that all future queries will be executed through the driver which created the transaction.
func (di *DeviceInfo) Unwrap() *DeviceInfo {
	_tx, ok := di.config.driver.(*txDriver)
	if !ok {
		panic("ent: DeviceInfo is not a transactional entity")
	}
	di.config.driver = _tx.drv
	return di
}

// String implements the fmt.Stringer.
func (di *DeviceInfo) String() string {
	var builder strings.Builder
	builder.WriteString("DeviceInfo(")
	builder.WriteString(fmt.Sprintf("id=%v, ", di.ID))
	if v := di.Type; v != nil {
		builder.WriteString("type=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	if v := di.Os; v != nil {
		builder.WriteString("os=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	if v := di.Browser; v != nil {
		builder.WriteString("browser=")
		builder.WriteString(*v)
	}
	builder.WriteString(", ")
	builder.WriteString("created_at=")
	builder.WriteString(di.CreatedAt.Format(time.ANSIC))
	builder.WriteString(", ")
	builder.WriteString("updated_at=")
	builder.WriteString(di.UpdatedAt.Format(time.ANSIC))
	builder.WriteByte(')')
	return builder.String()
}

// DeviceInfos is a parsable slice of DeviceInfo.
type DeviceInfos []*DeviceInfo

func (di DeviceInfos) config(cfg config) {
	for _i := range di {
		di[_i].config = cfg
	}
}
