// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/credit"
	"github.com/stablecog/sc-go/database/ent/credittype"
	"github.com/stablecog/sc-go/database/ent/predicate"
	"github.com/stablecog/sc-go/database/ent/user"
)

// CreditUpdate is the builder for updating Credit entities.
type CreditUpdate struct {
	config
	hooks     []Hook
	mutation  *CreditMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the CreditUpdate builder.
func (cu *CreditUpdate) Where(ps ...predicate.Credit) *CreditUpdate {
	cu.mutation.Where(ps...)
	return cu
}

// SetRemainingAmount sets the "remaining_amount" field.
func (cu *CreditUpdate) SetRemainingAmount(i int32) *CreditUpdate {
	cu.mutation.ResetRemainingAmount()
	cu.mutation.SetRemainingAmount(i)
	return cu
}

// SetNillableRemainingAmount sets the "remaining_amount" field if the given value is not nil.
func (cu *CreditUpdate) SetNillableRemainingAmount(i *int32) *CreditUpdate {
	if i != nil {
		cu.SetRemainingAmount(*i)
	}
	return cu
}

// AddRemainingAmount adds i to the "remaining_amount" field.
func (cu *CreditUpdate) AddRemainingAmount(i int32) *CreditUpdate {
	cu.mutation.AddRemainingAmount(i)
	return cu
}

// SetExpiresAt sets the "expires_at" field.
func (cu *CreditUpdate) SetExpiresAt(t time.Time) *CreditUpdate {
	cu.mutation.SetExpiresAt(t)
	return cu
}

// SetNillableExpiresAt sets the "expires_at" field if the given value is not nil.
func (cu *CreditUpdate) SetNillableExpiresAt(t *time.Time) *CreditUpdate {
	if t != nil {
		cu.SetExpiresAt(*t)
	}
	return cu
}

// SetStripeLineItemID sets the "stripe_line_item_id" field.
func (cu *CreditUpdate) SetStripeLineItemID(s string) *CreditUpdate {
	cu.mutation.SetStripeLineItemID(s)
	return cu
}

// SetNillableStripeLineItemID sets the "stripe_line_item_id" field if the given value is not nil.
func (cu *CreditUpdate) SetNillableStripeLineItemID(s *string) *CreditUpdate {
	if s != nil {
		cu.SetStripeLineItemID(*s)
	}
	return cu
}

// ClearStripeLineItemID clears the value of the "stripe_line_item_id" field.
func (cu *CreditUpdate) ClearStripeLineItemID() *CreditUpdate {
	cu.mutation.ClearStripeLineItemID()
	return cu
}

// SetReplenishedAt sets the "replenished_at" field.
func (cu *CreditUpdate) SetReplenishedAt(t time.Time) *CreditUpdate {
	cu.mutation.SetReplenishedAt(t)
	return cu
}

// SetNillableReplenishedAt sets the "replenished_at" field if the given value is not nil.
func (cu *CreditUpdate) SetNillableReplenishedAt(t *time.Time) *CreditUpdate {
	if t != nil {
		cu.SetReplenishedAt(*t)
	}
	return cu
}

// SetUserID sets the "user_id" field.
func (cu *CreditUpdate) SetUserID(u uuid.UUID) *CreditUpdate {
	cu.mutation.SetUserID(u)
	return cu
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (cu *CreditUpdate) SetNillableUserID(u *uuid.UUID) *CreditUpdate {
	if u != nil {
		cu.SetUserID(*u)
	}
	return cu
}

// SetCreditTypeID sets the "credit_type_id" field.
func (cu *CreditUpdate) SetCreditTypeID(u uuid.UUID) *CreditUpdate {
	cu.mutation.SetCreditTypeID(u)
	return cu
}

// SetNillableCreditTypeID sets the "credit_type_id" field if the given value is not nil.
func (cu *CreditUpdate) SetNillableCreditTypeID(u *uuid.UUID) *CreditUpdate {
	if u != nil {
		cu.SetCreditTypeID(*u)
	}
	return cu
}

// SetUpdatedAt sets the "updated_at" field.
func (cu *CreditUpdate) SetUpdatedAt(t time.Time) *CreditUpdate {
	cu.mutation.SetUpdatedAt(t)
	return cu
}

// SetUsersID sets the "users" edge to the User entity by ID.
func (cu *CreditUpdate) SetUsersID(id uuid.UUID) *CreditUpdate {
	cu.mutation.SetUsersID(id)
	return cu
}

// SetUsers sets the "users" edge to the User entity.
func (cu *CreditUpdate) SetUsers(u *User) *CreditUpdate {
	return cu.SetUsersID(u.ID)
}

// SetCreditType sets the "credit_type" edge to the CreditType entity.
func (cu *CreditUpdate) SetCreditType(c *CreditType) *CreditUpdate {
	return cu.SetCreditTypeID(c.ID)
}

// Mutation returns the CreditMutation object of the builder.
func (cu *CreditUpdate) Mutation() *CreditMutation {
	return cu.mutation
}

// ClearUsers clears the "users" edge to the User entity.
func (cu *CreditUpdate) ClearUsers() *CreditUpdate {
	cu.mutation.ClearUsers()
	return cu
}

// ClearCreditType clears the "credit_type" edge to the CreditType entity.
func (cu *CreditUpdate) ClearCreditType() *CreditUpdate {
	cu.mutation.ClearCreditType()
	return cu
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (cu *CreditUpdate) Save(ctx context.Context) (int, error) {
	cu.defaults()
	return withHooks(ctx, cu.sqlSave, cu.mutation, cu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (cu *CreditUpdate) SaveX(ctx context.Context) int {
	affected, err := cu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (cu *CreditUpdate) Exec(ctx context.Context) error {
	_, err := cu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (cu *CreditUpdate) ExecX(ctx context.Context) {
	if err := cu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (cu *CreditUpdate) defaults() {
	if _, ok := cu.mutation.UpdatedAt(); !ok {
		v := credit.UpdateDefaultUpdatedAt()
		cu.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (cu *CreditUpdate) check() error {
	if cu.mutation.UsersCleared() && len(cu.mutation.UsersIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Credit.users"`)
	}
	if cu.mutation.CreditTypeCleared() && len(cu.mutation.CreditTypeIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Credit.credit_type"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (cu *CreditUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *CreditUpdate {
	cu.modifiers = append(cu.modifiers, modifiers...)
	return cu
}

func (cu *CreditUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := cu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(credit.Table, credit.Columns, sqlgraph.NewFieldSpec(credit.FieldID, field.TypeUUID))
	if ps := cu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := cu.mutation.RemainingAmount(); ok {
		_spec.SetField(credit.FieldRemainingAmount, field.TypeInt32, value)
	}
	if value, ok := cu.mutation.AddedRemainingAmount(); ok {
		_spec.AddField(credit.FieldRemainingAmount, field.TypeInt32, value)
	}
	if value, ok := cu.mutation.ExpiresAt(); ok {
		_spec.SetField(credit.FieldExpiresAt, field.TypeTime, value)
	}
	if value, ok := cu.mutation.StripeLineItemID(); ok {
		_spec.SetField(credit.FieldStripeLineItemID, field.TypeString, value)
	}
	if cu.mutation.StripeLineItemIDCleared() {
		_spec.ClearField(credit.FieldStripeLineItemID, field.TypeString)
	}
	if value, ok := cu.mutation.ReplenishedAt(); ok {
		_spec.SetField(credit.FieldReplenishedAt, field.TypeTime, value)
	}
	if value, ok := cu.mutation.UpdatedAt(); ok {
		_spec.SetField(credit.FieldUpdatedAt, field.TypeTime, value)
	}
	if cu.mutation.UsersCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := cu.mutation.UsersIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if cu.mutation.CreditTypeCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := cu.mutation.CreditTypeIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(cu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, cu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{credit.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	cu.mutation.done = true
	return n, nil
}

// CreditUpdateOne is the builder for updating a single Credit entity.
type CreditUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *CreditMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetRemainingAmount sets the "remaining_amount" field.
func (cuo *CreditUpdateOne) SetRemainingAmount(i int32) *CreditUpdateOne {
	cuo.mutation.ResetRemainingAmount()
	cuo.mutation.SetRemainingAmount(i)
	return cuo
}

// SetNillableRemainingAmount sets the "remaining_amount" field if the given value is not nil.
func (cuo *CreditUpdateOne) SetNillableRemainingAmount(i *int32) *CreditUpdateOne {
	if i != nil {
		cuo.SetRemainingAmount(*i)
	}
	return cuo
}

// AddRemainingAmount adds i to the "remaining_amount" field.
func (cuo *CreditUpdateOne) AddRemainingAmount(i int32) *CreditUpdateOne {
	cuo.mutation.AddRemainingAmount(i)
	return cuo
}

// SetExpiresAt sets the "expires_at" field.
func (cuo *CreditUpdateOne) SetExpiresAt(t time.Time) *CreditUpdateOne {
	cuo.mutation.SetExpiresAt(t)
	return cuo
}

// SetNillableExpiresAt sets the "expires_at" field if the given value is not nil.
func (cuo *CreditUpdateOne) SetNillableExpiresAt(t *time.Time) *CreditUpdateOne {
	if t != nil {
		cuo.SetExpiresAt(*t)
	}
	return cuo
}

// SetStripeLineItemID sets the "stripe_line_item_id" field.
func (cuo *CreditUpdateOne) SetStripeLineItemID(s string) *CreditUpdateOne {
	cuo.mutation.SetStripeLineItemID(s)
	return cuo
}

// SetNillableStripeLineItemID sets the "stripe_line_item_id" field if the given value is not nil.
func (cuo *CreditUpdateOne) SetNillableStripeLineItemID(s *string) *CreditUpdateOne {
	if s != nil {
		cuo.SetStripeLineItemID(*s)
	}
	return cuo
}

// ClearStripeLineItemID clears the value of the "stripe_line_item_id" field.
func (cuo *CreditUpdateOne) ClearStripeLineItemID() *CreditUpdateOne {
	cuo.mutation.ClearStripeLineItemID()
	return cuo
}

// SetReplenishedAt sets the "replenished_at" field.
func (cuo *CreditUpdateOne) SetReplenishedAt(t time.Time) *CreditUpdateOne {
	cuo.mutation.SetReplenishedAt(t)
	return cuo
}

// SetNillableReplenishedAt sets the "replenished_at" field if the given value is not nil.
func (cuo *CreditUpdateOne) SetNillableReplenishedAt(t *time.Time) *CreditUpdateOne {
	if t != nil {
		cuo.SetReplenishedAt(*t)
	}
	return cuo
}

// SetUserID sets the "user_id" field.
func (cuo *CreditUpdateOne) SetUserID(u uuid.UUID) *CreditUpdateOne {
	cuo.mutation.SetUserID(u)
	return cuo
}

// SetNillableUserID sets the "user_id" field if the given value is not nil.
func (cuo *CreditUpdateOne) SetNillableUserID(u *uuid.UUID) *CreditUpdateOne {
	if u != nil {
		cuo.SetUserID(*u)
	}
	return cuo
}

// SetCreditTypeID sets the "credit_type_id" field.
func (cuo *CreditUpdateOne) SetCreditTypeID(u uuid.UUID) *CreditUpdateOne {
	cuo.mutation.SetCreditTypeID(u)
	return cuo
}

// SetNillableCreditTypeID sets the "credit_type_id" field if the given value is not nil.
func (cuo *CreditUpdateOne) SetNillableCreditTypeID(u *uuid.UUID) *CreditUpdateOne {
	if u != nil {
		cuo.SetCreditTypeID(*u)
	}
	return cuo
}

// SetUpdatedAt sets the "updated_at" field.
func (cuo *CreditUpdateOne) SetUpdatedAt(t time.Time) *CreditUpdateOne {
	cuo.mutation.SetUpdatedAt(t)
	return cuo
}

// SetUsersID sets the "users" edge to the User entity by ID.
func (cuo *CreditUpdateOne) SetUsersID(id uuid.UUID) *CreditUpdateOne {
	cuo.mutation.SetUsersID(id)
	return cuo
}

// SetUsers sets the "users" edge to the User entity.
func (cuo *CreditUpdateOne) SetUsers(u *User) *CreditUpdateOne {
	return cuo.SetUsersID(u.ID)
}

// SetCreditType sets the "credit_type" edge to the CreditType entity.
func (cuo *CreditUpdateOne) SetCreditType(c *CreditType) *CreditUpdateOne {
	return cuo.SetCreditTypeID(c.ID)
}

// Mutation returns the CreditMutation object of the builder.
func (cuo *CreditUpdateOne) Mutation() *CreditMutation {
	return cuo.mutation
}

// ClearUsers clears the "users" edge to the User entity.
func (cuo *CreditUpdateOne) ClearUsers() *CreditUpdateOne {
	cuo.mutation.ClearUsers()
	return cuo
}

// ClearCreditType clears the "credit_type" edge to the CreditType entity.
func (cuo *CreditUpdateOne) ClearCreditType() *CreditUpdateOne {
	cuo.mutation.ClearCreditType()
	return cuo
}

// Where appends a list predicates to the CreditUpdate builder.
func (cuo *CreditUpdateOne) Where(ps ...predicate.Credit) *CreditUpdateOne {
	cuo.mutation.Where(ps...)
	return cuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (cuo *CreditUpdateOne) Select(field string, fields ...string) *CreditUpdateOne {
	cuo.fields = append([]string{field}, fields...)
	return cuo
}

// Save executes the query and returns the updated Credit entity.
func (cuo *CreditUpdateOne) Save(ctx context.Context) (*Credit, error) {
	cuo.defaults()
	return withHooks(ctx, cuo.sqlSave, cuo.mutation, cuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (cuo *CreditUpdateOne) SaveX(ctx context.Context) *Credit {
	node, err := cuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (cuo *CreditUpdateOne) Exec(ctx context.Context) error {
	_, err := cuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (cuo *CreditUpdateOne) ExecX(ctx context.Context) {
	if err := cuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (cuo *CreditUpdateOne) defaults() {
	if _, ok := cuo.mutation.UpdatedAt(); !ok {
		v := credit.UpdateDefaultUpdatedAt()
		cuo.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (cuo *CreditUpdateOne) check() error {
	if cuo.mutation.UsersCleared() && len(cuo.mutation.UsersIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Credit.users"`)
	}
	if cuo.mutation.CreditTypeCleared() && len(cuo.mutation.CreditTypeIDs()) > 0 {
		return errors.New(`ent: clearing a required unique edge "Credit.credit_type"`)
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (cuo *CreditUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *CreditUpdateOne {
	cuo.modifiers = append(cuo.modifiers, modifiers...)
	return cuo
}

func (cuo *CreditUpdateOne) sqlSave(ctx context.Context) (_node *Credit, err error) {
	if err := cuo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(credit.Table, credit.Columns, sqlgraph.NewFieldSpec(credit.FieldID, field.TypeUUID))
	id, ok := cuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Credit.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := cuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, credit.FieldID)
		for _, f := range fields {
			if !credit.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != credit.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := cuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := cuo.mutation.RemainingAmount(); ok {
		_spec.SetField(credit.FieldRemainingAmount, field.TypeInt32, value)
	}
	if value, ok := cuo.mutation.AddedRemainingAmount(); ok {
		_spec.AddField(credit.FieldRemainingAmount, field.TypeInt32, value)
	}
	if value, ok := cuo.mutation.ExpiresAt(); ok {
		_spec.SetField(credit.FieldExpiresAt, field.TypeTime, value)
	}
	if value, ok := cuo.mutation.StripeLineItemID(); ok {
		_spec.SetField(credit.FieldStripeLineItemID, field.TypeString, value)
	}
	if cuo.mutation.StripeLineItemIDCleared() {
		_spec.ClearField(credit.FieldStripeLineItemID, field.TypeString)
	}
	if value, ok := cuo.mutation.ReplenishedAt(); ok {
		_spec.SetField(credit.FieldReplenishedAt, field.TypeTime, value)
	}
	if value, ok := cuo.mutation.UpdatedAt(); ok {
		_spec.SetField(credit.FieldUpdatedAt, field.TypeTime, value)
	}
	if cuo.mutation.UsersCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := cuo.mutation.UsersIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if cuo.mutation.CreditTypeCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := cuo.mutation.CreditTypeIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(cuo.modifiers...)
	_node = &Credit{config: cuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, cuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{credit.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	cuo.mutation.done = true
	return _node, nil
}
