// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/credit"
	"github.com/stablecog/sc-go/database/ent/credittype"
	"github.com/stablecog/sc-go/database/ent/user"
)

// CreditCreate is the builder for creating a Credit entity.
type CreditCreate struct {
	config
	mutation *CreditMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetRemainingAmount sets the "remaining_amount" field.
func (cc *CreditCreate) SetRemainingAmount(i int32) *CreditCreate {
	cc.mutation.SetRemainingAmount(i)
	return cc
}

// SetStartsAt sets the "starts_at" field.
func (cc *CreditCreate) SetStartsAt(t time.Time) *CreditCreate {
	cc.mutation.SetStartsAt(t)
	return cc
}

// SetNillableStartsAt sets the "starts_at" field if the given value is not nil.
func (cc *CreditCreate) SetNillableStartsAt(t *time.Time) *CreditCreate {
	if t != nil {
		cc.SetStartsAt(*t)
	}
	return cc
}

// SetExpiresAt sets the "expires_at" field.
func (cc *CreditCreate) SetExpiresAt(t time.Time) *CreditCreate {
	cc.mutation.SetExpiresAt(t)
	return cc
}

// SetPeriod sets the "period" field.
func (cc *CreditCreate) SetPeriod(i int) *CreditCreate {
	cc.mutation.SetPeriod(i)
	return cc
}

// SetNillablePeriod sets the "period" field if the given value is not nil.
func (cc *CreditCreate) SetNillablePeriod(i *int) *CreditCreate {
	if i != nil {
		cc.SetPeriod(*i)
	}
	return cc
}

// SetStripeLineItemID sets the "stripe_line_item_id" field.
func (cc *CreditCreate) SetStripeLineItemID(s string) *CreditCreate {
	cc.mutation.SetStripeLineItemID(s)
	return cc
}

// SetNillableStripeLineItemID sets the "stripe_line_item_id" field if the given value is not nil.
func (cc *CreditCreate) SetNillableStripeLineItemID(s *string) *CreditCreate {
	if s != nil {
		cc.SetStripeLineItemID(*s)
	}
	return cc
}

// SetReplenishedAt sets the "replenished_at" field.
func (cc *CreditCreate) SetReplenishedAt(t time.Time) *CreditCreate {
	cc.mutation.SetReplenishedAt(t)
	return cc
}

// SetNillableReplenishedAt sets the "replenished_at" field if the given value is not nil.
func (cc *CreditCreate) SetNillableReplenishedAt(t *time.Time) *CreditCreate {
	if t != nil {
		cc.SetReplenishedAt(*t)
	}
	return cc
}

// SetUserID sets the "user_id" field.
func (cc *CreditCreate) SetUserID(u uuid.UUID) *CreditCreate {
	cc.mutation.SetUserID(u)
	return cc
}

// SetCreditTypeID sets the "credit_type_id" field.
func (cc *CreditCreate) SetCreditTypeID(u uuid.UUID) *CreditCreate {
	cc.mutation.SetCreditTypeID(u)
	return cc
}

// SetCreatedAt sets the "created_at" field.
func (cc *CreditCreate) SetCreatedAt(t time.Time) *CreditCreate {
	cc.mutation.SetCreatedAt(t)
	return cc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (cc *CreditCreate) SetNillableCreatedAt(t *time.Time) *CreditCreate {
	if t != nil {
		cc.SetCreatedAt(*t)
	}
	return cc
}

// SetUpdatedAt sets the "updated_at" field.
func (cc *CreditCreate) SetUpdatedAt(t time.Time) *CreditCreate {
	cc.mutation.SetUpdatedAt(t)
	return cc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (cc *CreditCreate) SetNillableUpdatedAt(t *time.Time) *CreditCreate {
	if t != nil {
		cc.SetUpdatedAt(*t)
	}
	return cc
}

// SetID sets the "id" field.
func (cc *CreditCreate) SetID(u uuid.UUID) *CreditCreate {
	cc.mutation.SetID(u)
	return cc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (cc *CreditCreate) SetNillableID(u *uuid.UUID) *CreditCreate {
	if u != nil {
		cc.SetID(*u)
	}
	return cc
}

// SetUsersID sets the "users" edge to the User entity by ID.
func (cc *CreditCreate) SetUsersID(id uuid.UUID) *CreditCreate {
	cc.mutation.SetUsersID(id)
	return cc
}

// SetUsers sets the "users" edge to the User entity.
func (cc *CreditCreate) SetUsers(u *User) *CreditCreate {
	return cc.SetUsersID(u.ID)
}

// SetCreditType sets the "credit_type" edge to the CreditType entity.
func (cc *CreditCreate) SetCreditType(c *CreditType) *CreditCreate {
	return cc.SetCreditTypeID(c.ID)
}

// Mutation returns the CreditMutation object of the builder.
func (cc *CreditCreate) Mutation() *CreditMutation {
	return cc.mutation
}

// Save creates the Credit in the database.
func (cc *CreditCreate) Save(ctx context.Context) (*Credit, error) {
	cc.defaults()
	return withHooks(ctx, cc.sqlSave, cc.mutation, cc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (cc *CreditCreate) SaveX(ctx context.Context) *Credit {
	v, err := cc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (cc *CreditCreate) Exec(ctx context.Context) error {
	_, err := cc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (cc *CreditCreate) ExecX(ctx context.Context) {
	if err := cc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (cc *CreditCreate) defaults() {
	if _, ok := cc.mutation.StartsAt(); !ok {
		v := credit.DefaultStartsAt
		cc.mutation.SetStartsAt(v)
	}
	if _, ok := cc.mutation.Period(); !ok {
		v := credit.DefaultPeriod
		cc.mutation.SetPeriod(v)
	}
	if _, ok := cc.mutation.ReplenishedAt(); !ok {
		v := credit.DefaultReplenishedAt()
		cc.mutation.SetReplenishedAt(v)
	}
	if _, ok := cc.mutation.CreatedAt(); !ok {
		v := credit.DefaultCreatedAt()
		cc.mutation.SetCreatedAt(v)
	}
	if _, ok := cc.mutation.UpdatedAt(); !ok {
		v := credit.DefaultUpdatedAt()
		cc.mutation.SetUpdatedAt(v)
	}
	if _, ok := cc.mutation.ID(); !ok {
		v := credit.DefaultID()
		cc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (cc *CreditCreate) check() error {
	if _, ok := cc.mutation.RemainingAmount(); !ok {
		return &ValidationError{Name: "remaining_amount", err: errors.New(`ent: missing required field "Credit.remaining_amount"`)}
	}
	if _, ok := cc.mutation.StartsAt(); !ok {
		return &ValidationError{Name: "starts_at", err: errors.New(`ent: missing required field "Credit.starts_at"`)}
	}
	if _, ok := cc.mutation.ExpiresAt(); !ok {
		return &ValidationError{Name: "expires_at", err: errors.New(`ent: missing required field "Credit.expires_at"`)}
	}
	if _, ok := cc.mutation.Period(); !ok {
		return &ValidationError{Name: "period", err: errors.New(`ent: missing required field "Credit.period"`)}
	}
	if _, ok := cc.mutation.ReplenishedAt(); !ok {
		return &ValidationError{Name: "replenished_at", err: errors.New(`ent: missing required field "Credit.replenished_at"`)}
	}
	if _, ok := cc.mutation.UserID(); !ok {
		return &ValidationError{Name: "user_id", err: errors.New(`ent: missing required field "Credit.user_id"`)}
	}
	if _, ok := cc.mutation.CreditTypeID(); !ok {
		return &ValidationError{Name: "credit_type_id", err: errors.New(`ent: missing required field "Credit.credit_type_id"`)}
	}
	if _, ok := cc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "Credit.created_at"`)}
	}
	if _, ok := cc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "Credit.updated_at"`)}
	}
	if len(cc.mutation.UsersIDs()) == 0 {
		return &ValidationError{Name: "users", err: errors.New(`ent: missing required edge "Credit.users"`)}
	}
	if len(cc.mutation.CreditTypeIDs()) == 0 {
		return &ValidationError{Name: "credit_type", err: errors.New(`ent: missing required edge "Credit.credit_type"`)}
	}
	return nil
}

func (cc *CreditCreate) sqlSave(ctx context.Context) (*Credit, error) {
	if err := cc.check(); err != nil {
		return nil, err
	}
	_node, _spec := cc.createSpec()
	if err := sqlgraph.CreateNode(ctx, cc.driver, _spec); err != nil {
		if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	if _spec.ID.Value != nil {
		if id, ok := _spec.ID.Value.(*uuid.UUID); ok {
			_node.ID = *id
		} else if err := _node.ID.Scan(_spec.ID.Value); err != nil {
			return nil, err
		}
	}
	cc.mutation.id = &_node.ID
	cc.mutation.done = true
	return _node, nil
}

func (cc *CreditCreate) createSpec() (*Credit, *sqlgraph.CreateSpec) {
	var (
		_node = &Credit{config: cc.config}
		_spec = sqlgraph.NewCreateSpec(credit.Table, sqlgraph.NewFieldSpec(credit.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = cc.conflict
	if id, ok := cc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := cc.mutation.RemainingAmount(); ok {
		_spec.SetField(credit.FieldRemainingAmount, field.TypeInt32, value)
		_node.RemainingAmount = value
	}
	if value, ok := cc.mutation.StartsAt(); ok {
		_spec.SetField(credit.FieldStartsAt, field.TypeTime, value)
		_node.StartsAt = value
	}
	if value, ok := cc.mutation.ExpiresAt(); ok {
		_spec.SetField(credit.FieldExpiresAt, field.TypeTime, value)
		_node.ExpiresAt = value
	}
	if value, ok := cc.mutation.Period(); ok {
		_spec.SetField(credit.FieldPeriod, field.TypeInt, value)
		_node.Period = value
	}
	if value, ok := cc.mutation.StripeLineItemID(); ok {
		_spec.SetField(credit.FieldStripeLineItemID, field.TypeString, value)
		_node.StripeLineItemID = &value
	}
	if value, ok := cc.mutation.ReplenishedAt(); ok {
		_spec.SetField(credit.FieldReplenishedAt, field.TypeTime, value)
		_node.ReplenishedAt = value
	}
	if value, ok := cc.mutation.CreatedAt(); ok {
		_spec.SetField(credit.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := cc.mutation.UpdatedAt(); ok {
		_spec.SetField(credit.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if nodes := cc.mutation.UsersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   credit.UsersTable,
			Columns: []string{credit.UsersColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.UserID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := cc.mutation.CreditTypeIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   credit.CreditTypeTable,
			Columns: []string{credit.CreditTypeColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(credittype.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.CreditTypeID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Credit.Create().
//		SetRemainingAmount(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.CreditUpsert) {
//			SetRemainingAmount(v+v).
//		}).
//		Exec(ctx)
func (cc *CreditCreate) OnConflict(opts ...sql.ConflictOption) *CreditUpsertOne {
	cc.conflict = opts
	return &CreditUpsertOne{
		create: cc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Credit.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (cc *CreditCreate) OnConflictColumns(columns ...string) *CreditUpsertOne {
	cc.conflict = append(cc.conflict, sql.ConflictColumns(columns...))
	return &CreditUpsertOne{
		create: cc,
	}
}

type (
	// CreditUpsertOne is the builder for "upsert"-ing
	//  one Credit node.
	CreditUpsertOne struct {
		create *CreditCreate
	}

	// CreditUpsert is the "OnConflict" setter.
	CreditUpsert struct {
		*sql.UpdateSet
	}
)

// SetRemainingAmount sets the "remaining_amount" field.
func (u *CreditUpsert) SetRemainingAmount(v int32) *CreditUpsert {
	u.Set(credit.FieldRemainingAmount, v)
	return u
}

// UpdateRemainingAmount sets the "remaining_amount" field to the value that was provided on create.
func (u *CreditUpsert) UpdateRemainingAmount() *CreditUpsert {
	u.SetExcluded(credit.FieldRemainingAmount)
	return u
}

// AddRemainingAmount adds v to the "remaining_amount" field.
func (u *CreditUpsert) AddRemainingAmount(v int32) *CreditUpsert {
	u.Add(credit.FieldRemainingAmount, v)
	return u
}

// SetStartsAt sets the "starts_at" field.
func (u *CreditUpsert) SetStartsAt(v time.Time) *CreditUpsert {
	u.Set(credit.FieldStartsAt, v)
	return u
}

// UpdateStartsAt sets the "starts_at" field to the value that was provided on create.
func (u *CreditUpsert) UpdateStartsAt() *CreditUpsert {
	u.SetExcluded(credit.FieldStartsAt)
	return u
}

// SetExpiresAt sets the "expires_at" field.
func (u *CreditUpsert) SetExpiresAt(v time.Time) *CreditUpsert {
	u.Set(credit.FieldExpiresAt, v)
	return u
}

// UpdateExpiresAt sets the "expires_at" field to the value that was provided on create.
func (u *CreditUpsert) UpdateExpiresAt() *CreditUpsert {
	u.SetExcluded(credit.FieldExpiresAt)
	return u
}

// SetPeriod sets the "period" field.
func (u *CreditUpsert) SetPeriod(v int) *CreditUpsert {
	u.Set(credit.FieldPeriod, v)
	return u
}

// UpdatePeriod sets the "period" field to the value that was provided on create.
func (u *CreditUpsert) UpdatePeriod() *CreditUpsert {
	u.SetExcluded(credit.FieldPeriod)
	return u
}

// AddPeriod adds v to the "period" field.
func (u *CreditUpsert) AddPeriod(v int) *CreditUpsert {
	u.Add(credit.FieldPeriod, v)
	return u
}

// SetStripeLineItemID sets the "stripe_line_item_id" field.
func (u *CreditUpsert) SetStripeLineItemID(v string) *CreditUpsert {
	u.Set(credit.FieldStripeLineItemID, v)
	return u
}

// UpdateStripeLineItemID sets the "stripe_line_item_id" field to the value that was provided on create.
func (u *CreditUpsert) UpdateStripeLineItemID() *CreditUpsert {
	u.SetExcluded(credit.FieldStripeLineItemID)
	return u
}

// ClearStripeLineItemID clears the value of the "stripe_line_item_id" field.
func (u *CreditUpsert) ClearStripeLineItemID() *CreditUpsert {
	u.SetNull(credit.FieldStripeLineItemID)
	return u
}

// SetReplenishedAt sets the "replenished_at" field.
func (u *CreditUpsert) SetReplenishedAt(v time.Time) *CreditUpsert {
	u.Set(credit.FieldReplenishedAt, v)
	return u
}

// UpdateReplenishedAt sets the "replenished_at" field to the value that was provided on create.
func (u *CreditUpsert) UpdateReplenishedAt() *CreditUpsert {
	u.SetExcluded(credit.FieldReplenishedAt)
	return u
}

// SetUserID sets the "user_id" field.
func (u *CreditUpsert) SetUserID(v uuid.UUID) *CreditUpsert {
	u.Set(credit.FieldUserID, v)
	return u
}

// UpdateUserID sets the "user_id" field to the value that was provided on create.
func (u *CreditUpsert) UpdateUserID() *CreditUpsert {
	u.SetExcluded(credit.FieldUserID)
	return u
}

// SetCreditTypeID sets the "credit_type_id" field.
func (u *CreditUpsert) SetCreditTypeID(v uuid.UUID) *CreditUpsert {
	u.Set(credit.FieldCreditTypeID, v)
	return u
}

// UpdateCreditTypeID sets the "credit_type_id" field to the value that was provided on create.
func (u *CreditUpsert) UpdateCreditTypeID() *CreditUpsert {
	u.SetExcluded(credit.FieldCreditTypeID)
	return u
}

// SetUpdatedAt sets the "updated_at" field.
func (u *CreditUpsert) SetUpdatedAt(v time.Time) *CreditUpsert {
	u.Set(credit.FieldUpdatedAt, v)
	return u
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *CreditUpsert) UpdateUpdatedAt() *CreditUpsert {
	u.SetExcluded(credit.FieldUpdatedAt)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.Credit.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(credit.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *CreditUpsertOne) UpdateNewValues() *CreditUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(credit.FieldID)
		}
		if _, exists := u.create.mutation.CreatedAt(); exists {
			s.SetIgnore(credit.FieldCreatedAt)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Credit.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *CreditUpsertOne) Ignore() *CreditUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *CreditUpsertOne) DoNothing() *CreditUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the CreditCreate.OnConflict
// documentation for more info.
func (u *CreditUpsertOne) Update(set func(*CreditUpsert)) *CreditUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&CreditUpsert{UpdateSet: update})
	}))
	return u
}

// SetRemainingAmount sets the "remaining_amount" field.
func (u *CreditUpsertOne) SetRemainingAmount(v int32) *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.SetRemainingAmount(v)
	})
}

// AddRemainingAmount adds v to the "remaining_amount" field.
func (u *CreditUpsertOne) AddRemainingAmount(v int32) *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.AddRemainingAmount(v)
	})
}

// UpdateRemainingAmount sets the "remaining_amount" field to the value that was provided on create.
func (u *CreditUpsertOne) UpdateRemainingAmount() *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateRemainingAmount()
	})
}

// SetStartsAt sets the "starts_at" field.
func (u *CreditUpsertOne) SetStartsAt(v time.Time) *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.SetStartsAt(v)
	})
}

// UpdateStartsAt sets the "starts_at" field to the value that was provided on create.
func (u *CreditUpsertOne) UpdateStartsAt() *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateStartsAt()
	})
}

// SetExpiresAt sets the "expires_at" field.
func (u *CreditUpsertOne) SetExpiresAt(v time.Time) *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.SetExpiresAt(v)
	})
}

