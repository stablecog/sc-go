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

// GenerationOutput holds the schema definition for the GenerationOutput entity.
type GenerationOutput struct {
	ent.Schema
}

// Fields of the GenerationOutput.
func (GenerationOutput) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("image_url"),
		field.Text("upscaled_image_url").Optional().Nillable(),
		field.Enum("gallery_status").Values("not_submitted", "submitted", "accepted", "rejected").Default("not_submitted"),
		// ! Relationships / many-to-one
		field.UUID("generation_id", uuid.UUID{}),
		// ! End relationships
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the GenerationOutput.
func (GenerationOutput) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O with generations
		edge.From("generations", Generation.Type).
			Ref("generation_outputs").
			Field("generation_id").
			Required().
			Unique(),
	}
}

// Annotations of the GenerationOutput.
func (GenerationOutput) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "generation_outputs"},
	}
}
