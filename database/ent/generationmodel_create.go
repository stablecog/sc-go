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
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationmodel"
	"github.com/stablecog/sc-go/database/ent/scheduler"
)

// GenerationModelCreate is the builder for creating a GenerationModel entity.
type GenerationModelCreate struct {
	config
	mutation *GenerationModelMutation
	hooks    []Hook
	conflict []sql.ConflictOption
}

// SetNameInWorker sets the "name_in_worker" field.
func (gmc *GenerationModelCreate) SetNameInWorker(s string) *GenerationModelCreate {
	gmc.mutation.SetNameInWorker(s)
	return gmc
}

// SetIsActive sets the "is_active" field.
func (gmc *GenerationModelCreate) SetIsActive(b bool) *GenerationModelCreate {
	gmc.mutation.SetIsActive(b)
	return gmc
}

// SetNillableIsActive sets the "is_active" field if the given value is not nil.
func (gmc *GenerationModelCreate) SetNillableIsActive(b *bool) *GenerationModelCreate {
	if b != nil {
		gmc.SetIsActive(*b)
	}
	return gmc
}

// SetIsDefault sets the "is_default" field.
func (gmc *GenerationModelCreate) SetIsDefault(b bool) *GenerationModelCreate {
	gmc.mutation.SetIsDefault(b)
	return gmc
}

// SetNillableIsDefault sets the "is_default" field if the given value is not nil.
func (gmc *GenerationModelCreate) SetNillableIsDefault(b *bool) *GenerationModelCreate {
	if b != nil {
		gmc.SetIsDefault(*b)
	}
	return gmc
}

// SetIsHidden sets the "is_hidden" field.
func (gmc *GenerationModelCreate) SetIsHidden(b bool) *GenerationModelCreate {
	gmc.mutation.SetIsHidden(b)
	return gmc
}

// SetNillableIsHidden sets the "is_hidden" field if the given value is not nil.
func (gmc *GenerationModelCreate) SetNillableIsHidden(b *bool) *GenerationModelCreate {
	if b != nil {
		gmc.SetIsHidden(*b)
	}
	return gmc
}

// SetDisplayWeight sets the "display_weight" field.
func (gmc *GenerationModelCreate) SetDisplayWeight(i int32) *GenerationModelCreate {
	gmc.mutation.SetDisplayWeight(i)
	return gmc
}

// SetNillableDisplayWeight sets the "display_weight" field if the given value is not nil.
func (gmc *GenerationModelCreate) SetNillableDisplayWeight(i *int32) *GenerationModelCreate {
	if i != nil {
		gmc.SetDisplayWeight(*i)
	}
	return gmc
}

// SetDefaultSchedulerID sets the "default_scheduler_id" field.
func (gmc *GenerationModelCreate) SetDefaultSchedulerID(u uuid.UUID) *GenerationModelCreate {
	gmc.mutation.SetDefaultSchedulerID(u)
	return gmc
}

// SetNillableDefaultSchedulerID sets the "default_scheduler_id" field if the given value is not nil.
func (gmc *GenerationModelCreate) SetNillableDefaultSchedulerID(u *uuid.UUID) *GenerationModelCreate {
	if u != nil {
		gmc.SetDefaultSchedulerID(*u)
	}
	return gmc
}

// SetDefaultWidth sets the "default_width" field.
func (gmc *GenerationModelCreate) SetDefaultWidth(i int32) *GenerationModelCreate {
	gmc.mutation.SetDefaultWidth(i)
	return gmc
}

// SetNillableDefaultWidth sets the "default_width" field if the given value is not nil.
func (gmc *GenerationModelCreate) SetNillableDefaultWidth(i *int32) *GenerationModelCreate {
	if i != nil {
		gmc.SetDefaultWidth(*i)
	}
	return gmc
}

// SetDefaultHeight sets the "default_height" field.
func (gmc *GenerationModelCreate) SetDefaultHeight(i int32) *GenerationModelCreate {
	gmc.mutation.SetDefaultHeight(i)
	return gmc
}

