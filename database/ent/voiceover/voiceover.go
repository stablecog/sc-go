// Code generated by ent, DO NOT EDIT.

package voiceover

import (
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/enttypes"
)

const (
	// Label holds the string label denoting the voiceover type in the database.
	Label = "voiceover"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCountryCode holds the string denoting the country_code field in the database.
	FieldCountryCode = "country_code"
	// FieldStatus holds the string denoting the status field in the database.
	FieldStatus = "status"
	// FieldFailureReason holds the string denoting the failure_reason field in the database.
	FieldFailureReason = "failure_reason"
	// FieldStripeProductID holds the string denoting the stripe_product_id field in the database.
	FieldStripeProductID = "stripe_product_id"
	// FieldTemperature holds the string denoting the temperature field in the database.
	FieldTemperature = "temperature"
	// FieldSeed holds the string denoting the seed field in the database.
	FieldSeed = "seed"
	// FieldWasAutoSubmitted holds the string denoting the was_auto_submitted field in the database.
	FieldWasAutoSubmitted = "was_auto_submitted"
	// FieldDenoiseAudio holds the string denoting the denoise_audio field in the database.
	FieldDenoiseAudio = "denoise_audio"
	// FieldRemoveSilence holds the string denoting the remove_silence field in the database.
	FieldRemoveSilence = "remove_silence"
	// FieldCost holds the string denoting the cost field in the database.
	FieldCost = "cost"
	// FieldSourceType holds the string denoting the source_type field in the database.
	FieldSourceType = "source_type"
	// FieldPromptID holds the string denoting the prompt_id field in the database.
	FieldPromptID = "prompt_id"
	// FieldUserID holds the string denoting the user_id field in the database.
	FieldUserID = "user_id"
	// FieldDeviceInfoID holds the string denoting the device_info_id field in the database.
	FieldDeviceInfoID = "device_info_id"
	// FieldModelID holds the string denoting the model_id field in the database.
	FieldModelID = "model_id"
	// FieldSpeakerID holds the string denoting the speaker_id field in the database.
	FieldSpeakerID = "speaker_id"
	// FieldAPITokenID holds the string denoting the api_token_id field in the database.
	FieldAPITokenID = "api_token_id"
	// FieldStartedAt holds the string denoting the started_at field in the database.
	FieldStartedAt = "started_at"
	// FieldCompletedAt holds the string denoting the completed_at field in the database.
	FieldCompletedAt = "completed_at"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// EdgeUser holds the string denoting the user edge name in mutations.
	EdgeUser = "user"
	// EdgePrompt holds the string denoting the prompt edge name in mutations.
	EdgePrompt = "prompt"
	// EdgeDeviceInfo holds the string denoting the device_info edge name in mutations.
	EdgeDeviceInfo = "device_info"
	// EdgeVoiceoverModels holds the string denoting the voiceover_models edge name in mutations.
	EdgeVoiceoverModels = "voiceover_models"
	// EdgeVoiceoverSpeakers holds the string denoting the voiceover_speakers edge name in mutations.
	EdgeVoiceoverSpeakers = "voiceover_speakers"
	// EdgeAPITokens holds the string denoting the api_tokens edge name in mutations.
	EdgeAPITokens = "api_tokens"
	// EdgeVoiceoverOutputs holds the string denoting the voiceover_outputs edge name in mutations.
	EdgeVoiceoverOutputs = "voiceover_outputs"
	// Table holds the table name of the voiceover in the database.
	Table = "voiceovers"
	// UserTable is the table that holds the user relation/edge.
	UserTable = "voiceovers"
	// UserInverseTable is the table name for the User entity.
	// It exists in this package in order to avoid circular dependency with the "user" package.
	UserInverseTable = "users"
	// UserColumn is the table column denoting the user relation/edge.
	UserColumn = "user_id"
	// PromptTable is the table that holds the prompt relation/edge.
	PromptTable = "voiceovers"
	// PromptInverseTable is the table name for the Prompt entity.
	// It exists in this package in order to avoid circular dependency with the "prompt" package.
	PromptInverseTable = "prompts"
	// PromptColumn is the table column denoting the prompt relation/edge.
	PromptColumn = "prompt_id"
	// DeviceInfoTable is the table that holds the device_info relation/edge.
	DeviceInfoTable = "voiceovers"
	// DeviceInfoInverseTable is the table name for the DeviceInfo entity.
	// It exists in this package in order to avoid circular dependency with the "deviceinfo" package.
	DeviceInfoInverseTable = "device_info"
	// DeviceInfoColumn is the table column denoting the device_info relation/edge.
	DeviceInfoColumn = "device_info_id"
	// VoiceoverModelsTable is the table that holds the voiceover_models relation/edge.
	VoiceoverModelsTable = "voiceovers"
	// VoiceoverModelsInverseTable is the table name for the VoiceoverModel entity.
	// It exists in this package in order to avoid circular dependency with the "voiceovermodel" package.
	VoiceoverModelsInverseTable = "voiceover_models"
	// VoiceoverModelsColumn is the table column denoting the voiceover_models relation/edge.
	VoiceoverModelsColumn = "model_id"
	// VoiceoverSpeakersTable is the table that holds the voiceover_speakers relation/edge.
	VoiceoverSpeakersTable = "voiceovers"
	// VoiceoverSpeakersInverseTable is the table name for the VoiceoverSpeaker entity.
	// It exists in this package in order to avoid circular dependency with the "voiceoverspeaker" package.
	VoiceoverSpeakersInverseTable = "voiceover_speakers"
	// VoiceoverSpeakersColumn is the table column denoting the voiceover_speakers relation/edge.
	VoiceoverSpeakersColumn = "speaker_id"
	// APITokensTable is the table that holds the api_tokens relation/edge.
	APITokensTable = "voiceovers"
	// APITokensInverseTable is the table name for the ApiToken entity.
	// It exists in this package in order to avoid circular dependency with the "apitoken" package.
	APITokensInverseTable = "api_tokens"
	// APITokensColumn is the table column denoting the api_tokens relation/edge.
	APITokensColumn = "api_token_id"
	// VoiceoverOutputsTable is the table that holds the voiceover_outputs relation/edge.
	VoiceoverOutputsTable = "voiceover_outputs"
	// VoiceoverOutputsInverseTable is the table name for the VoiceoverOutput entity.
	// It exists in this package in order to avoid circular dependency with the "voiceoveroutput" package.
	VoiceoverOutputsInverseTable = "voiceover_outputs"
	// VoiceoverOutputsColumn is the table column denoting the voiceover_outputs relation/edge.
	VoiceoverOutputsColumn = "voiceover_id"
)

