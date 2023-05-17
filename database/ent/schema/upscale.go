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

// Upscale holds the schema definition for the Upscale entity.
type Upscale struct {
	ent.Schema
}

// Fields of the Upscale.
func (Upscale) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Int32("width"),
		field.Int32("height"),
		field.Int32("scale"),
		field.Text("country_code").Optional().Nillable(),
		field.Enum("status").Values("queued", "started", "succeeded", "failed"),
		field.Text("failure_reason").Optional().Nillable(),
		field.Text("stripe_product_id").Optional().Nillable(),
		field.Bool("system_generated").Default(false),
		// ! Relationships / many-to-one
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("device_info_id", uuid.UUID{}),
		field.UUID("model_id", uuid.UUID{}),
		field.UUID("api_token_id", uuid.UUID{}).Optional().Nillable(),
		// ! End relationships
		field.Time("started_at").Optional().Nillable(),
		field.Time("completed_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Upscale.
func (Upscale) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O with users
		edge.From("user", User.Type).
			Ref("upscales").
			Field("user_id").
			Required().
			Unique(),
		// M2O with device_info
		edge.From("device_info", DeviceInfo.Type).
			Ref("upscales").
			Field("device_info_id").
			Required().
			Unique(),
		// M2O with upscale_models
		edge.From("upscale_models", UpscaleModel.Type).
			Ref("upscales").
			Field("model_id").
			Required().
			Unique(),
		// M2O with api_tokens
		edge.From("api_tokens", ApiToken.Type).
			Ref("upscales").
			Field("api_token_id").
			Unique(),
		// O2M with upscale_outputs
		edge.To("upscale_outputs", UpscaleOutput.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

// Annotations of the Upscale.
func (Upscale) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "upscales"},
	}
}