// SetNillableDefaultHeight sets the "default_height" field if the given value is not nil.
func (gmc *GenerationModelCreate) SetNillableDefaultHeight(i *int32) *GenerationModelCreate {
	if i != nil {
		gmc.SetDefaultHeight(*i)
	}
	return gmc
}

// SetCreatedAt sets the "created_at" field.
func (gmc *GenerationModelCreate) SetCreatedAt(t time.Time) *GenerationModelCreate {
	gmc.mutation.SetCreatedAt(t)
	return gmc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (gmc *GenerationModelCreate) SetNillableCreatedAt(t *time.Time) *GenerationModelCreate {
	if t != nil {
		gmc.SetCreatedAt(*t)
	}
	return gmc
}

// SetUpdatedAt sets the "updated_at" field.
func (gmc *GenerationModelCreate) SetUpdatedAt(t time.Time) *GenerationModelCreate {
	gmc.mutation.SetUpdatedAt(t)
	return gmc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (gmc *GenerationModelCreate) SetNillableUpdatedAt(t *time.Time) *GenerationModelCreate {
	if t != nil {
		gmc.SetUpdatedAt(*t)
	}
	return gmc
}

// SetID sets the "id" field.
func (gmc *GenerationModelCreate) SetID(u uuid.UUID) *GenerationModelCreate {
	gmc.mutation.SetID(u)
	return gmc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (gmc *GenerationModelCreate) SetNillableID(u *uuid.UUID) *GenerationModelCreate {
	if u != nil {
		gmc.SetID(*u)
	}
	return gmc
}

// AddGenerationIDs adds the "generations" edge to the Generation entity by IDs.
func (gmc *GenerationModelCreate) AddGenerationIDs(ids ...uuid.UUID) *GenerationModelCreate {
	gmc.mutation.AddGenerationIDs(ids...)
	return gmc
}

// AddGenerations adds the "generations" edges to the Generation entity.
func (gmc *GenerationModelCreate) AddGenerations(g ...*Generation) *GenerationModelCreate {
	ids := make([]uuid.UUID, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return gmc.AddGenerationIDs(ids...)
}

// AddSchedulerIDs adds the "schedulers" edge to the Scheduler entity by IDs.
func (gmc *GenerationModelCreate) AddSchedulerIDs(ids ...uuid.UUID) *GenerationModelCreate {
	gmc.mutation.AddSchedulerIDs(ids...)
	return gmc
}

// AddSchedulers adds the "schedulers" edges to the Scheduler entity.
func (gmc *GenerationModelCreate) AddSchedulers(s ...*Scheduler) *GenerationModelCreate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return gmc.AddSchedulerIDs(ids...)
}

// Mutation returns the GenerationModelMutation object of the builder.
func (gmc *GenerationModelCreate) Mutation() *GenerationModelMutation {
	return gmc.mutation
}

// Save creates the GenerationModel in the database.
func (gmc *GenerationModelCreate) Save(ctx context.Context) (*GenerationModel, error) {
	gmc.defaults()
	return withHooks(ctx, gmc.sqlSave, gmc.mutation, gmc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (gmc *GenerationModelCreate) SaveX(ctx context.Context) *GenerationModel {
	v, err := gmc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (gmc *GenerationModelCreate) Exec(ctx context.Context) error {
	_, err := gmc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (gmc *GenerationModelCreate) ExecX(ctx context.Context) {
	if err := gmc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (gmc *GenerationModelCreate) defaults() {
	if _, ok := gmc.mutation.IsActive(); !ok {
		v := generationmodel.DefaultIsActive
		gmc.mutation.SetIsActive(v)
	}
	if _, ok := gmc.mutation.IsDefault(); !ok {
		v := generationmodel.DefaultIsDefault
		gmc.mutation.SetIsDefault(v)
	}
	if _, ok := gmc.mutation.IsHidden(); !ok {
		v := generationmodel.DefaultIsHidden
		gmc.mutation.SetIsHidden(v)
	}
	if _, ok := gmc.mutation.DisplayWeight(); !ok {
		v := generationmodel.DefaultDisplayWeight
		gmc.mutation.SetDisplayWeight(v)
	}
	if _, ok := gmc.mutation.DefaultWidth(); !ok {
		v := generationmodel.DefaultDefaultWidth
		gmc.mutation.SetDefaultWidth(v)
	}
	if _, ok := gmc.mutation.DefaultHeight(); !ok {
		v := generationmodel.DefaultDefaultHeight
		gmc.mutation.SetDefaultHeight(v)
	}
	if _, ok := gmc.mutation.CreatedAt(); !ok {
		v := generationmodel.DefaultCreatedAt()
		gmc.mutation.SetCreatedAt(v)
	}
	if _, ok := gmc.mutation.UpdatedAt(); !ok {
		v := generationmodel.DefaultUpdatedAt()
		gmc.mutation.SetUpdatedAt(v)
	}
	if _, ok := gmc.mutation.ID(); !ok {
		v := generationmodel.DefaultID()
		gmc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (gmc *GenerationModelCreate) check() error {
	if _, ok := gmc.mutation.NameInWorker(); !ok {
		return &ValidationError{Name: "name_in_worker", err: errors.New(`ent: missing required field "GenerationModel.name_in_worker"`)}
	}
	if _, ok := gmc.mutation.IsActive(); !ok {
		return &ValidationError{Name: "is_active", err: errors.New(`ent: missing required field "GenerationModel.is_active"`)}
	}
	if _, ok := gmc.mutation.IsDefault(); !ok {
		return &ValidationError{Name: "is_default", err: errors.New(`ent: missing required field "GenerationModel.is_default"`)}
	}
	if _, ok := gmc.mutation.IsHidden(); !ok {
		return &ValidationError{Name: "is_hidden", err: errors.New(`ent: missing required field "GenerationModel.is_hidden"`)}
	}
	if _, ok := gmc.mutation.DisplayWeight(); !ok {
		return &ValidationError{Name: "display_weight", err: errors.New(`ent: missing required field "GenerationModel.display_weight"`)}
	}
	if _, ok := gmc.mutation.DefaultWidth(); !ok {
		return &ValidationError{Name: "default_width", err: errors.New(`ent: missing required field "GenerationModel.default_width"`)}
	}
	if _, ok := gmc.mutation.DefaultHeight(); !ok {
		return &ValidationError{Name: "default_height", err: errors.New(`ent: missing required field "GenerationModel.default_height"`)}
	}
	if _, ok := gmc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "GenerationModel.created_at"`)}
	}
	if _, ok := gmc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "GenerationModel.updated_at"`)}
	}
	return nil
}

func (gmc *GenerationModelCreate) sqlSave(ctx context.Context) (*GenerationModel, error) {
	if err := gmc.check(); err != nil {
		return nil, err
	}
	_node, _spec := gmc.createSpec()
	if err := sqlgraph.CreateNode(ctx, gmc.driver, _spec); err != nil {
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
	gmc.mutation.id = &_node.ID
	gmc.mutation.done = true
	return _node, nil
}

func (gmc *GenerationModelCreate) createSpec() (*GenerationModel, *sqlgraph.CreateSpec) {
	var (
		_node = &GenerationModel{config: gmc.config}
		_spec = sqlgraph.NewCreateSpec(generationmodel.Table, sqlgraph.NewFieldSpec(generationmodel.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = gmc.conflict
	if id, ok := gmc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := gmc.mutation.NameInWorker(); ok {
		_spec.SetField(generationmodel.FieldNameInWorker, field.TypeString, value)
		_node.NameInWorker = value
	}
	if value, ok := gmc.mutation.IsActive(); ok {
		_spec.SetField(generationmodel.FieldIsActive, field.TypeBool, value)
		_node.IsActive = value
	}
	if value, ok := gmc.mutation.IsDefault(); ok {
		_spec.SetField(generationmodel.FieldIsDefault, field.TypeBool, value)
		_node.IsDefault = value
	}
	if value, ok := gmc.mutation.IsHidden(); ok {
		_spec.SetField(generationmodel.FieldIsHidden, field.TypeBool, value)
		_node.IsHidden = value
	}
	if value, ok := gmc.mutation.DisplayWeight(); ok {
		_spec.SetField(generationmodel.FieldDisplayWeight, field.TypeInt32, value)
		_node.DisplayWeight = value
	}
	if value, ok := gmc.mutation.DefaultSchedulerID(); ok {
		_spec.SetField(generationmodel.FieldDefaultSchedulerID, field.TypeUUID, value)
		_node.DefaultSchedulerID = &value
	}
	if value, ok := gmc.mutation.DefaultWidth(); ok {
		_spec.SetField(generationmodel.FieldDefaultWidth, field.TypeInt32, value)
		_node.DefaultWidth = value
	}
	if value, ok := gmc.mutation.DefaultHeight(); ok {
		_spec.SetField(generationmodel.FieldDefaultHeight, field.TypeInt32, value)
		_node.DefaultHeight = value
	}
	if value, ok := gmc.mutation.CreatedAt(); ok {
		_spec.SetField(generationmodel.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := gmc.mutation.UpdatedAt(); ok {
		_spec.SetField(generationmodel.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if nodes := gmc.mutation.GenerationsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   generationmodel.GenerationsTable,
			Columns: []string{generationmodel.GenerationsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(generation.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := gmc.mutation.SchedulersIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2M,
			Inverse: false,
			Table:   generationmodel.SchedulersTable,
			Columns: generationmodel.SchedulersPrimaryKey,
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(scheduler.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.GenerationModel.Create().
//		SetNameInWorker(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.GenerationModelUpsert) {
//			SetNameInWorker(v+v).
//		}).
//		Exec(ctx)
func (gmc *GenerationModelCreate) OnConflict(opts ...sql.ConflictOption) *GenerationModelUpsertOne {
	gmc.conflict = opts
	return &GenerationModelUpsertOne{
		create: gmc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.GenerationModel.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (gmc *GenerationModelCreate) OnConflictColumns(columns ...string) *GenerationModelUpsertOne {
	gmc.conflict = append(gmc.conflict, sql.ConflictColumns(columns...))
	return &GenerationModelUpsertOne{
		create: gmc,
	}
}

type (
	// GenerationModelUpsertOne is the builder for "upsert"-ing
	//  one GenerationModel node.
	GenerationModelUpsertOne struct {
		create *GenerationModelCreate
	}

	// GenerationModelUpsert is the "OnConflict" setter.
	GenerationModelUpsert struct {
		*sql.UpdateSet
	}
)

// SetNameInWorker sets the "name_in_worker" field.
func (u *GenerationModelUpsert) SetNameInWorker(v string) *GenerationModelUpsert {
	u.Set(generationmodel.FieldNameInWorker, v)
	return u
}

// UpdateNameInWorker sets the "name_in_worker" field to the value that was provided on create.
func (u *GenerationModelUpsert) UpdateNameInWorker() *GenerationModelUpsert {
	u.SetExcluded(generationmodel.FieldNameInWorker)
	return u
}

// SetIsActive sets the "is_active" field.
func (u *GenerationModelUpsert) SetIsActive(v bool) *GenerationModelUpsert {
	u.Set(generationmodel.FieldIsActive, v)
	return u
}

// UpdateIsActive sets the "is_active" field to the value that was provided on create.
func (u *GenerationModelUpsert) UpdateIsActive() *GenerationModelUpsert {
	u.SetExcluded(generationmodel.FieldIsActive)
	return u
}

// SetIsDefault sets the "is_default" field.
func (u *GenerationModelUpsert) SetIsDefault(v bool) *GenerationModelUpsert {
	u.Set(generationmodel.FieldIsDefault, v)
	return u
}

// UpdateIsDefault sets the "is_default" field to the value that was provided on create.
func (u *GenerationModelUpsert) UpdateIsDefault() *GenerationModelUpsert {
	u.SetExcluded(generationmodel.FieldIsDefault)
	return u
}

// SetIsHidden sets the "is_hidden" field.
func (u *GenerationModelUpsert) SetIsHidden(v bool) *GenerationModelUpsert {
	u.Set(generationmodel.FieldIsHidden, v)
	return u
}

// UpdateIsHidden sets the "is_hidden" field to the value that was provided on create.
func (u *GenerationModelUpsert) UpdateIsHidden() *GenerationModelUpsert {
	u.SetExcluded(generationmodel.FieldIsHidden)
	return u
}

// SetDisplayWeight sets the "display_weight" field.
func (u *GenerationModelUpsert) SetDisplayWeight(v int32) *GenerationModelUpsert {
	u.Set(generationmodel.FieldDisplayWeight, v)
	return u
}

// UpdateDisplayWeight sets the "display_weight" field to the value that was provided on create.
func (u *GenerationModelUpsert) UpdateDisplayWeight() *GenerationModelUpsert {
	u.SetExcluded(generationmodel.FieldDisplayWeight)
	return u
}

// AddDisplayWeight adds v to the "display_weight" field.
func (u *GenerationModelUpsert) AddDisplayWeight(v int32) *GenerationModelUpsert {
	u.Add(generationmodel.FieldDisplayWeight, v)
	return u
}

// SetDefaultSchedulerID sets the "default_scheduler_id" field.
func (u *GenerationModelUpsert) SetDefaultSchedulerID(v uuid.UUID) *GenerationModelUpsert {
	u.Set(generationmodel.FieldDefaultSchedulerID, v)
	return u
}

// UpdateDefaultSchedulerID sets the "default_scheduler_id" field to the value that was provided on create.
func (u *GenerationModelUpsert) UpdateDefaultSchedulerID() *GenerationModelUpsert {
	u.SetExcluded(generationmodel.FieldDefaultSchedulerID)
	return u
}

// ClearDefaultSchedulerID clears the value of the "default_scheduler_id" field.
func (u *GenerationModelUpsert) ClearDefaultSchedulerID() *GenerationModelUpsert {
	u.SetNull(generationmodel.FieldDefaultSchedulerID)
	return u
}

// SetDefaultWidth sets the "default_width" field.
func (u *GenerationModelUpsert) SetDefaultWidth(v int32) *GenerationModelUpsert {
	u.Set(generationmodel.FieldDefaultWidth, v)
	return u
}

// UpdateDefaultWidth sets the "default_width" field to the value that was provided on create.
func (u *GenerationModelUpsert) UpdateDefaultWidth() *GenerationModelUpsert {
	u.SetExcluded(generationmodel.FieldDefaultWidth)
	return u
}

// AddDefaultWidth adds v to the "default_width" field.
func (u *GenerationModelUpsert) AddDefaultWidth(v int32) *GenerationModelUpsert {
	u.Add(generationmodel.FieldDefaultWidth, v)
	return u
}

// SetDefaultHeight sets the "default_height" field.
func (u *GenerationModelUpsert) SetDefaultHeight(v int32) *GenerationModelUpsert {
	u.Set(generationmodel.FieldDefaultHeight, v)
	return u
}

// UpdateDefaultHeight sets the "default_height" field to the value that was provided on create.
func (u *GenerationModelUpsert) UpdateDefaultHeight() *GenerationModelUpsert {
	u.SetExcluded(generationmodel.FieldDefaultHeight)
	return u
}

// AddDefaultHeight adds v to the "default_height" field.
func (u *GenerationModelUpsert) AddDefaultHeight(v int32) *GenerationModelUpsert {
	u.Add(generationmodel.FieldDefaultHeight, v)
	return u
}

// SetUpdatedAt sets the "updated_at" field.
func (u *GenerationModelUpsert) SetUpdatedAt(v time.Time) *GenerationModelUpsert {
	u.Set(generationmodel.FieldUpdatedAt, v)
	return u
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *GenerationModelUpsert) UpdateUpdatedAt() *GenerationModelUpsert {
	u.SetExcluded(generationmodel.FieldUpdatedAt)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.GenerationModel.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(generationmodel.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *GenerationModelUpsertOne) UpdateNewValues() *GenerationModelUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(generationmodel.FieldID)
		}
		if _, exists := u.create.mutation.CreatedAt(); exists {
			s.SetIgnore(generationmodel.FieldCreatedAt)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.GenerationModel.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *GenerationModelUpsertOne) Ignore() *GenerationModelUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *GenerationModelUpsertOne) DoNothing() *GenerationModelUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the GenerationModelCreate.OnConflict
// documentation for more info.
func (u *GenerationModelUpsertOne) Update(set func(*GenerationModelUpsert)) *GenerationModelUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&GenerationModelUpsert{UpdateSet: update})
	}))
	return u
}

// SetNameInWorker sets the "name_in_worker" field.
func (u *GenerationModelUpsertOne) SetNameInWorker(v string) *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetNameInWorker(v)
	})
}

// UpdateNameInWorker sets the "name_in_worker" field to the value that was provided on create.
func (u *GenerationModelUpsertOne) UpdateNameInWorker() *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateNameInWorker()
	})
}

// SetIsActive sets the "is_active" field.
func (u *GenerationModelUpsertOne) SetIsActive(v bool) *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetIsActive(v)
	})
}

// UpdateIsActive sets the "is_active" field to the value that was provided on create.
func (u *GenerationModelUpsertOne) UpdateIsActive() *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateIsActive()
	})
}

// SetIsDefault sets the "is_default" field.
func (u *GenerationModelUpsertOne) SetIsDefault(v bool) *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetIsDefault(v)
	})
}