// UpdateExpiresAt sets the "expires_at" field to the value that was provided on create.
func (u *CreditUpsertOne) UpdateExpiresAt() *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateExpiresAt()
	})
}

// SetPeriod sets the "period" field.
func (u *CreditUpsertOne) SetPeriod(v int) *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.SetPeriod(v)
	})
}

// AddPeriod adds v to the "period" field.
func (u *CreditUpsertOne) AddPeriod(v int) *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.AddPeriod(v)
	})
}

// UpdatePeriod sets the "period" field to the value that was provided on create.
func (u *CreditUpsertOne) UpdatePeriod() *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.UpdatePeriod()
	})
}

// SetStripeLineItemID sets the "stripe_line_item_id" field.
func (u *CreditUpsertOne) SetStripeLineItemID(v string) *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.SetStripeLineItemID(v)
	})
}

// UpdateStripeLineItemID sets the "stripe_line_item_id" field to the value that was provided on create.
func (u *CreditUpsertOne) UpdateStripeLineItemID() *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateStripeLineItemID()
	})
}

// ClearStripeLineItemID clears the value of the "stripe_line_item_id" field.
func (u *CreditUpsertOne) ClearStripeLineItemID() *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.ClearStripeLineItemID()
	})
}

// SetReplenishedAt sets the "replenished_at" field.
func (u *CreditUpsertOne) SetReplenishedAt(v time.Time) *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.SetReplenishedAt(v)
	})
}

