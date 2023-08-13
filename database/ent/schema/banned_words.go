package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// BannedWords holds the schema definition for the BannedWords entity.
type BannedWords struct {
	ent.Schema
}

// Fields of the BannedWords.
func (BannedWords) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New),
		field.JSON("words", []string{}),
		field.Text("reason"),
		field.Bool("split_match").Default(false),
		field.Time("created_at").Default(time.Now).Immutable(),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Model.
func (BannedWords) Edges() []ent.Edge {
	return nil
}

// Annotations of the BannedWords.
func (BannedWords) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "banned_words"},
	}
}