// UpdateIsDefault sets the "is_default" field to the value that was provided on create.
func (u *GenerationModelUpsertOne) UpdateIsDefault() *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateIsDefault()
	})
}

// SetIsHidden sets the "is_hidden" field.
func (u *GenerationModelUpsertOne) SetIsHidden(v bool) *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetIsHidden(v)
	})
}

// UpdateIsHidden sets the "is_hidden" field to the value that was provided on create.
func (u *GenerationModelUpsertOne) UpdateIsHidden() *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateIsHidden()
	})
}

// SetDisplayWeight sets the "display_weight" field.
func (u *GenerationModelUpsertOne) SetDisplayWeight(v int32) *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetDisplayWeight(v)
	})
}

// AddDisplayWeight adds v to the "display_weight" field.
func (u *GenerationModelUpsertOne) AddDisplayWeight(v int32) *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.AddDisplayWeight(v)
	})
}

// UpdateDisplayWeight sets the "display_weight" field to the value that was provided on create.
func (u *GenerationModelUpsertOne) UpdateDisplayWeight() *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateDisplayWeight()
	})
}

// SetDefaultSchedulerID sets the "default_scheduler_id" field.
func (u *GenerationModelUpsertOne) SetDefaultSchedulerID(v uuid.UUID) *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetDefaultSchedulerID(v)
	})
}