// Columns holds all SQL columns for voiceover fields.
var Columns = []string{
	FieldID,
	FieldCountryCode,
	FieldStatus,
	FieldFailureReason,
	FieldStripeProductID,
	FieldTemperature,
	FieldSeed,
	FieldWasAutoSubmitted,
	FieldDenoiseAudio,
	FieldRemoveSilence,
	FieldCost,
	FieldSourceType,
	FieldPromptID,
	FieldUserID,
	FieldDeviceInfoID,
	FieldModelID,
	FieldSpeakerID,
	FieldAPITokenID,
	FieldStartedAt,
	FieldCompletedAt,
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
	// DefaultWasAutoSubmitted holds the default value on creation for the "was_auto_submitted" field.
	DefaultWasAutoSubmitted bool
	// DefaultDenoiseAudio holds the default value on creation for the "denoise_audio" field.
	DefaultDenoiseAudio bool
	// DefaultRemoveSilence holds the default value on creation for the "remove_silence" field.
	DefaultRemoveSilence bool
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// Status defines the type for the "status" enum field.
type Status string

// Status values.
const (
	StatusQueued    Status = "queued"
	StatusStarted   Status = "started"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
)

func (s Status) String() string {
	return string(s)
}

// StatusValidator is a validator for the "status" field enum values. It is called by the builders before save.
func StatusValidator(s Status) error {
	switch s {
	case StatusQueued, StatusStarted, StatusSucceeded, StatusFailed:
		return nil
	default:
		return fmt.Errorf("voiceover: invalid enum value for status field: %q", s)
	}
}

const DefaultSourceType enttypes.SourceType = "web-ui"

// SourceTypeValidator is a validator for the "source_type" field enum values. It is called by the builders before save.
func SourceTypeValidator(st enttypes.SourceType) error {
	switch st {
	case "web-ui", "api", "discord", "internal":
		return nil
	default:
		return fmt.Errorf("voiceover: invalid enum value for source_type field: %q", st)
	}
}

// OrderOption defines the ordering options for the Voiceover queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByCountryCode orders the results by the country_code field.
func ByCountryCode(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCountryCode, opts...).ToFunc()
}

