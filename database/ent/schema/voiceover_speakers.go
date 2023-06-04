package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// VoiceoverSpeaker holds the schema definition for the VoiceoverSpeaker entity.
type VoiceoverSpeaker struct {
	ent.Schema
}

// Fields of the VoiceoverSpeaker.
func (VoiceoverSpeaker) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("name_in_worker"),
		field.Bool("is_active").Default(true),
		field.Bool("is_default").Default(false),
		field.Bool("is_hidden").Default(false),
		// ! Relationships / many-to-one
		field.UUID("model_id", uuid.UUID{}),
		// ! End relationships
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Model.
func (VoiceoverSpeaker) Edges() []ent.Edge {
	return []ent.Edge{
		// O2M with voiceovers
		edge.To("voiceovers", Voiceover.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		// M2O with Voiceover_models
		edge.From("voiceover_models", VoiceoverModel.Type).
			Ref("voiceover_speakers").
			Field("model_id").
			Required().
			Unique(),
	}
}

// Annotations of the VoiceoverSpeaker.
func (VoiceoverSpeaker) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "voiceover_speakers"},
	}
}
