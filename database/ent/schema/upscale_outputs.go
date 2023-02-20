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

// UpscaleOutput holds the schema definition for the UpscaleOutput entity.
type UpscaleOutput struct {
	ent.Schema
}

// Fields of the UpscaleOutput.
func (UpscaleOutput) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("image_path"),
		field.Text("input_image_url").Optional().Nillable(),
		// ! Relationships / many-to-one
		field.UUID("upscale_id", uuid.UUID{}),
		// ! one-to-one
		field.UUID("generation_output_id", uuid.UUID{}).Optional().Nillable(),
		// ! End relationships
		field.Time("deleted_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the UpscaleOutput.
func (UpscaleOutput) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O with upscales
		edge.From("upscales", Upscale.Type).
			Ref("upscale_outputs").
			Field("upscale_id").
			Required().
			Unique(),
		// O2O with generation_outputs
		edge.From("generation_output", GenerationOutput.Type).
			Ref("upscale_outputs").
			Field("generation_output_id").
			Unique(),
	}
}

// Annotations of the UpscaleOutput.
func (UpscaleOutput) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "upscale_outputs"},
	}
}

// Indexes of the UpscaleOutput.
func (UpscaleOutput) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("image_path"),
	}
}
