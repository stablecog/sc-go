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

// CreditType holds the schema definition for the CreditType entity.
type CreditType struct {
	ent.Schema
}

// Fields of the CreditType.
func (CreditType) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("name").Unique(),
		field.Text("description").Optional().Nillable(),
		field.Int32("amount"),
		field.Text("stripe_product_id").Optional().Nillable(),
		field.Enum("type").Values("free", "subscription", "one_time", "tippable"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the CreditType.
func (CreditType) Edges() []ent.Edge {
	return []ent.Edge{
		// O2M with credits
		edge.To("credits", Credit.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

// Annotations of the CreditType.
func (CreditType) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "credit_types"},
	}
}
