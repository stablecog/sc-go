// Code generated by ent, DO NOT EDIT.

package user

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/predicate"
)

// ID filters vertices based on their ID field.
func ID(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldEQ(FieldID, id))
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldEQ(FieldID, id))
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldID, id))
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...uuid.UUID) predicate.User {
	return predicate.User(sql.FieldIn(FieldID, ids...))
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...uuid.UUID) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldID, ids...))
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldGT(FieldID, id))
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldGTE(FieldID, id))
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldLT(FieldID, id))
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id uuid.UUID) predicate.User {
	return predicate.User(sql.FieldLTE(FieldID, id))
}

// Email applies equality check predicate on the "email" field. It's identical to EmailEQ.
func Email(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldEmail, v))
}

// StripeCustomerID applies equality check predicate on the "stripe_customer_id" field. It's identical to StripeCustomerIDEQ.
func StripeCustomerID(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldStripeCustomerID, v))
}

// ActiveProductID applies equality check predicate on the "active_product_id" field. It's identical to ActiveProductIDEQ.
func ActiveProductID(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldActiveProductID, v))
}

// LastSignInAt applies equality check predicate on the "last_sign_in_at" field. It's identical to LastSignInAtEQ.
func LastSignInAt(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldLastSignInAt, v))
}

// LastSeenAt applies equality check predicate on the "last_seen_at" field. It's identical to LastSeenAtEQ.
func LastSeenAt(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldLastSeenAt, v))
}

// BannedAt applies equality check predicate on the "banned_at" field. It's identical to BannedAtEQ.
func BannedAt(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldBannedAt, v))
}

// ScheduledForDeletionOn applies equality check predicate on the "scheduled_for_deletion_on" field. It's identical to ScheduledForDeletionOnEQ.
func ScheduledForDeletionOn(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldScheduledForDeletionOn, v))
}

// DataDeletedAt applies equality check predicate on the "data_deleted_at" field. It's identical to DataDeletedAtEQ.
func DataDeletedAt(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldDataDeletedAt, v))
}

// WantsEmail applies equality check predicate on the "wants_email" field. It's identical to WantsEmailEQ.
func WantsEmail(v bool) predicate.User {
	return predicate.User(sql.FieldEQ(FieldWantsEmail, v))
}

// DiscordID applies equality check predicate on the "discord_id" field. It's identical to DiscordIDEQ.
func DiscordID(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldDiscordID, v))
}

// Username applies equality check predicate on the "username" field. It's identical to UsernameEQ.
func Username(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldUsername, v))
}

// CreatedAt applies equality check predicate on the "created_at" field. It's identical to CreatedAtEQ.
func CreatedAt(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldCreatedAt, v))
}

// UpdatedAt applies equality check predicate on the "updated_at" field. It's identical to UpdatedAtEQ.
func UpdatedAt(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldUpdatedAt, v))
}

// EmailEQ applies the EQ predicate on the "email" field.
func EmailEQ(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldEmail, v))
}

// EmailNEQ applies the NEQ predicate on the "email" field.
func EmailNEQ(v string) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldEmail, v))
}

// EmailIn applies the In predicate on the "email" field.
func EmailIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldIn(FieldEmail, vs...))
}

// EmailNotIn applies the NotIn predicate on the "email" field.
func EmailNotIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldEmail, vs...))
}

// EmailGT applies the GT predicate on the "email" field.
func EmailGT(v string) predicate.User {
	return predicate.User(sql.FieldGT(FieldEmail, v))
}

// EmailGTE applies the GTE predicate on the "email" field.
func EmailGTE(v string) predicate.User {
	return predicate.User(sql.FieldGTE(FieldEmail, v))
}

// EmailLT applies the LT predicate on the "email" field.
func EmailLT(v string) predicate.User {
	return predicate.User(sql.FieldLT(FieldEmail, v))
}

// EmailLTE applies the LTE predicate on the "email" field.
func EmailLTE(v string) predicate.User {
	return predicate.User(sql.FieldLTE(FieldEmail, v))
}