// ByStatus orders the results by the status field.
func ByStatus(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStatus, opts...).ToFunc()
}

// ByFailureReason orders the results by the failure_reason field.
func ByFailureReason(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldFailureReason, opts...).ToFunc()
}

// ByStripeProductID orders the results by the stripe_product_id field.
func ByStripeProductID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStripeProductID, opts...).ToFunc()
}

// ByTemperature orders the results by the temperature field.
func ByTemperature(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldTemperature, opts...).ToFunc()
}

// BySeed orders the results by the seed field.
func BySeed(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSeed, opts...).ToFunc()
}

// ByWasAutoSubmitted orders the results by the was_auto_submitted field.
func ByWasAutoSubmitted(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldWasAutoSubmitted, opts...).ToFunc()
}

// ByDenoiseAudio orders the results by the denoise_audio field.
func ByDenoiseAudio(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDenoiseAudio, opts...).ToFunc()
}

// ByRemoveSilence orders the results by the remove_silence field.
func ByRemoveSilence(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldRemoveSilence, opts...).ToFunc()
}

// ByCost orders the results by the cost field.
func ByCost(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCost, opts...).ToFunc()
}

// BySourceType orders the results by the source_type field.
func BySourceType(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSourceType, opts...).ToFunc()
}

// ByPromptID orders the results by the prompt_id field.
func ByPromptID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldPromptID, opts...).ToFunc()
}

// ByUserID orders the results by the user_id field.
func ByUserID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUserID, opts...).ToFunc()
}

// ByDeviceInfoID orders the results by the device_info_id field.
func ByDeviceInfoID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDeviceInfoID, opts...).ToFunc()
}

// ByModelID orders the results by the model_id field.
func ByModelID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldModelID, opts...).ToFunc()
}

// BySpeakerID orders the results by the speaker_id field.
func BySpeakerID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldSpeakerID, opts...).ToFunc()
}

// ByAPITokenID orders the results by the api_token_id field.
func ByAPITokenID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldAPITokenID, opts...).ToFunc()
}

// ByStartedAt orders the results by the started_at field.
func ByStartedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldStartedAt, opts...).ToFunc()
}

// ByCompletedAt orders the results by the completed_at field.
func ByCompletedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCompletedAt, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByUserField orders the results by user field.
func ByUserField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newUserStep(), sql.OrderByField(field, opts...))
	}
}

// ByPromptField orders the results by prompt field.
func ByPromptField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newPromptStep(), sql.OrderByField(field, opts...))
	}
}

// ByDeviceInfoField orders the results by device_info field.
func ByDeviceInfoField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newDeviceInfoStep(), sql.OrderByField(field, opts...))
	}
}

// ByVoiceoverModelsField orders the results by voiceover_models field.
func ByVoiceoverModelsField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newVoiceoverModelsStep(), sql.OrderByField(field, opts...))
	}
}

// ByVoiceoverSpeakersField orders the results by voiceover_speakers field.
func ByVoiceoverSpeakersField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newVoiceoverSpeakersStep(), sql.OrderByField(field, opts...))
	}
}

// ByAPITokensField orders the results by api_tokens field.
func ByAPITokensField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newAPITokensStep(), sql.OrderByField(field, opts...))
	}
}

// ByVoiceoverOutputsCount orders the results by voiceover_outputs count.
func ByVoiceoverOutputsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newVoiceoverOutputsStep(), opts...)
	}
}

// ByVoiceoverOutputs orders the results by voiceover_outputs terms.
func ByVoiceoverOutputs(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newVoiceoverOutputsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}
func newUserStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(UserInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, UserTable, UserColumn),
	)
}
func newPromptStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(PromptInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, PromptTable, PromptColumn),
	)
}
func newDeviceInfoStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(DeviceInfoInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, DeviceInfoTable, DeviceInfoColumn),
	)
}
func newVoiceoverModelsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(VoiceoverModelsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, VoiceoverModelsTable, VoiceoverModelsColumn),
	)
}
func newVoiceoverSpeakersStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(VoiceoverSpeakersInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, VoiceoverSpeakersTable, VoiceoverSpeakersColumn),
	)
}
func newAPITokensStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(APITokensInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, APITokensTable, APITokensColumn),
	)
}
func newVoiceoverOutputsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(VoiceoverOutputsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, VoiceoverOutputsTable, VoiceoverOutputsColumn),
	)
}
