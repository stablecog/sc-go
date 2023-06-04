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

// DeviceInfo holds the schema definition for the DeviceInfo entity.
type DeviceInfo struct {
	ent.Schema
}

// Fields of the DeviceInfo.
func (DeviceInfo) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("type").Optional().Nillable(),
		field.Text("os").Optional().Nillable(),
		field.Text("browser").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the DeviceInfo.
func (DeviceInfo) Edges() []ent.Edge {
	return []ent.Edge{
		// O2M with generations
		edge.To("generations", Generation.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		// O2M with upscales
		edge.To("upscales", Upscale.Type).
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

// Annotations of the DeviceInfo.
func (DeviceInfo) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "device_info"},
	}
}