// UpdateReplenishedAt sets the "replenished_at" field to the value that was provided on create.
func (u *CreditUpsertOne) UpdateReplenishedAt() *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateReplenishedAt()
	})
}

// SetUserID sets the "user_id" field.
func (u *CreditUpsertOne) SetUserID(v uuid.UUID) *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.SetUserID(v)
	})
}

// UpdateUserID sets the "user_id" field to the value that was provided on create.
func (u *CreditUpsertOne) UpdateUserID() *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateUserID()
	})
}

// SetCreditTypeID sets the "credit_type_id" field.
func (u *CreditUpsertOne) SetCreditTypeID(v uuid.UUID) *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.SetCreditTypeID(v)
	})
}

// UpdateCreditTypeID sets the "credit_type_id" field to the value that was provided on create.
func (u *CreditUpsertOne) UpdateCreditTypeID() *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateCreditTypeID()
	})
}

// SetUpdatedAt sets the "updated_at" field.
func (u *CreditUpsertOne) SetUpdatedAt(v time.Time) *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *CreditUpsertOne) UpdateUpdatedAt() *CreditUpsertOne {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateUpdatedAt()
	})
}

// Exec executes the query.
func (u *CreditUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for CreditCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *CreditUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *CreditUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: CreditUpsertOne.ID is not supported by MySQL driver. Use CreditUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *CreditUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// CreditCreateBulk is the builder for creating many Credit entities in bulk.
type CreditCreateBulk struct {
	config
	err      error
	builders []*CreditCreate
	conflict []sql.ConflictOption
}

// Save creates the Credit entities in the database.
func (ccb *CreditCreateBulk) Save(ctx context.Context) ([]*Credit, error) {
	if ccb.err != nil {
		return nil, ccb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(ccb.builders))
	nodes := make([]*Credit, len(ccb.builders))
	mutators := make([]Mutator, len(ccb.builders))
	for i := range ccb.builders {
		func(i int, root context.Context) {
			builder := ccb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*CreditMutation)
				if !ok {
					return nil, fmt.Errorf("unexpected mutation type %T", m)
				}
				if err := builder.check(); err != nil {
					return nil, err
				}
				builder.mutation = mutation
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, ccb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = ccb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ccb.driver, spec); err != nil {
						if sqlgraph.IsConstraintError(err) {
							err = &ConstraintError{msg: err.Error(), wrap: err}
						}
					}
				}
				if err != nil {
					return nil, err
				}
				mutation.id = &nodes[i].ID
				mutation.done = true
				return nodes[i], nil
			})
			for i := len(builder.hooks) - 1; i >= 0; i-- {
				mut = builder.hooks[i](mut)
			}
			mutators[i] = mut
		}(i, ctx)
	}
	if len(mutators) > 0 {
		if _, err := mutators[0].Mutate(ctx, ccb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ccb *CreditCreateBulk) SaveX(ctx context.Context) []*Credit {
	v, err := ccb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ccb *CreditCreateBulk) Exec(ctx context.Context) error {
	_, err := ccb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ccb *CreditCreateBulk) ExecX(ctx context.Context) {
	if err := ccb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Credit.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.CreditUpsert) {
//			SetRemainingAmount(v+v).
//		}).
//		Exec(ctx)
func (ccb *CreditCreateBulk) OnConflict(opts ...sql.ConflictOption) *CreditUpsertBulk {
	ccb.conflict = opts
	return &CreditUpsertBulk{
		create: ccb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Credit.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (ccb *CreditCreateBulk) OnConflictColumns(columns ...string) *CreditUpsertBulk {
	ccb.conflict = append(ccb.conflict, sql.ConflictColumns(columns...))
	return &CreditUpsertBulk{
		create: ccb,
	}
}

// CreditUpsertBulk is the builder for "upsert"-ing
// a bulk of Credit nodes.
type CreditUpsertBulk struct {
	create *CreditCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Credit.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(credit.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *CreditUpsertBulk) UpdateNewValues() *CreditUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(credit.FieldID)
			}
			if _, exists := b.mutation.CreatedAt(); exists {
				s.SetIgnore(credit.FieldCreatedAt)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Credit.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *CreditUpsertBulk) Ignore() *CreditUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *CreditUpsertBulk) DoNothing() *CreditUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the CreditCreateBulk.OnConflict
// documentation for more info.
func (u *CreditUpsertBulk) Update(set func(*CreditUpsert)) *CreditUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&CreditUpsert{UpdateSet: update})
	}))
	return u
}

