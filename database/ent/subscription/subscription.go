// Code generated by ent, DO NOT EDIT.

package subscription

import (
	"time"

	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the subscription type in the database.
	Label = "subscription"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldUserID holds the string denoting the user_id field in the database.
	FieldUserID = "user_id"
	// FieldSubscriptionTierID holds the string denoting the subscription_tier_id field in the database.
	FieldSubscriptionTierID = "subscription_tier_id"
	// FieldPaidStartedAt holds the string denoting the paid_started_at field in the database.
	FieldPaidStartedAt = "paid_started_at"
	// FieldPaidCancelledAt holds the string denoting the paid_cancelled_at field in the database.
	FieldPaidCancelledAt = "paid_cancelled_at"
	// FieldPaidExpiresAt holds the string denoting the paid_expires_at field in the database.
	FieldPaidExpiresAt = "paid_expires_at"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// EdgeUser holds the string denoting the user edge name in mutations.
	EdgeUser = "user"
	// EdgeSubscriptionTier holds the string denoting the subscription_tier edge name in mutations.
	EdgeSubscriptionTier = "subscription_tier"
	// Table holds the table name of the subscription in the database.
	Table = "subscriptions"
	// UserTable is the table that holds the user relation/edge.
	UserTable = "subscriptions"
	// UserInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	UserInverseTable = "users"
	// UserColumn is the table column denoting the user relation/edge.
	UserColumn = "user_id"
	// SubscriptionTierTable is the table that holds the subscription_tier relation/edge.
	SubscriptionTierTable = "subscriptions"
	// SubscriptionTierInverseTable is the table name for the SubscriptionTier entity.
	// It exists in this package in order to avoid circular dependency with the "subscriptiontier" package.
	SubscriptionTierInverseTable = "subscription_tiers"
	// SubscriptionTierColumn is the table column denoting the subscription_tier relation/edge.
	SubscriptionTierColumn = "subscription_tier_id"
)

// Columns holds all SQL columns for subscription fields.
var Columns = []string{
	FieldID,
	FieldUserID,
	FieldSubscriptionTierID,
	FieldPaidStartedAt,
	FieldPaidCancelledAt,
	FieldPaidExpiresAt,
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
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)
