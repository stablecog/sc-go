// Code generated by ent, DO NOT EDIT.

package user

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the user type in the database.
	Label = "user"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldEmail holds the string denoting the email field in the database.
	FieldEmail = "email"
	// FieldEmailNormalized holds the string denoting the email_normalized field in the database.
	FieldEmailNormalized = "email_normalized"
	// FieldStripeCustomerID holds the string denoting the stripe_customer_id field in the database.
	FieldStripeCustomerID = "stripe_customer_id"
	// FieldActiveProductID holds the string denoting the active_product_id field in the database.
	FieldActiveProductID = "active_product_id"
	// FieldLastSignInAt holds the string denoting the last_sign_in_at field in the database.
	FieldLastSignInAt = "last_sign_in_at"
	// FieldLastSeenAt holds the string denoting the last_seen_at field in the database.
	FieldLastSeenAt = "last_seen_at"
	// FieldBannedAt holds the string denoting the banned_at field in the database.
	FieldBannedAt = "banned_at"
	// FieldScheduledForDeletionOn holds the string denoting the scheduled_for_deletion_on field in the database.
	FieldScheduledForDeletionOn = "scheduled_for_deletion_on"
	// FieldDataDeletedAt holds the string denoting the data_deleted_at field in the database.
	FieldDataDeletedAt = "data_deleted_at"
	// FieldWantsEmail holds the string denoting the wants_email field in the database.
	FieldWantsEmail = "wants_email"
	// FieldDiscordID holds the string denoting the discord_id field in the database.
	FieldDiscordID = "discord_id"
	// FieldUsername holds the string denoting the username field in the database.
	FieldUsername = "username"
	// FieldUsernameChangedAt holds the string denoting the username_changed_at field in the database.
	FieldUsernameChangedAt = "username_changed_at"
	// FieldStripeHighestProductID holds the string denoting the stripe_highest_product_id field in the database.
	FieldStripeHighestProductID = "stripe_highest_product_id"
	// FieldStripeHighestPriceID holds the string denoting the stripe_highest_price_id field in the database.
	FieldStripeHighestPriceID = "stripe_highest_price_id"
	// FieldStripeCancelsAt holds the string denoting the stripe_cancels_at field in the database.
	FieldStripeCancelsAt = "stripe_cancels_at"
	// FieldStripeSyncedAt holds the string denoting the stripe_synced_at field in the database.
	FieldStripeSyncedAt = "stripe_synced_at"
	// FieldStripeRenewsAt holds the string denoting the stripe_renews_at field in the database.
	FieldStripeRenewsAt = "stripe_renews_at"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// EdgeGenerations holds the string denoting the generations edge name in mutations.
	EdgeGenerations = "generations"
	// EdgeUpscales holds the string denoting the upscales edge name in mutations.
	EdgeUpscales = "upscales"
	// EdgeVoiceovers holds the string denoting the voiceovers edge name in mutations.
	EdgeVoiceovers = "voiceovers"
	// EdgeCredits holds the string denoting the credits edge name in mutations.
	EdgeCredits = "credits"
	// EdgeAPITokens holds the string denoting the api_tokens edge name in mutations.
	EdgeAPITokens = "api_tokens"
	// EdgeTipsGiven holds the string denoting the tips_given edge name in mutations.
	EdgeTipsGiven = "tips_given"
	// EdgeTipsReceived holds the string denoting the tips_received edge name in mutations.
	EdgeTipsReceived = "tips_received"
	// EdgeRoles holds the string denoting the roles edge name in mutations.
	EdgeRoles = "roles"
	// EdgeGenerationOutputLikes holds the string denoting the generation_output_likes edge name in mutations.
	EdgeGenerationOutputLikes = "generation_output_likes"
	// Table holds the table name of the user in the database.
	Table = "users"
	// GenerationsTable is the table that holds the generations relation/edge.
	GenerationsTable = "generations"
	// GenerationsInverseTable is the table name for the Generation entity.
	// It exists in this package in order to avoid circular dependency with the "generation" package.
	GenerationsInverseTable = "generations"
	// GenerationsColumn is the table column denoting the generations relation/edge.
	GenerationsColumn = "user_id"
	// UpscalesTable is the table that holds the upscales relation/edge.
	UpscalesTable = "upscales"
	// UpscalesInverseTable is the table name for the Upscale entity.
	// It exists in this package in order to avoid circular dependency with the "upscale" package.
	UpscalesInverseTable = "upscales"
	// UpscalesColumn is the table column denoting the upscales relation/edge.
	UpscalesColumn = "user_id"
	// VoiceoversTable is the table that holds the voiceovers relation/edge.
	VoiceoversTable = "voiceovers"
	// VoiceoversInverseTable is the table name for the Voiceover entity.
	// It exists in this package in order to avoid circular dependency with the "voiceover" package.
	VoiceoversInverseTable = "voiceovers"
	// VoiceoversColumn is the table column denoting the voiceovers relation/edge.
	VoiceoversColumn = "user_id"
	// CreditsTable is the table that holds the credits relation/edge.
	CreditsTable = "credits"
	// CreditsInverseTable is the table name for the Credit entity.
	// It exists in this package in order to avoid circular dependency with the "credit" package.
	CreditsInverseTable = "credits"
	// CreditsColumn is the table column denoting the credits relation/edge.
	CreditsColumn = "user_id"
	// APITokensTable is the table that holds the api_tokens relation/edge.
	APITokensTable = "api_tokens"
	// APITokensInverseTable is the table name for the ApiToken entity.
	// It exists in this package in order to avoid circular dependency with the "apitoken" package.
	APITokensInverseTable = "api_tokens"
	// APITokensColumn is the table column denoting the api_tokens relation/edge.
	APITokensColumn = "user_id"
	// TipsGivenTable is the table that holds the tips_given relation/edge.
	TipsGivenTable = "tip_log"
	// TipsGivenInverseTable is the table name for the TipLog entity.
	// It exists in this package in order to avoid circular dependency with the "tiplog" package.
	TipsGivenInverseTable = "tip_log"
	// TipsGivenColumn is the table column denoting the tips_given relation/edge.
	TipsGivenColumn = "tipped_by"
	// TipsReceivedTable is the table that holds the tips_received relation/edge.
	TipsReceivedTable = "tip_log"
	// TipsReceivedInverseTable is the table name for the TipLog entity.
	// It exists in this package in order to avoid circular dependency with the "tiplog" package.
	TipsReceivedInverseTable = "tip_log"
	// TipsReceivedColumn is the table column denoting the tips_received relation/edge.
	TipsReceivedColumn = "tipped_to"
	// RolesTable is the table that holds the roles relation/edge. The primary key declared below.
	RolesTable = "user_role_users"
	// RolesInverseTable is the table name for the Role entity.
	// It exists in this package in order to avoid circular dependency with the "role" package.
	RolesInverseTable = "roles"
	// GenerationOutputLikesTable is the table that holds the generation_output_likes relation/edge.
	GenerationOutputLikesTable = "generation_output_likes"
	// GenerationOutputLikesInverseTable is the table name for the GenerationOutputLike entity.
	// It exists in this package in order to avoid circular dependency with the "generationoutputlike" package.
	GenerationOutputLikesInverseTable = "generation_output_likes"
	// GenerationOutputLikesColumn is the table column denoting the generation_output_likes relation/edge.
	GenerationOutputLikesColumn = "liked_by_user_id"
)

