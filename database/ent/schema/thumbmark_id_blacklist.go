package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// ThumbmarkIdBlackList holds the schema definition for the ThumbmarkIdBlackList entity.
type ThumbmarkIdBlackList struct {
	ent.Schema
}

// Fields of the ThumbmarkIdBlackList.
func (ThumbmarkIdBlackList) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("thumbmark_id").Unique(),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Model.
func (ThumbmarkIdBlackList) Edges() []ent.Edge {
	return nil
}

// Annotations of the ThumbmarkIdBlackList.
func (ThumbmarkIdBlackList) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "thumbmark_id_blacklist"},
	}
}
