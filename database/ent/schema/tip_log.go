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

// TipLog holds the schema definition for the TipLog entity.
type TipLog struct {
	ent.Schema
}

// Fields of the TipLog.
func (TipLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.Int32("amount"),
		field.Text("tipped_to_discord_id"),
		// ! Relationships / many-to-one
		field.UUID("tipped_by", uuid.UUID{}),
		field.UUID("tipped_to", uuid.UUID{}).Optional().Nillable(),
		// ! End relationships
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the TipLog.
func (TipLog) Edges() []ent.Edge {
	return []ent.Edge{
		// M2O with users
		edge.From("tips_received", User.Type).
			Ref("tips_received").
			Field("tipped_to").
			Unique(),
		// M2O with users
		edge.From("tips_given", User.Type).
			Ref("tips_given").
			Field("tipped_by").
			Required().
			Unique(),
	}
}

// Annotations of the TipLog.
func (TipLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "tip_log"},
	}
}