// UpdateDefaultSchedulerID sets the "default_scheduler_id" field to the value that was provided on create.
func (u *GenerationModelUpsertOne) UpdateDefaultSchedulerID() *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateDefaultSchedulerID()
	})
}

// ClearDefaultSchedulerID clears the value of the "default_scheduler_id" field.
func (u *GenerationModelUpsertOne) ClearDefaultSchedulerID() *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.ClearDefaultSchedulerID()
	})
}

// SetDefaultWidth sets the "default_width" field.
func (u *GenerationModelUpsertOne) SetDefaultWidth(v int32) *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetDefaultWidth(v)
	})
}

// AddDefaultWidth adds v to the "default_width" field.
func (u *GenerationModelUpsertOne) AddDefaultWidth(v int32) *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.AddDefaultWidth(v)
	})
}

// UpdateDefaultWidth sets the "default_width" field to the value that was provided on create.
func (u *GenerationModelUpsertOne) UpdateDefaultWidth() *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateDefaultWidth()
	})
}

// SetDefaultHeight sets the "default_height" field.
func (u *GenerationModelUpsertOne) SetDefaultHeight(v int32) *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetDefaultHeight(v)
	})
}

// AddDefaultHeight adds v to the "default_height" field.
func (u *GenerationModelUpsertOne) AddDefaultHeight(v int32) *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.AddDefaultHeight(v)
	})
}

