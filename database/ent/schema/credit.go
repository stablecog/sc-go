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

// Credit holds the schema definition for the Credit entity.
type Credit struct {
	ent.Schema
}

// Fields of the Credit.
func (Credit) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Int32("remaining_amount"),
		field.Time("expires_at"),
		field.String("stripe_line_item_id").Optional().Unique().Nillable(),
		field.Time("replenished_at").Default(time.Now),
		// ! Relationships / many-to-one
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("credit_type_id", uuid.UUID{}),
		// ! End relationships
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Credit.
func (Credit) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O with users
		edge.From("users", User.Type).
			Ref("credits").
			Field("user_id").
			Required().
			Unique(),
		// M2O with users
		edge.From("credit_type", CreditType.Type).
			Ref("credits").
			Field("credit_type_id").
			Required().
			Unique(),
	}
}

// Annotations of the Credit.
func (Credit) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "credits"},
	}
}

// Indexes of the Credit.
func (Credit) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("expires_at", "user_id", "remaining_amount"),
	}
}
