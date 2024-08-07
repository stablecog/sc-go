// Code generated by ent, DO NOT EDIT.

package negativeprompt

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldLTE(FieldID, id))
}

// Text applies equality check predicate on the "text" field. It's identical to TextEQ.
func Text(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEQ(FieldText, v))
}

// TranslatedText applies equality check predicate on the "translated_text" field. It's identical to TranslatedTextEQ.
func TranslatedText(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEQ(FieldTranslatedText, v))
}

// RanTranslation applies equality check predicate on the "ran_translation" field. It's identical to RanTranslationEQ.
func RanTranslation(v bool) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEQ(FieldRanTranslation, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEQ(FieldUpdatedAt, v))
}

// TextEQ applies the EQ predicate on the "text" field.
func TextEQ(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEQ(FieldText, v))
}

// TextNEQ applies the NEQ predicate on the "text" field.
func TextNEQ(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldNEQ(FieldText, v))
}

// TextIn applies the In predicate on the "text" field.
func TextIn(vs ...string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldIn(FieldText, vs...))
}

// TextNotIn applies the NotIn predicate on the "text" field.
func TextNotIn(vs ...string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldNotIn(FieldText, vs...))
}

// TextGT applies the GT predicate on the "text" field.
func TextGT(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldGT(FieldText, v))
}

// TextGTE applies the GTE predicate on the "text" field.
func TextGTE(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldGTE(FieldText, v))
}

// TextLT applies the LT predicate on the "text" field.
func TextLT(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldLT(FieldText, v))
}

// TextLTE applies the LTE predicate on the "text" field.
func TextLTE(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldLTE(FieldText, v))
}

// TextContains applies the Contains predicate on the "text" field.
func TextContains(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldContains(FieldText, v))
}

// TextHasPrefix applies the HasPrefix predicate on the "text" field.
func TextHasPrefix(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldHasPrefix(FieldText, v))
}

// TextHasSuffix applies the HasSuffix predicate on the "text" field.
func TextHasSuffix(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldHasSuffix(FieldText, v))
}

// TextEqualFold applies the EqualFold predicate on the "text" field.
func TextEqualFold(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEqualFold(FieldText, v))
}

// TextContainsFold applies the ContainsFold predicate on the "text" field.
func TextContainsFold(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldContainsFold(FieldText, v))
}

// TranslatedTextEQ applies the EQ predicate on the "translated_text" field.
func TranslatedTextEQ(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEQ(FieldTranslatedText, v))
}

// TranslatedTextNEQ applies the NEQ predicate on the "translated_text" field.
func TranslatedTextNEQ(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldNEQ(FieldTranslatedText, v))
}

// TranslatedTextIn applies the In predicate on the "translated_text" field.
func TranslatedTextIn(vs ...string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldIn(FieldTranslatedText, vs...))
}

// TranslatedTextNotIn applies the NotIn predicate on the "translated_text" field.
func TranslatedTextNotIn(vs ...string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldNotIn(FieldTranslatedText, vs...))
}

// TranslatedTextGT applies the GT predicate on the "translated_text" field.
func TranslatedTextGT(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldGT(FieldTranslatedText, v))
}

// TranslatedTextGTE applies the GTE predicate on the "translated_text" field.
func TranslatedTextGTE(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldGTE(FieldTranslatedText, v))
}

// TranslatedTextLT applies the LT predicate on the "translated_text" field.
func TranslatedTextLT(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldLT(FieldTranslatedText, v))
}

// TranslatedTextLTE applies the LTE predicate on the "translated_text" field.
func TranslatedTextLTE(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldLTE(FieldTranslatedText, v))
}

// TranslatedTextContains applies the Contains predicate on the "translated_text" field.
func TranslatedTextContains(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldContains(FieldTranslatedText, v))
}

// TranslatedTextHasPrefix applies the HasPrefix predicate on the "translated_text" field.
func TranslatedTextHasPrefix(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldHasPrefix(FieldTranslatedText, v))
}

// TranslatedTextHasSuffix applies the HasSuffix predicate on the "translated_text" field.
func TranslatedTextHasSuffix(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldHasSuffix(FieldTranslatedText, v))
}

// TranslatedTextIsNil applies the IsNil predicate on the "translated_text" field.
func TranslatedTextIsNil() predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldIsNull(FieldTranslatedText))
}

// TranslatedTextNotNil applies the NotNil predicate on the "translated_text" field.
func TranslatedTextNotNil() predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldNotNull(FieldTranslatedText))
}

// TranslatedTextEqualFold applies the EqualFold predicate on the "translated_text" field.
func TranslatedTextEqualFold(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEqualFold(FieldTranslatedText, v))
}

// TranslatedTextContainsFold applies the ContainsFold predicate on the "translated_text" field.
func TranslatedTextContainsFold(v string) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldContainsFold(FieldTranslatedText, v))
}

// RanTranslationEQ applies the EQ predicate on the "ran_translation" field.
func RanTranslationEQ(v bool) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEQ(FieldRanTranslation, v))
}

// RanTranslationNEQ applies the NEQ predicate on the "ran_translation" field.
func RanTranslationNEQ(v bool) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldNEQ(FieldRanTranslation, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldLTE(FieldCreatedAt, v))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v time.Time) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.FieldLTE(FieldUpdatedAt, v))
}

// HasGenerations applies the HasEdge predicate on the "generations" edge.
func HasGenerations() predicate.NegativePrompt {
	return predicate.NegativePrompt(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, GenerationsTable, GenerationsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasGenerationsWith applies the HasEdge predicate on the "generations" edge with a given conditions (other predicates).
func HasGenerationsWith(preds ...predicate.Generation) predicate.NegativePrompt {
	return predicate.NegativePrompt(func(s *sql.Selector) {
		step := newGenerationsStep()
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.NegativePrompt) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.AndPredicates(predicates...))
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.NegativePrompt) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.OrPredicates(predicates...))
}

// Not applies the not operator on the given predicate.
func Not(p predicate.NegativePrompt) predicate.NegativePrompt {
	return predicate.NegativePrompt(sql.NotPredicates(p))
}
