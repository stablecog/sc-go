// Code generated by ent, DO NOT EDIT.

package mqlog

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the mqlog type in the database.
	Label = "mq_log"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldMessageID holds the string denoting the message_id field in the database.
	FieldMessageID = "message_id"
	// FieldPriority holds the string denoting the priority field in the database.
	FieldPriority = "priority"
	// FieldIsProcessing holds the string denoting the is_processing field in the database.
	FieldIsProcessing = "is_processing"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// Table holds the table name of the mqlog in the database.
	Table = "mq_log"
)

// Columns holds all SQL columns for mqlog fields.
var Columns = []string{
	FieldID,
	FieldMessageID,
	FieldPriority,
	FieldIsProcessing,
	FieldCreatedAt,
	FieldUpdatedAt,
}

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
	// DefaultIsProcessing holds the default value on creation for the "is_processing" field.
	DefaultIsProcessing bool
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the MqLog queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByMessageID orders the results by the message_id field.
func ByMessageID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldMessageID, opts...).ToFunc()
}

// ByPriority orders the results by the priority field.
func ByPriority(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPriority, opts...).ToFunc()
}

// ByIsProcessing orders the results by the is_processing field.
func ByIsProcessing(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldIsProcessing, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}
