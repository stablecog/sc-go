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

// ApiToken holds the schema definition for the ApiToken entity.
type ApiToken struct {
	ent.Schema
}

// Fields of the ApiToken.
func (ApiToken) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("hashed_token"),
		field.Text("name"),
		field.Text("short_string"),
		field.Bool("is_active").Default(true),
		field.Int("uses").Default(0),
		field.Int("credits_spent").Default(0),
		// ! Relationships
		field.UUID("user_id", uuid.UUID{}),
		// ! End Relationships
		field.Time("last_used_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the ApiToken.
func (ApiToken) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O with users
		edge.From("user", User.Type).
			Ref("api_tokens").
			Field("user_id").
			Required().
			Unique(),
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
	}
}

// Annotations of the ApiToken.
func (ApiToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "api_tokens"},
	}
}
