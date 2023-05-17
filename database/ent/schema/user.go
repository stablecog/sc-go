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

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("email"),
		field.Text("stripe_customer_id").Unique(),
		field.Text("active_product_id").Optional().Nillable(),
		field.Time("last_sign_in_at").Optional().Nillable(),
		field.Time("last_seen_at").Default(time.Now),
		field.Time("banned_at").Optional().Nillable(),
		field.Time("scheduled_for_deletion_on").Optional().Nillable(),
		field.Time("data_deleted_at").Optional().Nillable(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		// O2M with user_roles
		edge.To("user_roles", UserRole.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
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
		// O2M with credits
		edge.To("credits", Credit.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
		// O2M with api_tokens
		edge.To("api_tokens", ApiToken.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

// Annotations of the User.
func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "users"},
	}
}
