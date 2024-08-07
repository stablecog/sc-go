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

// Prompt holds the schema definition for the Prompt entity.
type Prompt struct {
	ent.Schema
}

// Fields of the Prompt.
func (Prompt) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("text"),
		field.Text("translated_text").Optional().Nillable(),
		field.Bool("ran_translation").Default(false),
		field.Enum("type").Values("image", "voiceover"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Prompt.
func (Prompt) Edges() []ent.Edge {
	return []ent.Edge{
		// O2M with generations
		edge.To("generations", Generation.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		// O2M with voiceovers
		edge.To("voiceovers", Voiceover.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

// Annotations of the Prompt.
func (Prompt) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "prompts"},
	}
}