// EmailContains applies the Contains predicate on the "email" field.
func EmailContains(v string) predicate.User {
	return predicate.User(sql.FieldContains(FieldEmail, v))
}

// EmailHasPrefix applies the HasPrefix predicate on the "email" field.
func EmailHasPrefix(v string) predicate.User {
	return predicate.User(sql.FieldHasPrefix(FieldEmail, v))
}

// EmailHasSuffix applies the HasSuffix predicate on the "email" field.
func EmailHasSuffix(v string) predicate.User {
	return predicate.User(sql.FieldHasSuffix(FieldEmail, v))
}

// EmailEqualFold applies the EqualFold predicate on the "email" field.
func EmailEqualFold(v string) predicate.User {
	return predicate.User(sql.FieldEqualFold(FieldEmail, v))
}

// EmailContainsFold applies the ContainsFold predicate on the "email" field.
func EmailContainsFold(v string) predicate.User {
	return predicate.User(sql.FieldContainsFold(FieldEmail, v))
}

// StripeCustomerIDEQ applies the EQ predicate on the "stripe_customer_id" field.
func StripeCustomerIDEQ(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldStripeCustomerID, v))
}

// StripeCustomerIDNEQ applies the NEQ predicate on the "stripe_customer_id" field.
func StripeCustomerIDNEQ(v string) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldStripeCustomerID, v))
}

// StripeCustomerIDIn applies the In predicate on the "stripe_customer_id" field.
func StripeCustomerIDIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldIn(FieldStripeCustomerID, vs...))
}

// StripeCustomerIDNotIn applies the NotIn predicate on the "stripe_customer_id" field.
func StripeCustomerIDNotIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldStripeCustomerID, vs...))
}

// StripeCustomerIDGT applies the GT predicate on the "stripe_customer_id" field.
func StripeCustomerIDGT(v string) predicate.User {
	return predicate.User(sql.FieldGT(FieldStripeCustomerID, v))
}

// StripeCustomerIDGTE applies the GTE predicate on the "stripe_customer_id" field.
func StripeCustomerIDGTE(v string) predicate.User {
	return predicate.User(sql.FieldGTE(FieldStripeCustomerID, v))
}

// StripeCustomerIDLT applies the LT predicate on the "stripe_customer_id" field.
func StripeCustomerIDLT(v string) predicate.User {
	return predicate.User(sql.FieldLT(FieldStripeCustomerID, v))
}

// StripeCustomerIDLTE applies the LTE predicate on the "stripe_customer_id" field.
func StripeCustomerIDLTE(v string) predicate.User {
	return predicate.User(sql.FieldLTE(FieldStripeCustomerID, v))
}

// StripeCustomerIDContains applies the Contains predicate on the "stripe_customer_id" field.
func StripeCustomerIDContains(v string) predicate.User {
	return predicate.User(sql.FieldContains(FieldStripeCustomerID, v))
}

// StripeCustomerIDHasPrefix applies the HasPrefix predicate on the "stripe_customer_id" field.
func StripeCustomerIDHasPrefix(v string) predicate.User {
	return predicate.User(sql.FieldHasPrefix(FieldStripeCustomerID, v))
}

// StripeCustomerIDHasSuffix applies the HasSuffix predicate on the "stripe_customer_id" field.
func StripeCustomerIDHasSuffix(v string) predicate.User {
	return predicate.User(sql.FieldHasSuffix(FieldStripeCustomerID, v))
}

// StripeCustomerIDEqualFold applies the EqualFold predicate on the "stripe_customer_id" field.
func StripeCustomerIDEqualFold(v string) predicate.User {
	return predicate.User(sql.FieldEqualFold(FieldStripeCustomerID, v))
}

// StripeCustomerIDContainsFold applies the ContainsFold predicate on the "stripe_customer_id" field.
func StripeCustomerIDContainsFold(v string) predicate.User {
	return predicate.User(sql.FieldContainsFold(FieldStripeCustomerID, v))
}

// ActiveProductIDEQ applies the EQ predicate on the "active_product_id" field.
func ActiveProductIDEQ(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldActiveProductID, v))
}

// ActiveProductIDNEQ applies the NEQ predicate on the "active_product_id" field.
func ActiveProductIDNEQ(v string) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldActiveProductID, v))
}