// Columns holds all SQL columns for user fields.
var Columns = []string{
	FieldID,
	FieldEmail,
	FieldEmailNormalized,
	FieldStripeCustomerID,
	FieldActiveProductID,
	FieldLastSignInAt,
	FieldLastSeenAt,
	FieldBannedAt,
	FieldScheduledForDeletionOn,
	FieldDataDeletedAt,
	FieldWantsEmail,
	FieldDiscordID,
	FieldUsername,
	FieldUsernameChangedAt,
	FieldStripeHighestProductID,
	FieldStripeHighestPriceID,
	FieldStripeCancelsAt,
	FieldStripeSyncedAt,
	FieldStripeRenewsAt,
	FieldCreatedAt,
	FieldUpdatedAt,
}

var (
	// RolesPrimaryKey and RolesColumn2 are the table columns denoting the
	// primary key for the roles relation (M2M).
	RolesPrimaryKey = []string{"role_id", "user_id"}
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
	// DefaultLastSeenAt holds the default value on creation for the "last_seen_at" field.
	DefaultLastSeenAt func() time.Time
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the User queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByEmail orders the results by the email field.
func ByEmail(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldEmail, opts...).ToFunc()
}

// ByEmailNormalized orders the results by the email_normalized field.
func ByEmailNormalized(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldEmailNormalized, opts...).ToFunc()
}

// ByStripeCustomerID orders the results by the stripe_customer_id field.
func ByStripeCustomerID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStripeCustomerID, opts...).ToFunc()
}

// ByActiveProductID orders the results by the active_product_id field.
func ByActiveProductID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldActiveProductID, opts...).ToFunc()
}

// ByLastSignInAt orders the results by the last_sign_in_at field.
func ByLastSignInAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldLastSignInAt, opts...).ToFunc()
}

// ByLastSeenAt orders the results by the last_seen_at field.
func ByLastSeenAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldLastSeenAt, opts...).ToFunc()
}

