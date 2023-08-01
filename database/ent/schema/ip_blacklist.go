package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// IPBlackList holds the schema definition for the IPBlackList entity.
type IPBlackList struct {
	ent.Schema
}

// Fields of the IPBlackList.
func (IPBlackList) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("ip").Unique(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Model.
func (IPBlackList) Edges() []ent.Edge {
	return nil
}

// Annotations of the IPBlackList.
func (IPBlackList) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "ip_blacklist"},
	}
}
