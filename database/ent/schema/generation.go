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

// Generation holds the schema definition for the Generation entity.
type Generation struct {
	ent.Schema
}

// Fields of the Generation.
func (Generation) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Int32("width"),
		field.Int32("height"),
		field.Int32("interference_steps"),
		field.Float32("guidance_scale"),
		field.Int("seed").Nillable(),
		field.Int32("duration_ms"),
		field.Enum("status").Values("started", "succeeded", "failed", "rejected"),
		field.Text("failure_reason").Nillable(),
		field.Text("country_code"),
		// ! Relationships / many-to-one
		field.UUID("prompt_id", uuid.UUID{}),
		field.UUID("negative_prompt_id", uuid.UUID{}).Nillable(),
		field.UUID("model_id", uuid.UUID{}),
		field.UUID("scheduler_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("device_info_id", uuid.UUID{}),
		// ! End relationships
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Generation.
func (Generation) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O with device_info
		edge.From("device_info", DeviceInfo.Type).
			Ref("generations").
			Field("device_info_id").
			Required().
			Unique(),
		// M2O with schedulers
		edge.From("schedulers", Scheduler.Type).
			Ref("generations").
			Field("scheduler_id").
			Required().
			Unique(),
		// M2O with prompts
		edge.From("prompts", Prompt.Type).
			Ref("generations").
			Field("prompt_id").
			Required().
			Unique(),
		// M2O with negative_prompts
		edge.From("negative_prompts", NegativePrompt.Type).
			Ref("generations").
			Field("negative_prompt_id").
			Required().
			Unique(),
		// M2O with generation_models
		edge.From("generation_models", GenerationModel.Type).
			Ref("generations").
			Field("model_id").
			Required().
			Unique(),
		// M2O with users
		edge.From("users", User.Type).
			Ref("generations").
			Field("user_id").
			Required().
			Unique(),
		// O2M with generation_outputs
		edge.To("generation_outputs", GenerationOutput.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

// Annotations of the Generation.
func (Generation) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "generations"},
	}
}
