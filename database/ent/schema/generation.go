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
		field.Int32("num_interference_steps"),
		field.Float32("guidance_scale"),
		field.Int("seed"),
		field.Enum("status").Values("queued", "started", "succeeded", "failed", "rejected"),
		field.Text("failure_reason").Optional().Nillable(),
		field.Text("country_code"),
		field.Enum("gallery_status").Values("not_submitted", "submitted", "accepted", "rejected").Default("not_submitted"),
		field.Text("init_image_url").Optional().Nillable(),
		// ! Relationships / many-to-one
		field.UUID("prompt_id", uuid.UUID{}),
		field.UUID("negative_prompt_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("model_id", uuid.UUID{}),
		field.UUID("scheduler_id", uuid.UUID{}),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("device_info_id", uuid.UUID{}),
		// ! End relationships
		field.Time("started_at").Optional().Nillable(),
		field.Time("completed_at").Optional().Nillable(),
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
		edge.From("scheduler", Scheduler.Type).
			Ref("generations").
			Field("scheduler_id").
			Required().
			Unique(),
		// M2O with prompt
		edge.From("prompt", Prompt.Type).
			Ref("generations").
			Field("prompt_id").
			Required().
			Unique(),
		// M2O with negative_prompts
		edge.From("negative_prompt", NegativePrompt.Type).
			Ref("generations").
			Field("negative_prompt_id").
			Unique(),
		// M2O with generation_models
		edge.From("generation_model", GenerationModel.Type).
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