// UpdateDefaultHeight sets the "default_height" field to the value that was provided on create.
func (u *GenerationModelUpsertOne) UpdateDefaultHeight() *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateDefaultHeight()
	})
}

// SetUpdatedAt sets the "updated_at" field.
func (u *GenerationModelUpsertOne) SetUpdatedAt(v time.Time) *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *GenerationModelUpsertOne) UpdateUpdatedAt() *GenerationModelUpsertOne {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateUpdatedAt()
	})
}

// Exec executes the query.
func (u *GenerationModelUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for GenerationModelCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *GenerationModelUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *GenerationModelUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: GenerationModelUpsertOne.ID is not supported by MySQL driver. Use GenerationModelUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *GenerationModelUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// GenerationModelCreateBulk is the builder for creating many GenerationModel entities in bulk.
type GenerationModelCreateBulk struct {
	config
	err      error
	builders []*GenerationModelCreate
	conflict []sql.ConflictOption
}

// Save creates the GenerationModel entities in the database.
func (gmcb *GenerationModelCreateBulk) Save(ctx context.Context) ([]*GenerationModel, error) {
	if gmcb.err != nil {
		return nil, gmcb.err
	}
	specs := make([]*sqlgraph.CreateSpec, len(gmcb.builders))
	nodes := make([]*GenerationModel, len(gmcb.builders))
	mutators := make([]Mutator, len(gmcb.builders))
	for i := range gmcb.builders {
		func(i int, root context.Context) {
			builder := gmcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*GenerationModelMutation)
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
					_, err = mutators[i+1].Mutate(root, gmcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = gmcb.conflict
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, gmcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, gmcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (gmcb *GenerationModelCreateBulk) SaveX(ctx context.Context) []*GenerationModel {
	v, err := gmcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (gmcb *GenerationModelCreateBulk) Exec(ctx context.Context) error {
	_, err := gmcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (gmcb *GenerationModelCreateBulk) ExecX(ctx context.Context) {
	if err := gmcb.Exec(ctx); err != nil {
		panic(err)
	}
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.GenerationModel.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.GenerationModelUpsert) {
//			SetNameInWorker(v+v).
//		}).
//		Exec(ctx)
func (gmcb *GenerationModelCreateBulk) OnConflict(opts ...sql.ConflictOption) *GenerationModelUpsertBulk {
	gmcb.conflict = opts
	return &GenerationModelUpsertBulk{
		create: gmcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.GenerationModel.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (gmcb *GenerationModelCreateBulk) OnConflictColumns(columns ...string) *GenerationModelUpsertBulk {
	gmcb.conflict = append(gmcb.conflict, sql.ConflictColumns(columns...))
	return &GenerationModelUpsertBulk{
		create: gmcb,
	}
}

// GenerationModelUpsertBulk is the builder for "upsert"-ing
// a bulk of GenerationModel nodes.
type GenerationModelUpsertBulk struct {
	create *GenerationModelCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.GenerationModel.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(generationmodel.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *GenerationModelUpsertBulk) UpdateNewValues() *GenerationModelUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(generationmodel.FieldID)
			}
			if _, exists := b.mutation.CreatedAt(); exists {
				s.SetIgnore(generationmodel.FieldCreatedAt)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.GenerationModel.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *GenerationModelUpsertBulk) Ignore() *GenerationModelUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *GenerationModelUpsertBulk) DoNothing() *GenerationModelUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the GenerationModelCreateBulk.OnConflict
// documentation for more info.
func (u *GenerationModelUpsertBulk) Update(set func(*GenerationModelUpsert)) *GenerationModelUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&GenerationModelUpsert{UpdateSet: update})
	}))
	return u
}

// SetNameInWorker sets the "name_in_worker" field.
func (u *GenerationModelUpsertBulk) SetNameInWorker(v string) *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetNameInWorker(v)
	})
}