// ActiveProductIDIn applies the In predicate on the "active_product_id" field.
func ActiveProductIDIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldIn(FieldActiveProductID, vs...))
}

// ActiveProductIDNotIn applies the NotIn predicate on the "active_product_id" field.
func ActiveProductIDNotIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldActiveProductID, vs...))
}

// ActiveProductIDGT applies the GT predicate on the "active_product_id" field.
func ActiveProductIDGT(v string) predicate.User {
	return predicate.User(sql.FieldGT(FieldActiveProductID, v))
}

// ActiveProductIDGTE applies the GTE predicate on the "active_product_id" field.
func ActiveProductIDGTE(v string) predicate.User {
	return predicate.User(sql.FieldGTE(FieldActiveProductID, v))
}

// ActiveProductIDLT applies the LT predicate on the "active_product_id" field.
func ActiveProductIDLT(v string) predicate.User {
	return predicate.User(sql.FieldLT(FieldActiveProductID, v))
}

// ActiveProductIDLTE applies the LTE predicate on the "active_product_id" field.
func ActiveProductIDLTE(v string) predicate.User {
	return predicate.User(sql.FieldLTE(FieldActiveProductID, v))
}

// ActiveProductIDContains applies the Contains predicate on the "active_product_id" field.
func ActiveProductIDContains(v string) predicate.User {
	return predicate.User(sql.FieldContains(FieldActiveProductID, v))
}

// ActiveProductIDHasPrefix applies the HasPrefix predicate on the "active_product_id" field.
func ActiveProductIDHasPrefix(v string) predicate.User {
	return predicate.User(sql.FieldHasPrefix(FieldActiveProductID, v))
}

// ActiveProductIDHasSuffix applies the HasSuffix predicate on the "active_product_id" field.
func ActiveProductIDHasSuffix(v string) predicate.User {
	return predicate.User(sql.FieldHasSuffix(FieldActiveProductID, v))
}

// ActiveProductIDIsNil applies the IsNil predicate on the "active_product_id" field.
func ActiveProductIDIsNil() predicate.User {
	return predicate.User(sql.FieldIsNull(FieldActiveProductID))
}

// ActiveProductIDNotNil applies the NotNil predicate on the "active_product_id" field.
func ActiveProductIDNotNil() predicate.User {
	return predicate.User(sql.FieldNotNull(FieldActiveProductID))
}

// ActiveProductIDEqualFold applies the EqualFold predicate on the "active_product_id" field.
func ActiveProductIDEqualFold(v string) predicate.User {
	return predicate.User(sql.FieldEqualFold(FieldActiveProductID, v))
}

// ActiveProductIDContainsFold applies the ContainsFold predicate on the "active_product_id" field.
func ActiveProductIDContainsFold(v string) predicate.User {
	return predicate.User(sql.FieldContainsFold(FieldActiveProductID, v))
}

// LastSignInAtEQ applies the EQ predicate on the "last_sign_in_at" field.
func LastSignInAtEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldLastSignInAt, v))
}

// LastSignInAtNEQ applies the NEQ predicate on the "last_sign_in_at" field.
func LastSignInAtNEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldLastSignInAt, v))
}

// LastSignInAtIn applies the In predicate on the "last_sign_in_at" field.
func LastSignInAtIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldIn(FieldLastSignInAt, vs...))
}

// LastSignInAtNotIn applies the NotIn predicate on the "last_sign_in_at" field.
func LastSignInAtNotIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldLastSignInAt, vs...))
}

// LastSignInAtGT applies the GT predicate on the "last_sign_in_at" field.
func LastSignInAtGT(v time.Time) predicate.User {
	return predicate.User(sql.FieldGT(FieldLastSignInAt, v))
}

// LastSignInAtGTE applies the GTE predicate on the "last_sign_in_at" field.
func LastSignInAtGTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldGTE(FieldLastSignInAt, v))
}

// LastSignInAtLT applies the LT predicate on the "last_sign_in_at" field.
func LastSignInAtLT(v time.Time) predicate.User {
	return predicate.User(sql.FieldLT(FieldLastSignInAt, v))
}