// SetRemainingAmount sets the "remaining_amount" field.
func (u *CreditUpsertBulk) SetRemainingAmount(v int32) *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.SetRemainingAmount(v)
	})
}

// AddRemainingAmount adds v to the "remaining_amount" field.
func (u *CreditUpsertBulk) AddRemainingAmount(v int32) *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.AddRemainingAmount(v)
	})
}

// UpdateRemainingAmount sets the "remaining_amount" field to the value that was provided on create.
func (u *CreditUpsertBulk) UpdateRemainingAmount() *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateRemainingAmount()
	})
}

// SetStartsAt sets the "starts_at" field.
func (u *CreditUpsertBulk) SetStartsAt(v time.Time) *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.SetStartsAt(v)
	})
}

// UpdateStartsAt sets the "starts_at" field to the value that was provided on create.
func (u *CreditUpsertBulk) UpdateStartsAt() *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateStartsAt()
	})
}

// SetExpiresAt sets the "expires_at" field.
func (u *CreditUpsertBulk) SetExpiresAt(v time.Time) *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.SetExpiresAt(v)
	})
}

// UpdateExpiresAt sets the "expires_at" field to the value that was provided on create.
func (u *CreditUpsertBulk) UpdateExpiresAt() *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateExpiresAt()
	})
}