// UpdateNameInWorker sets the "name_in_worker" field to the value that was provided on create.
func (u *GenerationModelUpsertBulk) UpdateNameInWorker() *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateNameInWorker()
	})
}

// SetIsActive sets the "is_active" field.
func (u *GenerationModelUpsertBulk) SetIsActive(v bool) *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetIsActive(v)
	})
}

// UpdateIsActive sets the "is_active" field to the value that was provided on create.
func (u *GenerationModelUpsertBulk) UpdateIsActive() *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateIsActive()
	})
}

// SetIsDefault sets the "is_default" field.
func (u *GenerationModelUpsertBulk) SetIsDefault(v bool) *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetIsDefault(v)
	})
}

// UpdateIsDefault sets the "is_default" field to the value that was provided on create.
func (u *GenerationModelUpsertBulk) UpdateIsDefault() *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateIsDefault()
	})
}

// SetIsHidden sets the "is_hidden" field.
func (u *GenerationModelUpsertBulk) SetIsHidden(v bool) *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetIsHidden(v)
	})
}

// UpdateIsHidden sets the "is_hidden" field to the value that was provided on create.
func (u *GenerationModelUpsertBulk) UpdateIsHidden() *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateIsHidden()
	})
}

// SetDisplayWeight sets the "display_weight" field.
func (u *GenerationModelUpsertBulk) SetDisplayWeight(v int32) *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetDisplayWeight(v)
	})
}

