package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

// GenerationOutputEmbed holds the schema definition for the GenerationOutputEmbed entity.
type GenerationOutputEmbed struct {
	ent.Schema
}

// Fields of the GenerationOutputEmbed.
func (GenerationOutputEmbed) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Other("prompt_embedding", pgvector.Vector{}).
			SchemaType(map[string]string{
				dialect.Postgres: "vector(1024)",
			}),
		field.Other("image_embedding", pgvector.Vector{}).
			SchemaType(map[string]string{
				dialect.Postgres: "vector(1024)",
			}),
		// ! Relationships / many-to-one
		field.UUID("output_id", uuid.UUID{}),
		// ! End relationships
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the GenerationOutputEmbed.
func (GenerationOutputEmbed) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O with generation_outputs
		edge.From("generation_outputs", GenerationOutput.Type).
			Ref("generation_output_embeds").
			Field("output_id").
			Required().
			Unique(),
	}
}

// Annotations of the GenerationOutputEmbed.
func (GenerationOutputEmbed) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "generation_output_embeds"},
	}
}

// Indexes of the GenerationOutputEmbed.
func (GenerationOutputEmbed) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("output_id").Unique(),
		index.Fields("prompt_embedding").
			Annotations(
				entsql.IndexType("hnsw"),
				entsql.OpClass("vector_l2_ops"),
			),
		index.Fields("image_embedding").
			Annotations(
				entsql.IndexType("hnsw"),
				entsql.OpClass("vector_l2_ops"),
			),
	}
}
