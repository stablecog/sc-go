package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/enttypes"
)

// Voiceover holds the schema definition for the Voiceover entity.
type Voiceover struct {
	ent.Schema
}

// Fields of the Voiceover.
func (Voiceover) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("country_code").Optional().Nillable(),
		field.Enum("status").Values("queued", "started", "succeeded", "failed"),
		field.Text("failure_reason").Optional().Nillable(),
		field.Text("stripe_product_id").Optional().Nillable(),
		field.Float32("temperature"),
		field.Int("seed"),
		field.Bool("was_auto_submitted").Default(false),
		field.Bool("denoise_audio").Default(true),
		field.Bool("remove_silence").Default(true),
		field.Int32("cost"),
		field.Enum("source_type").GoType(enttypes.SourceType("")).Default(string(enttypes.SourceTypeWebUI)),
		// ! Relationships / many-to-one
		field.UUID("prompt_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("device_info_id", uuid.UUID{}),
		field.UUID("model_id", uuid.UUID{}),
		field.UUID("speaker_id", uuid.UUID{}),
		field.UUID("api_token_id", uuid.UUID{}).Optional().Nillable(),
		// ! End relationships
		field.Time("started_at").Optional().Nillable(),
		field.Time("completed_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Voiceover.
func (Voiceover) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O with users
		edge.From("user", User.Type).
			Ref("voiceovers").
			Field("user_id").
			Required().
			Unique(),
		// M2O with prompt
		edge.From("prompt", Prompt.Type).
			Ref("voiceovers").
			Field("prompt_id").
			Unique(),
		// M2O with device_info
		edge.From("device_info", DeviceInfo.Type).
			Ref("voiceovers").
			Field("device_info_id").
			Required().
			Unique(),
		// M2O with Voiceover_models
		edge.From("voiceover_models", VoiceoverModel.Type).
			Ref("voiceovers").
			Field("model_id").
			Required().
			Unique(),
		// M2O with Voiceover_speakers
		edge.From("voiceover_speakers", VoiceoverSpeaker.Type).
			Ref("voiceovers").
			Field("speaker_id").
			Required().
			Unique(),
		// M2O with api_tokens
		edge.From("api_tokens", ApiToken.Type).
			Ref("voiceovers").
			Field("api_token_id").
			Unique(),
		// O2M with Voiceover_outputs
		edge.To("voiceover_outputs", VoiceoverOutput.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

// Annotations of the Voiceover.
func (Voiceover) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "voiceovers"},
	}
}