// LastSignInAtLTE applies the LTE predicate on the "last_sign_in_at" field.
func LastSignInAtLTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldLTE(FieldLastSignInAt, v))
}

// LastSignInAtIsNil applies the IsNil predicate on the "last_sign_in_at" field.
func LastSignInAtIsNil() predicate.User {
	return predicate.User(sql.FieldIsNull(FieldLastSignInAt))
}

// LastSignInAtNotNil applies the NotNil predicate on the "last_sign_in_at" field.
func LastSignInAtNotNil() predicate.User {
	return predicate.User(sql.FieldNotNull(FieldLastSignInAt))
}

// LastSeenAtEQ applies the EQ predicate on the "last_seen_at" field.
func LastSeenAtEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldLastSeenAt, v))
}

// LastSeenAtNEQ applies the NEQ predicate on the "last_seen_at" field.
func LastSeenAtNEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldLastSeenAt, v))
}

// LastSeenAtIn applies the In predicate on the "last_seen_at" field.
func LastSeenAtIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldIn(FieldLastSeenAt, vs...))
}

// LastSeenAtNotIn applies the NotIn predicate on the "last_seen_at" field.
func LastSeenAtNotIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldLastSeenAt, vs...))
}

// LastSeenAtGT applies the GT predicate on the "last_seen_at" field.
func LastSeenAtGT(v time.Time) predicate.User {
	return predicate.User(sql.FieldGT(FieldLastSeenAt, v))
}

// LastSeenAtGTE applies the GTE predicate on the "last_seen_at" field.
func LastSeenAtGTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldGTE(FieldLastSeenAt, v))
}

// LastSeenAtLT applies the LT predicate on the "last_seen_at" field.
func LastSeenAtLT(v time.Time) predicate.User {
	return predicate.User(sql.FieldLT(FieldLastSeenAt, v))
}

// LastSeenAtLTE applies the LTE predicate on the "last_seen_at" field.
func LastSeenAtLTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldLTE(FieldLastSeenAt, v))
}

// BannedAtEQ applies the EQ predicate on the "banned_at" field.
func BannedAtEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldBannedAt, v))
}

// BannedAtNEQ applies the NEQ predicate on the "banned_at" field.
func BannedAtNEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldBannedAt, v))
}

// BannedAtIn applies the In predicate on the "banned_at" field.
func BannedAtIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldIn(FieldBannedAt, vs...))
}

// BannedAtNotIn applies the NotIn predicate on the "banned_at" field.
func BannedAtNotIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldBannedAt, vs...))
}

// BannedAtGT applies the GT predicate on the "banned_at" field.
func BannedAtGT(v time.Time) predicate.User {
	return predicate.User(sql.FieldGT(FieldBannedAt, v))
}

// BannedAtGTE applies the GTE predicate on the "banned_at" field.
func BannedAtGTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldGTE(FieldBannedAt, v))
}

// BannedAtLT applies the LT predicate on the "banned_at" field.
func BannedAtLT(v time.Time) predicate.User {
	return predicate.User(sql.FieldLT(FieldBannedAt, v))
}

// BannedAtLTE applies the LTE predicate on the "banned_at" field.
func BannedAtLTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldLTE(FieldBannedAt, v))
}

// BannedAtIsNil applies the IsNil predicate on the "banned_at" field.
func BannedAtIsNil() predicate.User {
	return predicate.User(sql.FieldIsNull(FieldBannedAt))
}

// BannedAtNotNil applies the NotNil predicate on the "banned_at" field.
func BannedAtNotNil() predicate.User {
	return predicate.User(sql.FieldNotNull(FieldBannedAt))
}

// ScheduledForDeletionOnEQ applies the EQ predicate on the "scheduled_for_deletion_on" field.
func ScheduledForDeletionOnEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldScheduledForDeletionOn, v))
}

// ScheduledForDeletionOnNEQ applies the NEQ predicate on the "scheduled_for_deletion_on" field.
func ScheduledForDeletionOnNEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldScheduledForDeletionOn, v))
}