// ByBannedAt orders the results by the banned_at field.
func ByBannedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldBannedAt, opts...).ToFunc()
}

// ByScheduledForDeletionOn orders the results by the scheduled_for_deletion_on field.
func ByScheduledForDeletionOn(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldScheduledForDeletionOn, opts...).ToFunc()
}

// ByDataDeletedAt orders the results by the data_deleted_at field.
func ByDataDeletedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDataDeletedAt, opts...).ToFunc()
}

// ByWantsEmail orders the results by the wants_email field.
func ByWantsEmail(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldWantsEmail, opts...).ToFunc()
}

// ByDiscordID orders the results by the discord_id field.
func ByDiscordID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDiscordID, opts...).ToFunc()
}

// ByUsername orders the results by the username field.
func ByUsername(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUsername, opts...).ToFunc()
}

// ByUsernameChangedAt orders the results by the username_changed_at field.
func ByUsernameChangedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUsernameChangedAt, opts...).ToFunc()
}

// ByStripeHighestProductID orders the results by the stripe_highest_product_id field.
func ByStripeHighestProductID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStripeHighestProductID, opts...).ToFunc()
}

// ByStripeHighestPriceID orders the results by the stripe_highest_price_id field.
func ByStripeHighestPriceID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStripeHighestPriceID, opts...).ToFunc()
}

// ByStripeCancelsAt orders the results by the stripe_cancels_at field.
func ByStripeCancelsAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStripeCancelsAt, opts...).ToFunc()
}

// ByStripeSyncedAt orders the results by the stripe_synced_at field.
func ByStripeSyncedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStripeSyncedAt, opts...).ToFunc()
}

// ByStripeRenewsAt orders the results by the stripe_renews_at field.
func ByStripeRenewsAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStripeRenewsAt, opts...).ToFunc()
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

// ByUpscalesCount orders the results by upscales count.
func ByUpscalesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newUpscalesStep(), opts...)
	}
}

// ByUpscales orders the results by upscales terms.
func ByUpscales(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newUpscalesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByVoiceoversCount orders the results by voiceovers count.
func ByVoiceoversCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newVoiceoversStep(), opts...)
	}
}

// ByVoiceovers orders the results by voiceovers terms.
func ByVoiceovers(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newVoiceoversStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByCreditsCount orders the results by credits count.
func ByCreditsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newCreditsStep(), opts...)
	}
}

// ByCredits orders the results by credits terms.
func ByCredits(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newCreditsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByAPITokensCount orders the results by api_tokens count.
func ByAPITokensCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newAPITokensStep(), opts...)
	}
}

// ByAPITokens orders the results by api_tokens terms.
func ByAPITokens(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newAPITokensStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByTipsGivenCount orders the results by tips_given count.
func ByTipsGivenCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newTipsGivenStep(), opts...)
	}
}

// ByTipsGiven orders the results by tips_given terms.
func ByTipsGiven(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newTipsGivenStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByTipsReceivedCount orders the results by tips_received count.
func ByTipsReceivedCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newTipsReceivedStep(), opts...)
	}
}

// ByTipsReceived orders the results by tips_received terms.
func ByTipsReceived(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newTipsReceivedStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByRolesCount orders the results by roles count.
func ByRolesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newRolesStep(), opts...)
	}
}

// ByRoles orders the results by roles terms.
func ByRoles(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newRolesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByGenerationOutputLikesCount orders the results by generation_output_likes count.
func ByGenerationOutputLikesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newGenerationOutputLikesStep(), opts...)
	}
}

// ByGenerationOutputLikes orders the results by generation_output_likes terms.
func ByGenerationOutputLikes(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newGenerationOutputLikesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newGenerationsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(GenerationsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, GenerationsTable, GenerationsColumn),
	)
}
func newUpscalesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(UpscalesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, UpscalesTable, UpscalesColumn),
	)
}
func newVoiceoversStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(VoiceoversInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, VoiceoversTable, VoiceoversColumn),
	)
}
func newCreditsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(CreditsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, CreditsTable, CreditsColumn),
	)
}
func newAPITokensStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(APITokensInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, APITokensTable, APITokensColumn),
	)
}
func newTipsGivenStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(TipsGivenInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, TipsGivenTable, TipsGivenColumn),
	)
}
func newTipsReceivedStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(TipsReceivedInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, TipsReceivedTable, TipsReceivedColumn),
	)
}
func newRolesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(RolesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2M, true, RolesTable, RolesPrimaryKey...),
	)
}
func newGenerationOutputLikesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(GenerationOutputLikesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, GenerationOutputLikesTable, GenerationOutputLikesColumn),
	)
}
