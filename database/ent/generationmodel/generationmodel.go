// Code generated by ent, DO NOT EDIT.

package generationmodel

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the generationmodel type in the database.
	Label = "generation_model"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldNameInWorker holds the string denoting the name_in_worker field in the database.
	FieldNameInWorker = "name_in_worker"
	// FieldShortName holds the string denoting the short_name field in the database.
	FieldShortName = "short_name"
	// FieldIsActive holds the string denoting the is_active field in the database.
	FieldIsActive = "is_active"
	// FieldIsDefault holds the string denoting the is_default field in the database.
	FieldIsDefault = "is_default"
	// FieldIsHidden holds the string denoting the is_hidden field in the database.
	FieldIsHidden = "is_hidden"
	// FieldRunpodEndpoint holds the string denoting the runpod_endpoint field in the database.
	FieldRunpodEndpoint = "runpod_endpoint"
	// FieldDisplayWeight holds the string denoting the display_weight field in the database.
	FieldDisplayWeight = "display_weight"
	// FieldDefaultSchedulerID holds the string denoting the default_scheduler_id field in the database.
	FieldDefaultSchedulerID = "default_scheduler_id"
	// FieldDefaultWidth holds the string denoting the default_width field in the database.
	FieldDefaultWidth = "default_width"
	// FieldDefaultHeight holds the string denoting the default_height field in the database.
	FieldDefaultHeight = "default_height"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// EdgeGenerations holds the string denoting the generations edge name in mutations.
	EdgeGenerations = "generations"
	// EdgeSchedulers holds the string denoting the schedulers edge name in mutations.
	EdgeSchedulers = "schedulers"
	// Table holds the table name of the generationmodel in the database.
	Table = "generation_models"
	// GenerationsTable is the table that holds the generations relation/edge.
	GenerationsTable = "generations"
	// GenerationsInverseTable is the table name for the Generation entity.
	// It exists in this package in order to avoid circular dependency with the "generation" package.
	GenerationsInverseTable = "generations"
	// GenerationsColumn is the table column denoting the generations relation/edge.
	GenerationsColumn = "model_id"
	// SchedulersTable is the table that holds the schedulers relation/edge. The primary key declared below.
	SchedulersTable = "generation_model_compatible_schedulers"
	// SchedulersInverseTable is the table name for the Scheduler entity.
	// It exists in this package in order to avoid circular dependency with the "scheduler" package.
	SchedulersInverseTable = "schedulers"
)

// Columns holds all SQL columns for generationmodel fields.
var Columns = []string{
	FieldID,
	FieldNameInWorker,
	FieldShortName,
	FieldIsActive,
	FieldIsDefault,
	FieldIsHidden,
	FieldRunpodEndpoint,
	FieldDisplayWeight,
	FieldDefaultSchedulerID,
	FieldDefaultWidth,
	FieldDefaultHeight,
	FieldCreatedAt,
	FieldUpdatedAt,
}

var (
	// SchedulersPrimaryKey and SchedulersColumn2 are the table columns denoting the
	// primary key for the schedulers relation (M2M).
	SchedulersPrimaryKey = []string{"generation_model_id", "scheduler_id"}
)

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultIsActive holds the default value on creation for the "is_active" field.
	DefaultIsActive bool
	// DefaultIsDefault holds the default value on creation for the "is_default" field.
	DefaultIsDefault bool
	// DefaultIsHidden holds the default value on creation for the "is_hidden" field.
	DefaultIsHidden bool
	// DefaultDisplayWeight holds the default value on creation for the "display_weight" field.
	DefaultDisplayWeight int32
	// DefaultDefaultWidth holds the default value on creation for the "default_width" field.
	DefaultDefaultWidth int32
	// DefaultDefaultHeight holds the default value on creation for the "default_height" field.
	DefaultDefaultHeight int32
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the GenerationModel queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByNameInWorker orders the results by the name_in_worker field.
func ByNameInWorker(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldNameInWorker, opts...).ToFunc()
}

// ByShortName orders the results by the short_name field.
func ByShortName(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldShortName, opts...).ToFunc()
}

// ByIsActive orders the results by the is_active field.
func ByIsActive(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsActive, opts...).ToFunc()
}

// ByIsDefault orders the results by the is_default field.
func ByIsDefault(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsDefault, opts...).ToFunc()
}

// ByIsHidden orders the results by the is_hidden field.
func ByIsHidden(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsHidden, opts...).ToFunc()
}

// ByRunpodEndpoint orders the results by the runpod_endpoint field.
func ByRunpodEndpoint(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldRunpodEndpoint, opts...).ToFunc()
}

// ByDisplayWeight orders the results by the display_weight field.
func ByDisplayWeight(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDisplayWeight, opts...).ToFunc()
}

// ByDefaultSchedulerID orders the results by the default_scheduler_id field.
func ByDefaultSchedulerID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDefaultSchedulerID, opts...).ToFunc()
}

// ByDefaultWidth orders the results by the default_width field.
func ByDefaultWidth(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDefaultWidth, opts...).ToFunc()
}

// ByDefaultHeight orders the results by the default_height field.
func ByDefaultHeight(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDefaultHeight, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByGenerationsCount orders the results by generations count.
func ByGenerationsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newGenerationsStep(), opts...)
	}
}

// ByGenerations orders the results by generations terms.
func ByGenerations(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newGenerationsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// BySchedulersCount orders the results by schedulers count.
func BySchedulersCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newSchedulersStep(), opts...)
	}
}

// BySchedulers orders the results by schedulers terms.
func BySchedulers(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newSchedulersStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newGenerationsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(GenerationsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, GenerationsTable, GenerationsColumn),
	)
}
func newSchedulersStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(SchedulersInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, false, SchedulersTable, SchedulersPrimaryKey...),
	)
}
