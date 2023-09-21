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

// AuthClient holds the schema definition for the AuthClient entity.
type AuthClient struct {
	ent.Schema
}

// Fields of the AuthClient.
func (AuthClient) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("name"),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the AuthClient.
func (AuthClient) Edges() []ent.Edge {
	return []ent.Edge{
		// O2M with api_tokens
		edge.To("api_tokens", ApiToken.Type).
			Annotations(entsql.Annotation{
				OnDelete: entsql.Cascade,
			}),
	}
}

// Annotations of the AuthClient.
func (AuthClient) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "auth_clients"},
	}
}
