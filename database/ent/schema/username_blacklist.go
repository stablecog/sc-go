package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// UsernameBlacklist holds the schema definition for the UsernameBlacklist entity.
type UsernameBlacklist struct {
	ent.Schema
}

// Fields of the UsernameBlacklist.
func (UsernameBlacklist) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("username").Unique(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Model.
func (UsernameBlacklist) Edges() []ent.Edge {
	return nil
}

// Annotations of the UsernameBlacklist.
func (UsernameBlacklist) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "username_blacklist"},
	}
}