// SetPeriod sets the "period" field.
func (u *CreditUpsertBulk) SetPeriod(v int) *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.SetPeriod(v)
	})
}

// AddPeriod adds v to the "period" field.
func (u *CreditUpsertBulk) AddPeriod(v int) *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.AddPeriod(v)
	})
}

// UpdatePeriod sets the "period" field to the value that was provided on create.
func (u *CreditUpsertBulk) UpdatePeriod() *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.UpdatePeriod()
	})
}

// SetStripeLineItemID sets the "stripe_line_item_id" field.
func (u *CreditUpsertBulk) SetStripeLineItemID(v string) *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.SetStripeLineItemID(v)
	})
}

// UpdateStripeLineItemID sets the "stripe_line_item_id" field to the value that was provided on create.
func (u *CreditUpsertBulk) UpdateStripeLineItemID() *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateStripeLineItemID()
	})
}

// ClearStripeLineItemID clears the value of the "stripe_line_item_id" field.
func (u *CreditUpsertBulk) ClearStripeLineItemID() *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.ClearStripeLineItemID()
	})
}

// SetReplenishedAt sets the "replenished_at" field.
func (u *CreditUpsertBulk) SetReplenishedAt(v time.Time) *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.SetReplenishedAt(v)
	})
}

// UpdateReplenishedAt sets the "replenished_at" field to the value that was provided on create.
func (u *CreditUpsertBulk) UpdateReplenishedAt() *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateReplenishedAt()
	})
}

// SetUserID sets the "user_id" field.
func (u *CreditUpsertBulk) SetUserID(v uuid.UUID) *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.SetUserID(v)
	})
}

// UpdateUserID sets the "user_id" field to the value that was provided on create.
func (u *CreditUpsertBulk) UpdateUserID() *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateUserID()
	})
}

// SetCreditTypeID sets the "credit_type_id" field.
func (u *CreditUpsertBulk) SetCreditTypeID(v uuid.UUID) *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.SetCreditTypeID(v)
	})
}

// UpdateCreditTypeID sets the "credit_type_id" field to the value that was provided on create.
func (u *CreditUpsertBulk) UpdateCreditTypeID() *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateCreditTypeID()
	})
}

// SetUpdatedAt sets the "updated_at" field.
func (u *CreditUpsertBulk) SetUpdatedAt(v time.Time) *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *CreditUpsertBulk) UpdateUpdatedAt() *CreditUpsertBulk {
	return u.Update(func(s *CreditUpsert) {
		s.UpdateUpdatedAt()
	})
}

// Exec executes the query.
func (u *CreditUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the CreditCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for CreditCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *CreditUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
