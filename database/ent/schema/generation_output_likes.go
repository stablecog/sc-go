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

// GenerationOutputLike holds the schema definition for the GenerationOutputLike entity.
type GenerationOutputLike struct {
	ent.Schema
}

// Fields of the GenerationOutputLike.
func (GenerationOutputLike) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		// ! Relationships / many-to-one
		field.UUID("output_id", uuid.UUID{}),
		field.UUID("liked_by_user_id", uuid.UUID{}), // Liked
		// ! End relationships
		field.Time("created_at").Default(time.Now).Immutable(),
	}
}

// Edges of the GenerationOutputLike.
func (GenerationOutputLike) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O with generation_outputs
		edge.From("generation_outputs", GenerationOutput.Type).
			Ref("generation_output_likes").
			Field("output_id").
			Required().
			Unique(),
		// M2O with users
		edge.From("users", User.Type).
			Ref("generation_output_likes").
			Field("liked_by_user_id").
			Required().
			Unique(),
	}
}

// Annotations of the GenerationOutputLike.
func (GenerationOutputLike) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "generation_output_likes"},
	}
}

// Indexes of the GenerationOutputLike.
func (GenerationOutputLike) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("output_id", "liked_by_user_id").Unique(),
	}
}
