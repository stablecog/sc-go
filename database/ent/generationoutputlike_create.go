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
	"github.com/stablecog/sc-go/database/ent/generationoutput"
	"github.com/stablecog/sc-go/database/ent/generationoutputlike"
	"github.com/stablecog/sc-go/database/ent/user"
)

// GenerationOutputLikeCreate is the builder for creating a GenerationOutputLike entity.
type GenerationOutputLikeCreate struct {
	config
	mutation *GenerationOutputLikeMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetOutputID sets the "output_id" field.
func (golc *GenerationOutputLikeCreate) SetOutputID(u uuid.UUID) *GenerationOutputLikeCreate {
	golc.mutation.SetOutputID(u)
	return golc
}

// SetLikedByUserID sets the "liked_by_user_id" field.
func (golc *GenerationOutputLikeCreate) SetLikedByUserID(u uuid.UUID) *GenerationOutputLikeCreate {
	golc.mutation.SetLikedByUserID(u)
	return golc
}

// SetCreatedAt sets the "created_at" field.
func (golc *GenerationOutputLikeCreate) SetCreatedAt(t time.Time) *GenerationOutputLikeCreate {
	golc.mutation.SetCreatedAt(t)
	return golc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (golc *GenerationOutputLikeCreate) SetNillableCreatedAt(t *time.Time) *GenerationOutputLikeCreate {
	if t != nil {
		golc.SetCreatedAt(*t)
	}
	return golc
}

// SetID sets the "id" field.
func (golc *GenerationOutputLikeCreate) SetID(u uuid.UUID) *GenerationOutputLikeCreate {
	golc.mutation.SetID(u)
	return golc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (golc *GenerationOutputLikeCreate) SetNillableID(u *uuid.UUID) *GenerationOutputLikeCreate {
	if u != nil {
		golc.SetID(*u)
	}
	return golc
}

// SetGenerationOutputsID sets the "generation_outputs" edge to the GenerationOutput entity by ID.
func (golc *GenerationOutputLikeCreate) SetGenerationOutputsID(id uuid.UUID) *GenerationOutputLikeCreate {
	golc.mutation.SetGenerationOutputsID(id)
	return golc
}

// SetGenerationOutputs sets the "generation_outputs" edge to the GenerationOutput entity.
func (golc *GenerationOutputLikeCreate) SetGenerationOutputs(g *GenerationOutput) *GenerationOutputLikeCreate {
	return golc.SetGenerationOutputsID(g.ID)
}

// SetUsersID sets the "users" edge to the User entity by ID.
func (golc *GenerationOutputLikeCreate) SetUsersID(id uuid.UUID) *GenerationOutputLikeCreate {
	golc.mutation.SetUsersID(id)
	return golc
}

// SetUsers sets the "users" edge to the User entity.
func (golc *GenerationOutputLikeCreate) SetUsers(u *User) *GenerationOutputLikeCreate {
	return golc.SetUsersID(u.ID)
}

// Mutation returns the GenerationOutputLikeMutation object of the builder.
func (golc *GenerationOutputLikeCreate) Mutation() *GenerationOutputLikeMutation {
	return golc.mutation
}

// Save creates the GenerationOutputLike in the database.
func (golc *GenerationOutputLikeCreate) Save(ctx context.Context) (*GenerationOutputLike, error) {
	golc.defaults()
	return withHooks(ctx, golc.sqlSave, golc.mutation, golc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (golc *GenerationOutputLikeCreate) SaveX(ctx context.Context) *GenerationOutputLike {
	v, err := golc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (golc *GenerationOutputLikeCreate) Exec(ctx context.Context) error {
	_, err := golc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (golc *GenerationOutputLikeCreate) ExecX(ctx context.Context) {
	if err := golc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (golc *GenerationOutputLikeCreate) defaults() {
	if _, ok := golc.mutation.CreatedAt(); !ok {
		v := generationoutputlike.DefaultCreatedAt()
		golc.mutation.SetCreatedAt(v)
	}
	if _, ok := golc.mutation.ID(); !ok {
		v := generationoutputlike.DefaultID()
		golc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (golc *GenerationOutputLikeCreate) check() error {
	if _, ok := golc.mutation.OutputID(); !ok {
		return &ValidationError{Name: "output_id", err: errors.New(`ent: missing required field "GenerationOutputLike.output_id"`)}
	}
	if _, ok := golc.mutation.LikedByUserID(); !ok {
		return &ValidationError{Name: "liked_by_user_id", err: errors.New(`ent: missing required field "GenerationOutputLike.liked_by_user_id"`)}
	}
	if _, ok := golc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "GenerationOutputLike.created_at"`)}
	}
	if len(golc.mutation.GenerationOutputsIDs()) == 0 {
		return &ValidationError{Name: "generation_outputs", err: errors.New(`ent: missing required edge "GenerationOutputLike.generation_outputs"`)}
	}
	if len(golc.mutation.UsersIDs()) == 0 {
		return &ValidationError{Name: "users", err: errors.New(`ent: missing required edge "GenerationOutputLike.users"`)}
	}
	return nil
}

func (golc *GenerationOutputLikeCreate) sqlSave(ctx context.Context) (*GenerationOutputLike, error) {
	if err := golc.check(); err != nil {
		return nil, err
	}
	_node, _spec := golc.createSpec()
	if err := sqlgraph.CreateNode(ctx, golc.driver, _spec); err != nil {
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
	golc.mutation.id = &_node.ID
	golc.mutation.done = true
	return _node, nil
}

func (golc *GenerationOutputLikeCreate) createSpec() (*GenerationOutputLike, *sqlgraph.CreateSpec) {
	var (
		_node = &GenerationOutputLike{config: golc.config}
		_spec = sqlgraph.NewCreateSpec(generationoutputlike.Table, sqlgraph.NewFieldSpec(generationoutputlike.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = golc.conflict
	if id, ok := golc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := golc.mutation.CreatedAt(); ok {
		_spec.SetField(generationoutputlike.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if nodes := golc.mutation.GenerationOutputsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   generationoutputlike.GenerationOutputsTable,
			Columns: []string{generationoutputlike.GenerationOutputsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(generationoutput.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.OutputID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := golc.mutation.UsersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   generationoutputlike.UsersTable,
			Columns: []string{generationoutputlike.UsersColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(user.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.LikedByUserID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.GenerationOutputLike.Create().
//		SetOutputID(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.GenerationOutputLikeUpsert) {
//			SetOutputID(v+v).
//		}).
//		Exec(ctx)
func (golc *GenerationOutputLikeCreate) OnConflict(opts ...sql.ConflictOption) *GenerationOutputLikeUpsertOne {
	golc.conflict = opts
	return &GenerationOutputLikeUpsertOne{
		create: golc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.GenerationOutputLike.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (golc *GenerationOutputLikeCreate) OnConflictColumns(columns ...string) *GenerationOutputLikeUpsertOne {
	golc.conflict = append(golc.conflict, sql.ConflictColumns(columns...))
	return &GenerationOutputLikeUpsertOne{
		create: golc,
	}
}

type (
	// GenerationOutputLikeUpsertOne is the builder for "upsert"-ing
	//  one GenerationOutputLike node.
	GenerationOutputLikeUpsertOne struct {
		create *GenerationOutputLikeCreate
	}

	// GenerationOutputLikeUpsert is the "OnConflict" setter.
	GenerationOutputLikeUpsert struct {
		*sql.UpdateSet
	}
)

// SetOutputID sets the "output_id" field.
func (u *GenerationOutputLikeUpsert) SetOutputID(v uuid.UUID) *GenerationOutputLikeUpsert {
	u.Set(generationoutputlike.FieldOutputID, v)
	return u
}

// UpdateOutputID sets the "output_id" field to the value that was provided on create.
func (u *GenerationOutputLikeUpsert) UpdateOutputID() *GenerationOutputLikeUpsert {
	u.SetExcluded(generationoutputlike.FieldOutputID)
	return u
}

// SetLikedByUserID sets the "liked_by_user_id" field.
func (u *GenerationOutputLikeUpsert) SetLikedByUserID(v uuid.UUID) *GenerationOutputLikeUpsert {
	u.Set(generationoutputlike.FieldLikedByUserID, v)
	return u
}

// UpdateLikedByUserID sets the "liked_by_user_id" field to the value that was provided on create.
func (u *GenerationOutputLikeUpsert) UpdateLikedByUserID() *GenerationOutputLikeUpsert {
	u.SetExcluded(generationoutputlike.FieldLikedByUserID)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.GenerationOutputLike.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(generationoutputlike.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *GenerationOutputLikeUpsertOne) UpdateNewValues() *GenerationOutputLikeUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(generationoutputlike.FieldID)
		}
		if _, exists := u.create.mutation.CreatedAt(); exists {
			s.SetIgnore(generationoutputlike.FieldCreatedAt)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.GenerationOutputLike.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *GenerationOutputLikeUpsertOne) Ignore() *GenerationOutputLikeUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *GenerationOutputLikeUpsertOne) DoNothing() *GenerationOutputLikeUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the GenerationOutputLikeCreate.OnConflict
// documentation for more info.
func (u *GenerationOutputLikeUpsertOne) Update(set func(*GenerationOutputLikeUpsert)) *GenerationOutputLikeUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&GenerationOutputLikeUpsert{UpdateSet: update})
	}))
	return u
}

// SetOutputID sets the "output_id" field.
func (u *GenerationOutputLikeUpsertOne) SetOutputID(v uuid.UUID) *GenerationOutputLikeUpsertOne {
	return u.Update(func(s *GenerationOutputLikeUpsert) {
		s.SetOutputID(v)
	})
}

// UpdateOutputID sets the "output_id" field to the value that was provided on create.
func (u *GenerationOutputLikeUpsertOne) UpdateOutputID() *GenerationOutputLikeUpsertOne {
	return u.Update(func(s *GenerationOutputLikeUpsert) {
		s.UpdateOutputID()
	})
}

// SetLikedByUserID sets the "liked_by_user_id" field.
func (u *GenerationOutputLikeUpsertOne) SetLikedByUserID(v uuid.UUID) *GenerationOutputLikeUpsertOne {
	return u.Update(func(s *GenerationOutputLikeUpsert) {
		s.SetLikedByUserID(v)
	})
}

// UpdateLikedByUserID sets the "liked_by_user_id" field to the value that was provided on create.
func (u *GenerationOutputLikeUpsertOne) UpdateLikedByUserID() *GenerationOutputLikeUpsertOne {
	return u.Update(func(s *GenerationOutputLikeUpsert) {
		s.UpdateLikedByUserID()
	})
}

// Exec executes the query.
func (u *GenerationOutputLikeUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for GenerationOutputLikeCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *GenerationOutputLikeUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *GenerationOutputLikeUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: GenerationOutputLikeUpsertOne.ID is not supported by MySQL driver. Use GenerationOutputLikeUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *GenerationOutputLikeUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// GenerationOutputLikeCreateBulk is the builder for creating many GenerationOutputLike entities in bulk.
type GenerationOutputLikeCreateBulk struct {
	config
	err      error
	builders []*GenerationOutputLikeCreate
	conflict []sql.ConflictOption
}

// Save creates the GenerationOutputLike entities in the database.
func (golcb *GenerationOutputLikeCreateBulk) Save(ctx context.Context) ([]*GenerationOutputLike, error) {
	if golcb.err != nil {
		return nil, golcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(golcb.builders))
	nodes := make([]*GenerationOutputLike, len(golcb.builders))
	mutators := make([]Mutator, len(golcb.builders))
	for i := range golcb.builders {
		func(i int, root context.Context) {
			builder := golcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*GenerationOutputLikeMutation)
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
					_, err = mutators[i+1].Mutate(root, golcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = golcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, golcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, golcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (golcb *GenerationOutputLikeCreateBulk) SaveX(ctx context.Context) []*GenerationOutputLike {
	v, err := golcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (golcb *GenerationOutputLikeCreateBulk) Exec(ctx context.Context) error {
	_, err := golcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (golcb *GenerationOutputLikeCreateBulk) ExecX(ctx context.Context) {
	if err := golcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.GenerationOutputLike.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.GenerationOutputLikeUpsert) {
//			SetOutputID(v+v).
//		}).
//		Exec(ctx)
func (golcb *GenerationOutputLikeCreateBulk) OnConflict(opts ...sql.ConflictOption) *GenerationOutputLikeUpsertBulk {
	golcb.conflict = opts
	return &GenerationOutputLikeUpsertBulk{
		create: golcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.GenerationOutputLike.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (golcb *GenerationOutputLikeCreateBulk) OnConflictColumns(columns ...string) *GenerationOutputLikeUpsertBulk {
	golcb.conflict = append(golcb.conflict, sql.ConflictColumns(columns...))
	return &GenerationOutputLikeUpsertBulk{
		create: golcb,
	}
}

// GenerationOutputLikeUpsertBulk is the builder for "upsert"-ing
// a bulk of GenerationOutputLike nodes.
type GenerationOutputLikeUpsertBulk struct {
	create *GenerationOutputLikeCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.GenerationOutputLike.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(generationoutputlike.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *GenerationOutputLikeUpsertBulk) UpdateNewValues() *GenerationOutputLikeUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(generationoutputlike.FieldID)
			}
			if _, exists := b.mutation.CreatedAt(); exists {
				s.SetIgnore(generationoutputlike.FieldCreatedAt)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.GenerationOutputLike.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *GenerationOutputLikeUpsertBulk) Ignore() *GenerationOutputLikeUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *GenerationOutputLikeUpsertBulk) DoNothing() *GenerationOutputLikeUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the GenerationOutputLikeCreateBulk.OnConflict
// documentation for more info.
func (u *GenerationOutputLikeUpsertBulk) Update(set func(*GenerationOutputLikeUpsert)) *GenerationOutputLikeUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&GenerationOutputLikeUpsert{UpdateSet: update})
	}))
	return u
}

// SetOutputID sets the "output_id" field.
func (u *GenerationOutputLikeUpsertBulk) SetOutputID(v uuid.UUID) *GenerationOutputLikeUpsertBulk {
	return u.Update(func(s *GenerationOutputLikeUpsert) {
		s.SetOutputID(v)
	})
}

// UpdateOutputID sets the "output_id" field to the value that was provided on create.
func (u *GenerationOutputLikeUpsertBulk) UpdateOutputID() *GenerationOutputLikeUpsertBulk {
	return u.Update(func(s *GenerationOutputLikeUpsert) {
		s.UpdateOutputID()
	})
}

// SetLikedByUserID sets the "liked_by_user_id" field.
func (u *GenerationOutputLikeUpsertBulk) SetLikedByUserID(v uuid.UUID) *GenerationOutputLikeUpsertBulk {
	return u.Update(func(s *GenerationOutputLikeUpsert) {
		s.SetLikedByUserID(v)
	})
}

// UpdateLikedByUserID sets the "liked_by_user_id" field to the value that was provided on create.
func (u *GenerationOutputLikeUpsertBulk) UpdateLikedByUserID() *GenerationOutputLikeUpsertBulk {
	return u.Update(func(s *GenerationOutputLikeUpsert) {
		s.UpdateLikedByUserID()
	})
}

// Exec executes the query.
func (u *GenerationOutputLikeUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the GenerationOutputLikeCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for GenerationOutputLikeCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *GenerationOutputLikeUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
