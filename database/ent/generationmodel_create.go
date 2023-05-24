// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	return withHooks[*GenerationModel, GenerationModelMutation](ctx, gmc.sqlSave, gmc.mutation, gmc.hooks)
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
		_spec = &sqlgraph.CreateSpec{
			Table: generationmodel.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: generationmodel.FieldID,
			},
		}
	)
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
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: generation.FieldID,
				},
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
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: scheduler.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// GenerationModelCreateBulk is the builder for creating many GenerationModel entities in bulk.
type GenerationModelCreateBulk struct {
	config
	builders []*GenerationModelCreate
}

// Save creates the GenerationModel entities in the database.
func (gmcb *GenerationModelCreateBulk) Save(ctx context.Context) ([]*GenerationModel, error) {
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
				nodes[i], specs[i] = builder.createSpec()
				var err error
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, gmcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
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
