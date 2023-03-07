package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// DisposableEmail holds the schema definition for the DisposableEmail entity.
type DisposableEmail struct {
	ent.Schema
}

// Fields of the DisposableEmail.
func (DisposableEmail) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("domain").Unique(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Model.
func (DisposableEmail) Edges() []ent.Edge {
	return nil
}

// Annotations of the DisposableEmail.
func (DisposableEmail) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "disposable_emails"},
	}
}
