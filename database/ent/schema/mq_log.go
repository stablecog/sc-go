package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// MqLog holds the schema definition for the MqLog entity.
type MqLog struct {
	ent.Schema
}

// Fields of the MqLog.
func (MqLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Text("message_id").Unique(),
		field.Int("priority"),
		field.Bool("is_processing").Default(false),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Model.
func (MqLog) Edges() []ent.Edge {
	return nil
}

// Annotations of the MqLog.
func (MqLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "mq_log"},
	}
}
