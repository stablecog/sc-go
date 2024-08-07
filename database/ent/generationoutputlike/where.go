// Code generated by ent, DO NOT EDIT.

package generationoutputlike

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldLTE(FieldID, id))
}

// OutputID applies equality check predicate on the "output_id" field. It's identical to OutputIDEQ.
func OutputID(v uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldEQ(FieldOutputID, v))
}

// LikedByUserID applies equality check predicate on the "liked_by_user_id" field. It's identical to LikedByUserIDEQ.
func LikedByUserID(v uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldEQ(FieldLikedByUserID, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldEQ(FieldCreatedAt, v))
}

// OutputIDEQ applies the EQ predicate on the "output_id" field.
func OutputIDEQ(v uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldEQ(FieldOutputID, v))
}

// OutputIDNEQ applies the NEQ predicate on the "output_id" field.
func OutputIDNEQ(v uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldNEQ(FieldOutputID, v))
}

// OutputIDIn applies the In predicate on the "output_id" field.
func OutputIDIn(vs ...uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldIn(FieldOutputID, vs...))
}

// OutputIDNotIn applies the NotIn predicate on the "output_id" field.
func OutputIDNotIn(vs ...uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldNotIn(FieldOutputID, vs...))
}

// LikedByUserIDEQ applies the EQ predicate on the "liked_by_user_id" field.
func LikedByUserIDEQ(v uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldEQ(FieldLikedByUserID, v))
}

// LikedByUserIDNEQ applies the NEQ predicate on the "liked_by_user_id" field.
func LikedByUserIDNEQ(v uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldNEQ(FieldLikedByUserID, v))
}

// LikedByUserIDIn applies the In predicate on the "liked_by_user_id" field.
func LikedByUserIDIn(vs ...uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldIn(FieldLikedByUserID, vs...))
}

// LikedByUserIDNotIn applies the NotIn predicate on the "liked_by_user_id" field.
func LikedByUserIDNotIn(vs ...uuid.UUID) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldNotIn(FieldLikedByUserID, vs...))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.FieldLTE(FieldCreatedAt, v))
}

// HasGenerationOutputs applies the HasEdge predicate on the "generation_outputs" edge.
func HasGenerationOutputs() predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, GenerationOutputsTable, GenerationOutputsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasGenerationOutputsWith applies the HasEdge predicate on the "generation_outputs" edge with a given conditions (other predicates).
func HasGenerationOutputsWith(preds ...predicate.GenerationOutput) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(func(s *sql.Selector) {
		step := newGenerationOutputsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasUsers applies the HasEdge predicate on the "users" edge.
func HasUsers() predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2O, true, UsersTable, UsersColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasUsersWith applies the HasEdge predicate on the "users" edge with a given conditions (other predicates).
func HasUsersWith(preds ...predicate.User) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(func(s *sql.Selector) {
		step := newUsersStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.GenerationOutputLike) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.GenerationOutputLike) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.GenerationOutputLike) predicate.GenerationOutputLike {
	return predicate.GenerationOutputLike(sql.NotPredicates(p))
}
