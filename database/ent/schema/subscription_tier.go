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

// SubscriptionTier holds the schema definition for the SubscriptionTier entity.
type SubscriptionTier struct {
	ent.Schema
}

// Fields of the SubscriptionTier.
func (SubscriptionTier) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.String("name").Unique(),
		field.Int32("base_credits"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Model.
func (SubscriptionTier) Edges() []ent.Edge {
	return []ent.Edge{
		// O2M with subscriptions
		edge.To("subscriptions", Subscription.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

// Annotations of the SubscriptionTier.
func (SubscriptionTier) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "subscription_tiers"},
	}
}