// AddDisplayWeight adds v to the "display_weight" field.
func (u *GenerationModelUpsertBulk) AddDisplayWeight(v int32) *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.AddDisplayWeight(v)
	})
}

// UpdateDisplayWeight sets the "display_weight" field to the value that was provided on create.
func (u *GenerationModelUpsertBulk) UpdateDisplayWeight() *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateDisplayWeight()
	})
}

// SetDefaultSchedulerID sets the "default_scheduler_id" field.
func (u *GenerationModelUpsertBulk) SetDefaultSchedulerID(v uuid.UUID) *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetDefaultSchedulerID(v)
	})
}

// UpdateDefaultSchedulerID sets the "default_scheduler_id" field to the value that was provided on create.
func (u *GenerationModelUpsertBulk) UpdateDefaultSchedulerID() *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateDefaultSchedulerID()
	})
}

// ClearDefaultSchedulerID clears the value of the "default_scheduler_id" field.
func (u *GenerationModelUpsertBulk) ClearDefaultSchedulerID() *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.ClearDefaultSchedulerID()
	})
}

// SetDefaultWidth sets the "default_width" field.
func (u *GenerationModelUpsertBulk) SetDefaultWidth(v int32) *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetDefaultWidth(v)
	})
}

// AddDefaultWidth adds v to the "default_width" field.
func (u *GenerationModelUpsertBulk) AddDefaultWidth(v int32) *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.AddDefaultWidth(v)
	})
}

// UpdateDefaultWidth sets the "default_width" field to the value that was provided on create.
func (u *GenerationModelUpsertBulk) UpdateDefaultWidth() *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateDefaultWidth()
	})
}

// SetDefaultHeight sets the "default_height" field.
func (u *GenerationModelUpsertBulk) SetDefaultHeight(v int32) *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetDefaultHeight(v)
	})
}

// AddDefaultHeight adds v to the "default_height" field.
func (u *GenerationModelUpsertBulk) AddDefaultHeight(v int32) *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.AddDefaultHeight(v)
	})
}

// UpdateDefaultHeight sets the "default_height" field to the value that was provided on create.
func (u *GenerationModelUpsertBulk) UpdateDefaultHeight() *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateDefaultHeight()
	})
}

// SetUpdatedAt sets the "updated_at" field.
func (u *GenerationModelUpsertBulk) SetUpdatedAt(v time.Time) *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *GenerationModelUpsertBulk) UpdateUpdatedAt() *GenerationModelUpsertBulk {
	return u.Update(func(s *GenerationModelUpsert) {
		s.UpdateUpdatedAt()
	})
}

// Exec executes the query.
func (u *GenerationModelUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the GenerationModelCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for GenerationModelCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *GenerationModelUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