// ScheduledForDeletionOnIn applies the In predicate on the "scheduled_for_deletion_on" field.
func ScheduledForDeletionOnIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldIn(FieldScheduledForDeletionOn, vs...))
}

// ScheduledForDeletionOnNotIn applies the NotIn predicate on the "scheduled_for_deletion_on" field.
func ScheduledForDeletionOnNotIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldScheduledForDeletionOn, vs...))
}

// ScheduledForDeletionOnGT applies the GT predicate on the "scheduled_for_deletion_on" field.
func ScheduledForDeletionOnGT(v time.Time) predicate.User {
	return predicate.User(sql.FieldGT(FieldScheduledForDeletionOn, v))
}

// ScheduledForDeletionOnGTE applies the GTE predicate on the "scheduled_for_deletion_on" field.
func ScheduledForDeletionOnGTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldGTE(FieldScheduledForDeletionOn, v))
}

// ScheduledForDeletionOnLT applies the LT predicate on the "scheduled_for_deletion_on" field.
func ScheduledForDeletionOnLT(v time.Time) predicate.User {
	return predicate.User(sql.FieldLT(FieldScheduledForDeletionOn, v))
}

// ScheduledForDeletionOnLTE applies the LTE predicate on the "scheduled_for_deletion_on" field.
func ScheduledForDeletionOnLTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldLTE(FieldScheduledForDeletionOn, v))
}

// ScheduledForDeletionOnIsNil applies the IsNil predicate on the "scheduled_for_deletion_on" field.
func ScheduledForDeletionOnIsNil() predicate.User {
	return predicate.User(sql.FieldIsNull(FieldScheduledForDeletionOn))
}

// ScheduledForDeletionOnNotNil applies the NotNil predicate on the "scheduled_for_deletion_on" field.
func ScheduledForDeletionOnNotNil() predicate.User {
	return predicate.User(sql.FieldNotNull(FieldScheduledForDeletionOn))
}

// DataDeletedAtEQ applies the EQ predicate on the "data_deleted_at" field.
func DataDeletedAtEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldDataDeletedAt, v))
}

// DataDeletedAtNEQ applies the NEQ predicate on the "data_deleted_at" field.
func DataDeletedAtNEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldDataDeletedAt, v))
}

// DataDeletedAtIn applies the In predicate on the "data_deleted_at" field.
func DataDeletedAtIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldIn(FieldDataDeletedAt, vs...))
}

// DataDeletedAtNotIn applies the NotIn predicate on the "data_deleted_at" field.
func DataDeletedAtNotIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldDataDeletedAt, vs...))
}

// DataDeletedAtGT applies the GT predicate on the "data_deleted_at" field.
func DataDeletedAtGT(v time.Time) predicate.User {
	return predicate.User(sql.FieldGT(FieldDataDeletedAt, v))
}

// DataDeletedAtGTE applies the GTE predicate on the "data_deleted_at" field.
func DataDeletedAtGTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldGTE(FieldDataDeletedAt, v))
}

// DataDeletedAtLT applies the LT predicate on the "data_deleted_at" field.
func DataDeletedAtLT(v time.Time) predicate.User {
	return predicate.User(sql.FieldLT(FieldDataDeletedAt, v))
}

// DataDeletedAtLTE applies the LTE predicate on the "data_deleted_at" field.
func DataDeletedAtLTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldLTE(FieldDataDeletedAt, v))
}

// DataDeletedAtIsNil applies the IsNil predicate on the "data_deleted_at" field.
func DataDeletedAtIsNil() predicate.User {
	return predicate.User(sql.FieldIsNull(FieldDataDeletedAt))
}

// DataDeletedAtNotNil applies the NotNil predicate on the "data_deleted_at" field.
func DataDeletedAtNotNil() predicate.User {
	return predicate.User(sql.FieldNotNull(FieldDataDeletedAt))
}

// WantsEmailEQ applies the EQ predicate on the "wants_email" field.
func WantsEmailEQ(v bool) predicate.User {
	return predicate.User(sql.FieldEQ(FieldWantsEmail, v))
}

// WantsEmailNEQ applies the NEQ predicate on the "wants_email" field.
func WantsEmailNEQ(v bool) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldWantsEmail, v))
}

