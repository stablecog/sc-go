// Code generated by ent, DO NOT EDIT.

package voiceover

import (
	"fmt"
	"time"

	"github.com/google/uuid"
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
