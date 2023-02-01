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

// Subscription holds the schema definition for the Subscription entity.
type Subscription struct {
	ent.Schema
}

// Fields of the Subscription.
func (Subscription) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("subscription_tier_id", uuid.UUID{}),
		field.Time("paid_started_at").Optional().Nillable(),
		field.Time("paid_cancelled_at").Optional().Nillable(),
		field.Time("paid_expires_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Model.
func (Subscription) Edges() []ent.Edge {
	return []ent.Edge{
		// O2O with users
		edge.From("user", User.Type).
			Ref("subscriptions").
			Field("user_id").
			Unique().
			Required(),
		// M2O with subscription_tiers
		edge.From("subscription_tier", SubscriptionTier.Type).
			Ref("subscriptions").
			Field("subscription_tier_id").
			Required().
			Unique(),
	}
}

// Annotations of the Subscription.
func (Subscription) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "subscriptions"},
	}
}