// WantsEmailIsNil applies the IsNil predicate on the "wants_email" field.
func WantsEmailIsNil() predicate.User {
	return predicate.User(sql.FieldIsNull(FieldWantsEmail))
}

// WantsEmailNotNil applies the NotNil predicate on the "wants_email" field.
func WantsEmailNotNil() predicate.User {
	return predicate.User(sql.FieldNotNull(FieldWantsEmail))
}

// DiscordIDEQ applies the EQ predicate on the "discord_id" field.
func DiscordIDEQ(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldDiscordID, v))
}

// DiscordIDNEQ applies the NEQ predicate on the "discord_id" field.
func DiscordIDNEQ(v string) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldDiscordID, v))
}

// DiscordIDIn applies the In predicate on the "discord_id" field.
func DiscordIDIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldIn(FieldDiscordID, vs...))
}

// DiscordIDNotIn applies the NotIn predicate on the "discord_id" field.
func DiscordIDNotIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldDiscordID, vs...))
}

// DiscordIDGT applies the GT predicate on the "discord_id" field.
func DiscordIDGT(v string) predicate.User {
	return predicate.User(sql.FieldGT(FieldDiscordID, v))
}

// DiscordIDGTE applies the GTE predicate on the "discord_id" field.
func DiscordIDGTE(v string) predicate.User {
	return predicate.User(sql.FieldGTE(FieldDiscordID, v))
}

// DiscordIDLT applies the LT predicate on the "discord_id" field.
func DiscordIDLT(v string) predicate.User {
	return predicate.User(sql.FieldLT(FieldDiscordID, v))
}

// DiscordIDLTE applies the LTE predicate on the "discord_id" field.
func DiscordIDLTE(v string) predicate.User {
	return predicate.User(sql.FieldLTE(FieldDiscordID, v))
}

// DiscordIDContains applies the Contains predicate on the "discord_id" field.
func DiscordIDContains(v string) predicate.User {
	return predicate.User(sql.FieldContains(FieldDiscordID, v))
}

// DiscordIDHasPrefix applies the HasPrefix predicate on the "discord_id" field.
func DiscordIDHasPrefix(v string) predicate.User {
	return predicate.User(sql.FieldHasPrefix(FieldDiscordID, v))
}

// DiscordIDHasSuffix applies the HasSuffix predicate on the "discord_id" field.
func DiscordIDHasSuffix(v string) predicate.User {
	return predicate.User(sql.FieldHasSuffix(FieldDiscordID, v))
}

// DiscordIDIsNil applies the IsNil predicate on the "discord_id" field.
func DiscordIDIsNil() predicate.User {
	return predicate.User(sql.FieldIsNull(FieldDiscordID))
}

// DiscordIDNotNil applies the NotNil predicate on the "discord_id" field.
func DiscordIDNotNil() predicate.User {
	return predicate.User(sql.FieldNotNull(FieldDiscordID))
}

// DiscordIDEqualFold applies the EqualFold predicate on the "discord_id" field.
func DiscordIDEqualFold(v string) predicate.User {
	return predicate.User(sql.FieldEqualFold(FieldDiscordID, v))
}

// DiscordIDContainsFold applies the ContainsFold predicate on the "discord_id" field.
func DiscordIDContainsFold(v string) predicate.User {
	return predicate.User(sql.FieldContainsFold(FieldDiscordID, v))
}

// UsernameEQ applies the EQ predicate on the "username" field.
func UsernameEQ(v string) predicate.User {
	return predicate.User(sql.FieldEQ(FieldUsername, v))
}

// UsernameNEQ applies the NEQ predicate on the "username" field.
func UsernameNEQ(v string) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldUsername, v))
}

// UsernameIn applies the In predicate on the "username" field.
func UsernameIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldIn(FieldUsername, vs...))
}

// UsernameNotIn applies the NotIn predicate on the "username" field.
func UsernameNotIn(vs ...string) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldUsername, vs...))
}

// UsernameGT applies the GT predicate on the "username" field.
func UsernameGT(v string) predicate.User {
	return predicate.User(sql.FieldGT(FieldUsername, v))
}

