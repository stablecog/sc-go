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

// GenerationModel holds the schema definition for the GenerationModel entity.
type GenerationModel struct {
	ent.Schema
}

// Fields of the GenerationModel.
func (GenerationModel) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("name_in_worker"),
		field.Bool("is_active").Default(true),
		field.Bool("is_default").Default(false),
		field.Bool("is_hidden").Default(false),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Model.
func (GenerationModel) Edges() []ent.Edge {
	return []ent.Edge{
		// O2M with generation
		edge.To("generations", Generation.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

// Annotations of the GenerationModel.
func (GenerationModel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "generation_models"},
	}
}
