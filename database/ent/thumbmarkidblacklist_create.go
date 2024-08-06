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
	"github.com/stablecog/sc-go/database/ent/thumbmarkidblacklist"
)

// ThumbmarkIdBlackListCreate is the builder for creating a ThumbmarkIdBlackList entity.
type ThumbmarkIdBlackListCreate struct {
	config
	mutation *ThumbmarkIdBlackListMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetThumbmarkID sets the "thumbmark_id" field.
func (tiblc *ThumbmarkIdBlackListCreate) SetThumbmarkID(s string) *ThumbmarkIdBlackListCreate {
	tiblc.mutation.SetThumbmarkID(s)
	return tiblc
}

// SetCreatedAt sets the "created_at" field.
func (tiblc *ThumbmarkIdBlackListCreate) SetCreatedAt(t time.Time) *ThumbmarkIdBlackListCreate {
	tiblc.mutation.SetCreatedAt(t)
	return tiblc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (tiblc *ThumbmarkIdBlackListCreate) SetNillableCreatedAt(t *time.Time) *ThumbmarkIdBlackListCreate {
	if t != nil {
		tiblc.SetCreatedAt(*t)
	}
	return tiblc
}

// SetUpdatedAt sets the "updated_at" field.
func (tiblc *ThumbmarkIdBlackListCreate) SetUpdatedAt(t time.Time) *ThumbmarkIdBlackListCreate {
	tiblc.mutation.SetUpdatedAt(t)
	return tiblc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (tiblc *ThumbmarkIdBlackListCreate) SetNillableUpdatedAt(t *time.Time) *ThumbmarkIdBlackListCreate {
	if t != nil {
		tiblc.SetUpdatedAt(*t)
	}
	return tiblc
}

// SetID sets the "id" field.
func (tiblc *ThumbmarkIdBlackListCreate) SetID(u uuid.UUID) *ThumbmarkIdBlackListCreate {
	tiblc.mutation.SetID(u)
	return tiblc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (tiblc *ThumbmarkIdBlackListCreate) SetNillableID(u *uuid.UUID) *ThumbmarkIdBlackListCreate {
	if u != nil {
		tiblc.SetID(*u)
	}
	return tiblc
}

// Mutation returns the ThumbmarkIdBlackListMutation object of the builder.
func (tiblc *ThumbmarkIdBlackListCreate) Mutation() *ThumbmarkIdBlackListMutation {
	return tiblc.mutation
}

// Save creates the ThumbmarkIdBlackList in the database.
func (tiblc *ThumbmarkIdBlackListCreate) Save(ctx context.Context) (*ThumbmarkIdBlackList, error) {
	tiblc.defaults()
	return withHooks(ctx, tiblc.sqlSave, tiblc.mutation, tiblc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (tiblc *ThumbmarkIdBlackListCreate) SaveX(ctx context.Context) *ThumbmarkIdBlackList {
	v, err := tiblc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tiblc *ThumbmarkIdBlackListCreate) Exec(ctx context.Context) error {
	_, err := tiblc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tiblc *ThumbmarkIdBlackListCreate) ExecX(ctx context.Context) {
	if err := tiblc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (tiblc *ThumbmarkIdBlackListCreate) defaults() {
	if _, ok := tiblc.mutation.CreatedAt(); !ok {
		v := thumbmarkidblacklist.DefaultCreatedAt()
		tiblc.mutation.SetCreatedAt(v)
	}
	if _, ok := tiblc.mutation.UpdatedAt(); !ok {
		v := thumbmarkidblacklist.DefaultUpdatedAt()
		tiblc.mutation.SetUpdatedAt(v)
	}
	if _, ok := tiblc.mutation.ID(); !ok {
		v := thumbmarkidblacklist.DefaultID()
		tiblc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (tiblc *ThumbmarkIdBlackListCreate) check() error {
	if _, ok := tiblc.mutation.ThumbmarkID(); !ok {
		return &ValidationError{Name: "thumbmark_id", err: errors.New(`ent: missing required field "ThumbmarkIdBlackList.thumbmark_id"`)}
	}
	if _, ok := tiblc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "ThumbmarkIdBlackList.created_at"`)}
	}
	if _, ok := tiblc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "ThumbmarkIdBlackList.updated_at"`)}
	}
	return nil
}

func (tiblc *ThumbmarkIdBlackListCreate) sqlSave(ctx context.Context) (*ThumbmarkIdBlackList, error) {
	if err := tiblc.check(); err != nil {
		return nil, err
	}
	_node, _spec := tiblc.createSpec()
	if err := sqlgraph.CreateNode(ctx, tiblc.driver, _spec); err != nil {
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
	tiblc.mutation.id = &_node.ID
	tiblc.mutation.done = true
	return _node, nil
}

func (tiblc *ThumbmarkIdBlackListCreate) createSpec() (*ThumbmarkIdBlackList, *sqlgraph.CreateSpec) {
	var (
		_node = &ThumbmarkIdBlackList{config: tiblc.config}
		_spec = sqlgraph.NewCreateSpec(thumbmarkidblacklist.Table, sqlgraph.NewFieldSpec(thumbmarkidblacklist.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = tiblc.conflict
	if id, ok := tiblc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := tiblc.mutation.ThumbmarkID(); ok {
		_spec.SetField(thumbmarkidblacklist.FieldThumbmarkID, field.TypeString, value)
		_node.ThumbmarkID = value
	}
	if value, ok := tiblc.mutation.CreatedAt(); ok {
		_spec.SetField(thumbmarkidblacklist.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := tiblc.mutation.UpdatedAt(); ok {
		_spec.SetField(thumbmarkidblacklist.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.ThumbmarkIdBlackList.Create().
//		SetThumbmarkID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.ThumbmarkIdBlackListUpsert) {
//			SetThumbmarkID(v+v).
//		}).
//		Exec(ctx)
func (tiblc *ThumbmarkIdBlackListCreate) OnConflict(opts ...sql.ConflictOption) *ThumbmarkIdBlackListUpsertOne {
	tiblc.conflict = opts
	return &ThumbmarkIdBlackListUpsertOne{
		create: tiblc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.ThumbmarkIdBlackList.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (tiblc *ThumbmarkIdBlackListCreate) OnConflictColumns(columns ...string) *ThumbmarkIdBlackListUpsertOne {
	tiblc.conflict = append(tiblc.conflict, sql.ConflictColumns(columns...))
	return &ThumbmarkIdBlackListUpsertOne{
		create: tiblc,
	}
}

type (
	// ThumbmarkIdBlackListUpsertOne is the builder for "upsert"-ing
	//  one ThumbmarkIdBlackList node.
	ThumbmarkIdBlackListUpsertOne struct {
		create *ThumbmarkIdBlackListCreate
	}

	// ThumbmarkIdBlackListUpsert is the "OnConflict" setter.
	ThumbmarkIdBlackListUpsert struct {
		*sql.UpdateSet
	}
)

// SetThumbmarkID sets the "thumbmark_id" field.
func (u *ThumbmarkIdBlackListUpsert) SetThumbmarkID(v string) *ThumbmarkIdBlackListUpsert {
	u.Set(thumbmarkidblacklist.FieldThumbmarkID, v)
	return u
}

// UpdateThumbmarkID sets the "thumbmark_id" field to the value that was provided on create.
func (u *ThumbmarkIdBlackListUpsert) UpdateThumbmarkID() *ThumbmarkIdBlackListUpsert {
	u.SetExcluded(thumbmarkidblacklist.FieldThumbmarkID)
	return u
}

// SetUpdatedAt sets the "updated_at" field.
func (u *ThumbmarkIdBlackListUpsert) SetUpdatedAt(v time.Time) *ThumbmarkIdBlackListUpsert {
	u.Set(thumbmarkidblacklist.FieldUpdatedAt, v)
	return u
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *ThumbmarkIdBlackListUpsert) UpdateUpdatedAt() *ThumbmarkIdBlackListUpsert {
	u.SetExcluded(thumbmarkidblacklist.FieldUpdatedAt)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.ThumbmarkIdBlackList.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(thumbmarkidblacklist.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *ThumbmarkIdBlackListUpsertOne) UpdateNewValues() *ThumbmarkIdBlackListUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(thumbmarkidblacklist.FieldID)
		}
		if _, exists := u.create.mutation.CreatedAt(); exists {
			s.SetIgnore(thumbmarkidblacklist.FieldCreatedAt)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.ThumbmarkIdBlackList.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *ThumbmarkIdBlackListUpsertOne) Ignore() *ThumbmarkIdBlackListUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *ThumbmarkIdBlackListUpsertOne) DoNothing() *ThumbmarkIdBlackListUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the ThumbmarkIdBlackListCreate.OnConflict
// documentation for more info.
func (u *ThumbmarkIdBlackListUpsertOne) Update(set func(*ThumbmarkIdBlackListUpsert)) *ThumbmarkIdBlackListUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&ThumbmarkIdBlackListUpsert{UpdateSet: update})
	}))
	return u
}

// SetThumbmarkID sets the "thumbmark_id" field.
func (u *ThumbmarkIdBlackListUpsertOne) SetThumbmarkID(v string) *ThumbmarkIdBlackListUpsertOne {
	return u.Update(func(s *ThumbmarkIdBlackListUpsert) {
		s.SetThumbmarkID(v)
	})
}

// UpdateThumbmarkID sets the "thumbmark_id" field to the value that was provided on create.
func (u *ThumbmarkIdBlackListUpsertOne) UpdateThumbmarkID() *ThumbmarkIdBlackListUpsertOne {
	return u.Update(func(s *ThumbmarkIdBlackListUpsert) {
		s.UpdateThumbmarkID()
	})
}

// SetUpdatedAt sets the "updated_at" field.
func (u *ThumbmarkIdBlackListUpsertOne) SetUpdatedAt(v time.Time) *ThumbmarkIdBlackListUpsertOne {
	return u.Update(func(s *ThumbmarkIdBlackListUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *ThumbmarkIdBlackListUpsertOne) UpdateUpdatedAt() *ThumbmarkIdBlackListUpsertOne {
	return u.Update(func(s *ThumbmarkIdBlackListUpsert) {
		s.UpdateUpdatedAt()
	})
}

// Exec executes the query.
func (u *ThumbmarkIdBlackListUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for ThumbmarkIdBlackListCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *ThumbmarkIdBlackListUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *ThumbmarkIdBlackListUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: ThumbmarkIdBlackListUpsertOne.ID is not supported by MySQL driver. Use ThumbmarkIdBlackListUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *ThumbmarkIdBlackListUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// ThumbmarkIdBlackListCreateBulk is the builder for creating many ThumbmarkIdBlackList entities in bulk.
type ThumbmarkIdBlackListCreateBulk struct {
	config
	err      error
	builders []*ThumbmarkIdBlackListCreate
	conflict []sql.ConflictOption
}

// Save creates the ThumbmarkIdBlackList entities in the database.
func (tiblcb *ThumbmarkIdBlackListCreateBulk) Save(ctx context.Context) ([]*ThumbmarkIdBlackList, error) {
	if tiblcb.err != nil {
		return nil, tiblcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(tiblcb.builders))
	nodes := make([]*ThumbmarkIdBlackList, len(tiblcb.builders))
	mutators := make([]Mutator, len(tiblcb.builders))
	for i := range tiblcb.builders {
		func(i int, root context.Context) {
			builder := tiblcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*ThumbmarkIdBlackListMutation)
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
					_, err = mutators[i+1].Mutate(root, tiblcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = tiblcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, tiblcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, tiblcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (tiblcb *ThumbmarkIdBlackListCreateBulk) SaveX(ctx context.Context) []*ThumbmarkIdBlackList {
	v, err := tiblcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (tiblcb *ThumbmarkIdBlackListCreateBulk) Exec(ctx context.Context) error {
	_, err := tiblcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (tiblcb *ThumbmarkIdBlackListCreateBulk) ExecX(ctx context.Context) {
	if err := tiblcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.ThumbmarkIdBlackList.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.ThumbmarkIdBlackListUpsert) {
//			SetThumbmarkID(v+v).
//		}).
//		Exec(ctx)
func (tiblcb *ThumbmarkIdBlackListCreateBulk) OnConflict(opts ...sql.ConflictOption) *ThumbmarkIdBlackListUpsertBulk {
	tiblcb.conflict = opts
	return &ThumbmarkIdBlackListUpsertBulk{
		create: tiblcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.ThumbmarkIdBlackList.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (tiblcb *ThumbmarkIdBlackListCreateBulk) OnConflictColumns(columns ...string) *ThumbmarkIdBlackListUpsertBulk {
	tiblcb.conflict = append(tiblcb.conflict, sql.ConflictColumns(columns...))
	return &ThumbmarkIdBlackListUpsertBulk{
		create: tiblcb,
	}
}

// ThumbmarkIdBlackListUpsertBulk is the builder for "upsert"-ing
// a bulk of ThumbmarkIdBlackList nodes.
type ThumbmarkIdBlackListUpsertBulk struct {
	create *ThumbmarkIdBlackListCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.ThumbmarkIdBlackList.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(thumbmarkidblacklist.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *ThumbmarkIdBlackListUpsertBulk) UpdateNewValues() *ThumbmarkIdBlackListUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(thumbmarkidblacklist.FieldID)
			}
			if _, exists := b.mutation.CreatedAt(); exists {
				s.SetIgnore(thumbmarkidblacklist.FieldCreatedAt)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.ThumbmarkIdBlackList.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *ThumbmarkIdBlackListUpsertBulk) Ignore() *ThumbmarkIdBlackListUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *ThumbmarkIdBlackListUpsertBulk) DoNothing() *ThumbmarkIdBlackListUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the ThumbmarkIdBlackListCreateBulk.OnConflict
// documentation for more info.
func (u *ThumbmarkIdBlackListUpsertBulk) Update(set func(*ThumbmarkIdBlackListUpsert)) *ThumbmarkIdBlackListUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&ThumbmarkIdBlackListUpsert{UpdateSet: update})
	}))
	return u
}

// SetThumbmarkID sets the "thumbmark_id" field.
func (u *ThumbmarkIdBlackListUpsertBulk) SetThumbmarkID(v string) *ThumbmarkIdBlackListUpsertBulk {
	return u.Update(func(s *ThumbmarkIdBlackListUpsert) {
		s.SetThumbmarkID(v)
	})
}

// UpdateThumbmarkID sets the "thumbmark_id" field to the value that was provided on create.
func (u *ThumbmarkIdBlackListUpsertBulk) UpdateThumbmarkID() *ThumbmarkIdBlackListUpsertBulk {
	return u.Update(func(s *ThumbmarkIdBlackListUpsert) {
		s.UpdateThumbmarkID()
	})
}

// SetUpdatedAt sets the "updated_at" field.
func (u *ThumbmarkIdBlackListUpsertBulk) SetUpdatedAt(v time.Time) *ThumbmarkIdBlackListUpsertBulk {
	return u.Update(func(s *ThumbmarkIdBlackListUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *ThumbmarkIdBlackListUpsertBulk) UpdateUpdatedAt() *ThumbmarkIdBlackListUpsertBulk {
	return u.Update(func(s *ThumbmarkIdBlackListUpsert) {
		s.UpdateUpdatedAt()
	})
}

// Exec executes the query.
func (u *ThumbmarkIdBlackListUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the ThumbmarkIdBlackListCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for ThumbmarkIdBlackListCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *ThumbmarkIdBlackListUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