// UsernameGTE applies the GTE predicate on the "username" field.
func UsernameGTE(v string) predicate.User {
	return predicate.User(sql.FieldGTE(FieldUsername, v))
}

// UsernameLT applies the LT predicate on the "username" field.
func UsernameLT(v string) predicate.User {
	return predicate.User(sql.FieldLT(FieldUsername, v))
}

// UsernameLTE applies the LTE predicate on the "username" field.
func UsernameLTE(v string) predicate.User {
	return predicate.User(sql.FieldLTE(FieldUsername, v))
}

// UsernameContains applies the Contains predicate on the "username" field.
func UsernameContains(v string) predicate.User {
	return predicate.User(sql.FieldContains(FieldUsername, v))
}

// UsernameHasPrefix applies the HasPrefix predicate on the "username" field.
func UsernameHasPrefix(v string) predicate.User {
	return predicate.User(sql.FieldHasPrefix(FieldUsername, v))
}

// UsernameHasSuffix applies the HasSuffix predicate on the "username" field.
func UsernameHasSuffix(v string) predicate.User {
	return predicate.User(sql.FieldHasSuffix(FieldUsername, v))
}

// UsernameEqualFold applies the EqualFold predicate on the "username" field.
func UsernameEqualFold(v string) predicate.User {
	return predicate.User(sql.FieldEqualFold(FieldUsername, v))
}

// UsernameContainsFold applies the ContainsFold predicate on the "username" field.
func UsernameContainsFold(v string) predicate.User {
	return predicate.User(sql.FieldContainsFold(FieldUsername, v))
}

// CreatedAtEQ applies the EQ predicate on the "created_at" field.
func CreatedAtEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldCreatedAt, v))
}

// CreatedAtNEQ applies the NEQ predicate on the "created_at" field.
func CreatedAtNEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldCreatedAt, v))
}

// CreatedAtIn applies the In predicate on the "created_at" field.
func CreatedAtIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldIn(FieldCreatedAt, vs...))
}

// CreatedAtNotIn applies the NotIn predicate on the "created_at" field.
func CreatedAtNotIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldCreatedAt, vs...))
}

// CreatedAtGT applies the GT predicate on the "created_at" field.
func CreatedAtGT(v time.Time) predicate.User {
	return predicate.User(sql.FieldGT(FieldCreatedAt, v))
}

// CreatedAtGTE applies the GTE predicate on the "created_at" field.
func CreatedAtGTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldGTE(FieldCreatedAt, v))
}

// CreatedAtLT applies the LT predicate on the "created_at" field.
func CreatedAtLT(v time.Time) predicate.User {
	return predicate.User(sql.FieldLT(FieldCreatedAt, v))
}

// CreatedAtLTE applies the LTE predicate on the "created_at" field.
func CreatedAtLTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldLTE(FieldCreatedAt, v))
}

// UpdatedAtEQ applies the EQ predicate on the "updated_at" field.
func UpdatedAtEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldEQ(FieldUpdatedAt, v))
}

// UpdatedAtNEQ applies the NEQ predicate on the "updated_at" field.
func UpdatedAtNEQ(v time.Time) predicate.User {
	return predicate.User(sql.FieldNEQ(FieldUpdatedAt, v))
}

// UpdatedAtIn applies the In predicate on the "updated_at" field.
func UpdatedAtIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldIn(FieldUpdatedAt, vs...))
}

// UpdatedAtNotIn applies the NotIn predicate on the "updated_at" field.
func UpdatedAtNotIn(vs ...time.Time) predicate.User {
	return predicate.User(sql.FieldNotIn(FieldUpdatedAt, vs...))
}

// UpdatedAtGT applies the GT predicate on the "updated_at" field.
func UpdatedAtGT(v time.Time) predicate.User {
	return predicate.User(sql.FieldGT(FieldUpdatedAt, v))
}

// UpdatedAtGTE applies the GTE predicate on the "updated_at" field.
func UpdatedAtGTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldGTE(FieldUpdatedAt, v))
}

