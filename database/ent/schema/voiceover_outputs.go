package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// VoiceoverOutput holds the schema definition for the VoiceoverOutput entity.
type VoiceoverOutput struct {
	ent.Schema
}

// Fields of the VoiceoverOutput.
func (VoiceoverOutput) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("audio_path"),
		// ! TODO - these 2 optional since not all have them yet
		field.Text("video_path").Optional().Nillable(),
		field.JSON("audio_array", []float64{}).Optional(),
		field.Bool("is_favorited").Default(false),
		field.Float32("audio_duration"),
		field.Enum("gallery_status").Values("not_submitted", "submitted", "approved", "rejected").Default("not_submitted"),
		// ! Relationships / many-to-one
		field.UUID("voiceover_id", uuid.UUID{}),
		// ! End relationships
		field.Time("deleted_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the VoiceoverOutput.
func (VoiceoverOutput) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O with voiceovers
		edge.From("voiceovers", Voiceover.Type).
			Ref("voiceover_outputs").
			Field("voiceover_id").
			Required().
			Unique(),
	}
}

// Annotations of the VoiceoverOutput.
func (VoiceoverOutput) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "voiceover_outputs"},
	}
}

// Indexes of the VoiceoverOutput.
func (VoiceoverOutput) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("audio_path"),
	}
}