// UpdatedAtLT applies the LT predicate on the "updated_at" field.
func UpdatedAtLT(v time.Time) predicate.User {
	return predicate.User(sql.FieldLT(FieldUpdatedAt, v))
}

// UpdatedAtLTE applies the LTE predicate on the "updated_at" field.
func UpdatedAtLTE(v time.Time) predicate.User {
	return predicate.User(sql.FieldLTE(FieldUpdatedAt, v))
}

// HasGenerations applies the HasEdge predicate on the "generations" edge.
func HasGenerations() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, GenerationsTable, GenerationsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasGenerationsWith applies the HasEdge predicate on the "generations" edge with a given conditions (other predicates).
func HasGenerationsWith(preds ...predicate.Generation) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(GenerationsInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, GenerationsTable, GenerationsColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasUpscales applies the HasEdge predicate on the "upscales" edge.
func HasUpscales() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, UpscalesTable, UpscalesColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasUpscalesWith applies the HasEdge predicate on the "upscales" edge with a given conditions (other predicates).
func HasUpscalesWith(preds ...predicate.Upscale) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(UpscalesInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, UpscalesTable, UpscalesColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasVoiceovers applies the HasEdge predicate on the "voiceovers" edge.
func HasVoiceovers() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, VoiceoversTable, VoiceoversColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasVoiceoversWith applies the HasEdge predicate on the "voiceovers" edge with a given conditions (other predicates).
func HasVoiceoversWith(preds ...predicate.Voiceover) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(VoiceoversInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, VoiceoversTable, VoiceoversColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasCredits applies the HasEdge predicate on the "credits" edge.
func HasCredits() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, CreditsTable, CreditsColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasCreditsWith applies the HasEdge predicate on the "credits" edge with a given conditions (other predicates).
func HasCreditsWith(preds ...predicate.Credit) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(CreditsInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, CreditsTable, CreditsColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasAPITokens applies the HasEdge predicate on the "api_tokens" edge.
func HasAPITokens() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, APITokensTable, APITokensColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasAPITokensWith applies the HasEdge predicate on the "api_tokens" edge with a given conditions (other predicates).
func HasAPITokensWith(preds ...predicate.ApiToken) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(APITokensInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, APITokensTable, APITokensColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasTipsGiven applies the HasEdge predicate on the "tips_given" edge.
func HasTipsGiven() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, TipsGivenTable, TipsGivenColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasTipsGivenWith applies the HasEdge predicate on the "tips_given" edge with a given conditions (other predicates).
func HasTipsGivenWith(preds ...predicate.TipLog) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(TipsGivenInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, TipsGivenTable, TipsGivenColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasTipsReceived applies the HasEdge predicate on the "tips_received" edge.
func HasTipsReceived() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, TipsReceivedTable, TipsReceivedColumn),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasTipsReceivedWith applies the HasEdge predicate on the "tips_received" edge with a given conditions (other predicates).
func HasTipsReceivedWith(preds ...predicate.TipLog) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(TipsReceivedInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.O2M, false, TipsReceivedTable, TipsReceivedColumn),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// HasRoles applies the HasEdge predicate on the "roles" edge.
func HasRoles() predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, RolesTable, RolesPrimaryKey...),
		)
		sqlgraph.HasNeighbors(s, step)
	})
}

// HasRolesWith applies the HasEdge predicate on the "roles" edge with a given conditions (other predicates).
func HasRolesWith(preds ...predicate.Role) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		step := sqlgraph.NewStep(
			sqlgraph.From(Table, FieldID),
			sqlgraph.To(RolesInverseTable, FieldID),
			sqlgraph.Edge(sqlgraph.M2M, true, RolesTable, RolesPrimaryKey...),
		)
		sqlgraph.HasNeighborsWith(s, step, func(s *sql.Selector) {
			for _, p := range preds {
				p(s)
			}
		})
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.User) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for _, p := range predicates {
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.User) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for i, p := range predicates {
			if i > 0 {
				s1.Or()
			}
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Not applies the not operator on the given predicate.
func Not(p predicate.User) predicate.User {
	return predicate.User(func(s *sql.Selector) {
		p(s.Not())
	})
}
