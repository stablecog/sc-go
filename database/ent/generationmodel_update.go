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
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/generationmodel"
	"github.com/stablecog/sc-go/database/ent/predicate"
	"github.com/stablecog/sc-go/database/ent/scheduler"
)

// GenerationModelUpdate is the builder for updating GenerationModel entities.
type GenerationModelUpdate struct {
	config
	hooks     []Hook
	mutation  *GenerationModelMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the GenerationModelUpdate builder.
func (gmu *GenerationModelUpdate) Where(ps ...predicate.GenerationModel) *GenerationModelUpdate {
	gmu.mutation.Where(ps...)
	return gmu
}

// SetNameInWorker sets the "name_in_worker" field.
func (gmu *GenerationModelUpdate) SetNameInWorker(s string) *GenerationModelUpdate {
	gmu.mutation.SetNameInWorker(s)
	return gmu
}

// SetNillableNameInWorker sets the "name_in_worker" field if the given value is not nil.
func (gmu *GenerationModelUpdate) SetNillableNameInWorker(s *string) *GenerationModelUpdate {
	if s != nil {
		gmu.SetNameInWorker(*s)
	}
	return gmu
}

// SetIsActive sets the "is_active" field.
func (gmu *GenerationModelUpdate) SetIsActive(b bool) *GenerationModelUpdate {
	gmu.mutation.SetIsActive(b)
	return gmu
}

// SetNillableIsActive sets the "is_active" field if the given value is not nil.
func (gmu *GenerationModelUpdate) SetNillableIsActive(b *bool) *GenerationModelUpdate {
	if b != nil {
		gmu.SetIsActive(*b)
	}
	return gmu
}

// SetIsDefault sets the "is_default" field.
func (gmu *GenerationModelUpdate) SetIsDefault(b bool) *GenerationModelUpdate {
	gmu.mutation.SetIsDefault(b)
	return gmu
}

// SetNillableIsDefault sets the "is_default" field if the given value is not nil.
func (gmu *GenerationModelUpdate) SetNillableIsDefault(b *bool) *GenerationModelUpdate {
	if b != nil {
		gmu.SetIsDefault(*b)
	}
	return gmu
}

// SetIsHidden sets the "is_hidden" field.
func (gmu *GenerationModelUpdate) SetIsHidden(b bool) *GenerationModelUpdate {
	gmu.mutation.SetIsHidden(b)
	return gmu
}

// SetNillableIsHidden sets the "is_hidden" field if the given value is not nil.
func (gmu *GenerationModelUpdate) SetNillableIsHidden(b *bool) *GenerationModelUpdate {
	if b != nil {
		gmu.SetIsHidden(*b)
	}
	return gmu
}

// SetDisplayWeight sets the "display_weight" field.
func (gmu *GenerationModelUpdate) SetDisplayWeight(i int32) *GenerationModelUpdate {
	gmu.mutation.ResetDisplayWeight()
	gmu.mutation.SetDisplayWeight(i)
	return gmu
}

// SetNillableDisplayWeight sets the "display_weight" field if the given value is not nil.
func (gmu *GenerationModelUpdate) SetNillableDisplayWeight(i *int32) *GenerationModelUpdate {
	if i != nil {
		gmu.SetDisplayWeight(*i)
	}
	return gmu
}

// AddDisplayWeight adds i to the "display_weight" field.
func (gmu *GenerationModelUpdate) AddDisplayWeight(i int32) *GenerationModelUpdate {
	gmu.mutation.AddDisplayWeight(i)
	return gmu
}

// SetDefaultSchedulerID sets the "default_scheduler_id" field.
func (gmu *GenerationModelUpdate) SetDefaultSchedulerID(u uuid.UUID) *GenerationModelUpdate {
	gmu.mutation.SetDefaultSchedulerID(u)
	return gmu
}

// SetNillableDefaultSchedulerID sets the "default_scheduler_id" field if the given value is not nil.
func (gmu *GenerationModelUpdate) SetNillableDefaultSchedulerID(u *uuid.UUID) *GenerationModelUpdate {
	if u != nil {
		gmu.SetDefaultSchedulerID(*u)
	}
	return gmu
}

// ClearDefaultSchedulerID clears the value of the "default_scheduler_id" field.
func (gmu *GenerationModelUpdate) ClearDefaultSchedulerID() *GenerationModelUpdate {
	gmu.mutation.ClearDefaultSchedulerID()
	return gmu
}

// SetDefaultWidth sets the "default_width" field.
func (gmu *GenerationModelUpdate) SetDefaultWidth(i int32) *GenerationModelUpdate {
	gmu.mutation.ResetDefaultWidth()
	gmu.mutation.SetDefaultWidth(i)
	return gmu
}

// SetNillableDefaultWidth sets the "default_width" field if the given value is not nil.
func (gmu *GenerationModelUpdate) SetNillableDefaultWidth(i *int32) *GenerationModelUpdate {
	if i != nil {
		gmu.SetDefaultWidth(*i)
	}
	return gmu
}

// AddDefaultWidth adds i to the "default_width" field.
func (gmu *GenerationModelUpdate) AddDefaultWidth(i int32) *GenerationModelUpdate {
	gmu.mutation.AddDefaultWidth(i)
	return gmu
}

// SetDefaultHeight sets the "default_height" field.
func (gmu *GenerationModelUpdate) SetDefaultHeight(i int32) *GenerationModelUpdate {
	gmu.mutation.ResetDefaultHeight()
	gmu.mutation.SetDefaultHeight(i)
	return gmu
}

// SetNillableDefaultHeight sets the "default_height" field if the given value is not nil.
func (gmu *GenerationModelUpdate) SetNillableDefaultHeight(i *int32) *GenerationModelUpdate {
	if i != nil {
		gmu.SetDefaultHeight(*i)
	}
	return gmu
}

// AddDefaultHeight adds i to the "default_height" field.
func (gmu *GenerationModelUpdate) AddDefaultHeight(i int32) *GenerationModelUpdate {
	gmu.mutation.AddDefaultHeight(i)
	return gmu
}

// SetUpdatedAt sets the "updated_at" field.
func (gmu *GenerationModelUpdate) SetUpdatedAt(t time.Time) *GenerationModelUpdate {
	gmu.mutation.SetUpdatedAt(t)
	return gmu
}

// AddGenerationIDs adds the "generations" edge to the Generation entity by IDs.
func (gmu *GenerationModelUpdate) AddGenerationIDs(ids ...uuid.UUID) *GenerationModelUpdate {
	gmu.mutation.AddGenerationIDs(ids...)
	return gmu
}

// AddGenerations adds the "generations" edges to the Generation entity.
func (gmu *GenerationModelUpdate) AddGenerations(g ...*Generation) *GenerationModelUpdate {
	ids := make([]uuid.UUID, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return gmu.AddGenerationIDs(ids...)
}

// AddSchedulerIDs adds the "schedulers" edge to the Scheduler entity by IDs.
func (gmu *GenerationModelUpdate) AddSchedulerIDs(ids ...uuid.UUID) *GenerationModelUpdate {
	gmu.mutation.AddSchedulerIDs(ids...)
	return gmu
}

// AddSchedulers adds the "schedulers" edges to the Scheduler entity.
func (gmu *GenerationModelUpdate) AddSchedulers(s ...*Scheduler) *GenerationModelUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return gmu.AddSchedulerIDs(ids...)
}

// Mutation returns the GenerationModelMutation object of the builder.
func (gmu *GenerationModelUpdate) Mutation() *GenerationModelMutation {
	return gmu.mutation
}

// ClearGenerations clears all "generations" edges to the Generation entity.
func (gmu *GenerationModelUpdate) ClearGenerations() *GenerationModelUpdate {
	gmu.mutation.ClearGenerations()
	return gmu
}

// RemoveGenerationIDs removes the "generations" edge to Generation entities by IDs.
func (gmu *GenerationModelUpdate) RemoveGenerationIDs(ids ...uuid.UUID) *GenerationModelUpdate {
	gmu.mutation.RemoveGenerationIDs(ids...)
	return gmu
}

// RemoveGenerations removes "generations" edges to Generation entities.
func (gmu *GenerationModelUpdate) RemoveGenerations(g ...*Generation) *GenerationModelUpdate {
	ids := make([]uuid.UUID, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return gmu.RemoveGenerationIDs(ids...)
}

// ClearSchedulers clears all "schedulers" edges to the Scheduler entity.
func (gmu *GenerationModelUpdate) ClearSchedulers() *GenerationModelUpdate {
	gmu.mutation.ClearSchedulers()
	return gmu
}

// RemoveSchedulerIDs removes the "schedulers" edge to Scheduler entities by IDs.
func (gmu *GenerationModelUpdate) RemoveSchedulerIDs(ids ...uuid.UUID) *GenerationModelUpdate {
	gmu.mutation.RemoveSchedulerIDs(ids...)
	return gmu
}

// RemoveSchedulers removes "schedulers" edges to Scheduler entities.
func (gmu *GenerationModelUpdate) RemoveSchedulers(s ...*Scheduler) *GenerationModelUpdate {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return gmu.RemoveSchedulerIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (gmu *GenerationModelUpdate) Save(ctx context.Context) (int, error) {
	gmu.defaults()
	return withHooks(ctx, gmu.sqlSave, gmu.mutation, gmu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (gmu *GenerationModelUpdate) SaveX(ctx context.Context) int {
	affected, err := gmu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (gmu *GenerationModelUpdate) Exec(ctx context.Context) error {
	_, err := gmu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (gmu *GenerationModelUpdate) ExecX(ctx context.Context) {
	if err := gmu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (gmu *GenerationModelUpdate) defaults() {
	if _, ok := gmu.mutation.UpdatedAt(); !ok {
		v := generationmodel.UpdateDefaultUpdatedAt()
		gmu.mutation.SetUpdatedAt(v)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (gmu *GenerationModelUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *GenerationModelUpdate {
	gmu.modifiers = append(gmu.modifiers, modifiers...)
	return gmu
}

func (gmu *GenerationModelUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(generationmodel.Table, generationmodel.Columns, sqlgraph.NewFieldSpec(generationmodel.FieldID, field.TypeUUID))
	if ps := gmu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := gmu.mutation.NameInWorker(); ok {
		_spec.SetField(generationmodel.FieldNameInWorker, field.TypeString, value)
	}
	if value, ok := gmu.mutation.IsActive(); ok {
		_spec.SetField(generationmodel.FieldIsActive, field.TypeBool, value)
	}
	if value, ok := gmu.mutation.IsDefault(); ok {
		_spec.SetField(generationmodel.FieldIsDefault, field.TypeBool, value)
	}
	if value, ok := gmu.mutation.IsHidden(); ok {
		_spec.SetField(generationmodel.FieldIsHidden, field.TypeBool, value)
	}
	if value, ok := gmu.mutation.DisplayWeight(); ok {
		_spec.SetField(generationmodel.FieldDisplayWeight, field.TypeInt32, value)
	}
	if value, ok := gmu.mutation.AddedDisplayWeight(); ok {
		_spec.AddField(generationmodel.FieldDisplayWeight, field.TypeInt32, value)
	}
	if value, ok := gmu.mutation.DefaultSchedulerID(); ok {
		_spec.SetField(generationmodel.FieldDefaultSchedulerID, field.TypeUUID, value)
	}
	if gmu.mutation.DefaultSchedulerIDCleared() {
		_spec.ClearField(generationmodel.FieldDefaultSchedulerID, field.TypeUUID)
	}
	if value, ok := gmu.mutation.DefaultWidth(); ok {
		_spec.SetField(generationmodel.FieldDefaultWidth, field.TypeInt32, value)
	}
	if value, ok := gmu.mutation.AddedDefaultWidth(); ok {
		_spec.AddField(generationmodel.FieldDefaultWidth, field.TypeInt32, value)
	}
	if value, ok := gmu.mutation.DefaultHeight(); ok {
		_spec.SetField(generationmodel.FieldDefaultHeight, field.TypeInt32, value)
	}
	if value, ok := gmu.mutation.AddedDefaultHeight(); ok {
		_spec.AddField(generationmodel.FieldDefaultHeight, field.TypeInt32, value)
	}
	if value, ok := gmu.mutation.UpdatedAt(); ok {
		_spec.SetField(generationmodel.FieldUpdatedAt, field.TypeTime, value)
	}
	if gmu.mutation.GenerationsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := gmu.mutation.RemovedGenerationsIDs(); len(nodes) > 0 && !gmu.mutation.GenerationsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := gmu.mutation.GenerationsIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if gmu.mutation.SchedulersCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := gmu.mutation.RemovedSchedulersIDs(); len(nodes) > 0 && !gmu.mutation.SchedulersCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := gmu.mutation.SchedulersIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(gmu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, gmu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{generationmodel.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	gmu.mutation.done = true
	return n, nil
}

// GenerationModelUpdateOne is the builder for updating a single GenerationModel entity.
type GenerationModelUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *GenerationModelMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetNameInWorker sets the "name_in_worker" field.
func (gmuo *GenerationModelUpdateOne) SetNameInWorker(s string) *GenerationModelUpdateOne {
	gmuo.mutation.SetNameInWorker(s)
	return gmuo
}

// SetNillableNameInWorker sets the "name_in_worker" field if the given value is not nil.
func (gmuo *GenerationModelUpdateOne) SetNillableNameInWorker(s *string) *GenerationModelUpdateOne {
	if s != nil {
		gmuo.SetNameInWorker(*s)
	}
	return gmuo
}

// SetIsActive sets the "is_active" field.
func (gmuo *GenerationModelUpdateOne) SetIsActive(b bool) *GenerationModelUpdateOne {
	gmuo.mutation.SetIsActive(b)
	return gmuo
}

// SetNillableIsActive sets the "is_active" field if the given value is not nil.
func (gmuo *GenerationModelUpdateOne) SetNillableIsActive(b *bool) *GenerationModelUpdateOne {
	if b != nil {
		gmuo.SetIsActive(*b)
	}
	return gmuo
}

// SetIsDefault sets the "is_default" field.
func (gmuo *GenerationModelUpdateOne) SetIsDefault(b bool) *GenerationModelUpdateOne {
	gmuo.mutation.SetIsDefault(b)
	return gmuo
}

// SetNillableIsDefault sets the "is_default" field if the given value is not nil.
func (gmuo *GenerationModelUpdateOne) SetNillableIsDefault(b *bool) *GenerationModelUpdateOne {
	if b != nil {
		gmuo.SetIsDefault(*b)
	}
	return gmuo
}

// SetIsHidden sets the "is_hidden" field.
func (gmuo *GenerationModelUpdateOne) SetIsHidden(b bool) *GenerationModelUpdateOne {
	gmuo.mutation.SetIsHidden(b)
	return gmuo
}

// SetNillableIsHidden sets the "is_hidden" field if the given value is not nil.
func (gmuo *GenerationModelUpdateOne) SetNillableIsHidden(b *bool) *GenerationModelUpdateOne {
	if b != nil {
		gmuo.SetIsHidden(*b)
	}
	return gmuo
}

// SetDisplayWeight sets the "display_weight" field.
func (gmuo *GenerationModelUpdateOne) SetDisplayWeight(i int32) *GenerationModelUpdateOne {
	gmuo.mutation.ResetDisplayWeight()
	gmuo.mutation.SetDisplayWeight(i)
	return gmuo
}

// SetNillableDisplayWeight sets the "display_weight" field if the given value is not nil.
func (gmuo *GenerationModelUpdateOne) SetNillableDisplayWeight(i *int32) *GenerationModelUpdateOne {
	if i != nil {
		gmuo.SetDisplayWeight(*i)
	}
	return gmuo
}

// AddDisplayWeight adds i to the "display_weight" field.
func (gmuo *GenerationModelUpdateOne) AddDisplayWeight(i int32) *GenerationModelUpdateOne {
	gmuo.mutation.AddDisplayWeight(i)
	return gmuo
}

// SetDefaultSchedulerID sets the "default_scheduler_id" field.
func (gmuo *GenerationModelUpdateOne) SetDefaultSchedulerID(u uuid.UUID) *GenerationModelUpdateOne {
	gmuo.mutation.SetDefaultSchedulerID(u)
	return gmuo
}

// SetNillableDefaultSchedulerID sets the "default_scheduler_id" field if the given value is not nil.
func (gmuo *GenerationModelUpdateOne) SetNillableDefaultSchedulerID(u *uuid.UUID) *GenerationModelUpdateOne {
	if u != nil {
		gmuo.SetDefaultSchedulerID(*u)
	}
	return gmuo
}

// ClearDefaultSchedulerID clears the value of the "default_scheduler_id" field.
func (gmuo *GenerationModelUpdateOne) ClearDefaultSchedulerID() *GenerationModelUpdateOne {
	gmuo.mutation.ClearDefaultSchedulerID()
	return gmuo
}

// SetDefaultWidth sets the "default_width" field.
func (gmuo *GenerationModelUpdateOne) SetDefaultWidth(i int32) *GenerationModelUpdateOne {
	gmuo.mutation.ResetDefaultWidth()
	gmuo.mutation.SetDefaultWidth(i)
	return gmuo
}

// SetNillableDefaultWidth sets the "default_width" field if the given value is not nil.
func (gmuo *GenerationModelUpdateOne) SetNillableDefaultWidth(i *int32) *GenerationModelUpdateOne {
	if i != nil {
		gmuo.SetDefaultWidth(*i)
	}
	return gmuo
}

// AddDefaultWidth adds i to the "default_width" field.
func (gmuo *GenerationModelUpdateOne) AddDefaultWidth(i int32) *GenerationModelUpdateOne {
	gmuo.mutation.AddDefaultWidth(i)
	return gmuo
}

// SetDefaultHeight sets the "default_height" field.
func (gmuo *GenerationModelUpdateOne) SetDefaultHeight(i int32) *GenerationModelUpdateOne {
	gmuo.mutation.ResetDefaultHeight()
	gmuo.mutation.SetDefaultHeight(i)
	return gmuo
}

// SetNillableDefaultHeight sets the "default_height" field if the given value is not nil.
func (gmuo *GenerationModelUpdateOne) SetNillableDefaultHeight(i *int32) *GenerationModelUpdateOne {
	if i != nil {
		gmuo.SetDefaultHeight(*i)
	}
	return gmuo
}

// AddDefaultHeight adds i to the "default_height" field.
func (gmuo *GenerationModelUpdateOne) AddDefaultHeight(i int32) *GenerationModelUpdateOne {
	gmuo.mutation.AddDefaultHeight(i)
	return gmuo
}

// SetUpdatedAt sets the "updated_at" field.
func (gmuo *GenerationModelUpdateOne) SetUpdatedAt(t time.Time) *GenerationModelUpdateOne {
	gmuo.mutation.SetUpdatedAt(t)
	return gmuo
}

// AddGenerationIDs adds the "generations" edge to the Generation entity by IDs.
func (gmuo *GenerationModelUpdateOne) AddGenerationIDs(ids ...uuid.UUID) *GenerationModelUpdateOne {
	gmuo.mutation.AddGenerationIDs(ids...)
	return gmuo
}

// AddGenerations adds the "generations" edges to the Generation entity.
func (gmuo *GenerationModelUpdateOne) AddGenerations(g ...*Generation) *GenerationModelUpdateOne {
	ids := make([]uuid.UUID, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return gmuo.AddGenerationIDs(ids...)
}

// AddSchedulerIDs adds the "schedulers" edge to the Scheduler entity by IDs.
func (gmuo *GenerationModelUpdateOne) AddSchedulerIDs(ids ...uuid.UUID) *GenerationModelUpdateOne {
	gmuo.mutation.AddSchedulerIDs(ids...)
	return gmuo
}

// AddSchedulers adds the "schedulers" edges to the Scheduler entity.
func (gmuo *GenerationModelUpdateOne) AddSchedulers(s ...*Scheduler) *GenerationModelUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return gmuo.AddSchedulerIDs(ids...)
}

// Mutation returns the GenerationModelMutation object of the builder.
func (gmuo *GenerationModelUpdateOne) Mutation() *GenerationModelMutation {
	return gmuo.mutation
}

// ClearGenerations clears all "generations" edges to the Generation entity.
func (gmuo *GenerationModelUpdateOne) ClearGenerations() *GenerationModelUpdateOne {
	gmuo.mutation.ClearGenerations()
	return gmuo
}

// RemoveGenerationIDs removes the "generations" edge to Generation entities by IDs.
func (gmuo *GenerationModelUpdateOne) RemoveGenerationIDs(ids ...uuid.UUID) *GenerationModelUpdateOne {
	gmuo.mutation.RemoveGenerationIDs(ids...)
	return gmuo
}

// RemoveGenerations removes "generations" edges to Generation entities.
func (gmuo *GenerationModelUpdateOne) RemoveGenerations(g ...*Generation) *GenerationModelUpdateOne {
	ids := make([]uuid.UUID, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return gmuo.RemoveGenerationIDs(ids...)
}

// ClearSchedulers clears all "schedulers" edges to the Scheduler entity.
func (gmuo *GenerationModelUpdateOne) ClearSchedulers() *GenerationModelUpdateOne {
	gmuo.mutation.ClearSchedulers()
	return gmuo
}

// RemoveSchedulerIDs removes the "schedulers" edge to Scheduler entities by IDs.
func (gmuo *GenerationModelUpdateOne) RemoveSchedulerIDs(ids ...uuid.UUID) *GenerationModelUpdateOne {
	gmuo.mutation.RemoveSchedulerIDs(ids...)
	return gmuo
}

// RemoveSchedulers removes "schedulers" edges to Scheduler entities.
func (gmuo *GenerationModelUpdateOne) RemoveSchedulers(s ...*Scheduler) *GenerationModelUpdateOne {
	ids := make([]uuid.UUID, len(s))
	for i := range s {
		ids[i] = s[i].ID
	}
	return gmuo.RemoveSchedulerIDs(ids...)
}

// Where appends a list predicates to the GenerationModelUpdate builder.
func (gmuo *GenerationModelUpdateOne) Where(ps ...predicate.GenerationModel) *GenerationModelUpdateOne {
	gmuo.mutation.Where(ps...)
	return gmuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (gmuo *GenerationModelUpdateOne) Select(field string, fields ...string) *GenerationModelUpdateOne {
	gmuo.fields = append([]string{field}, fields...)
	return gmuo
}

// Save executes the query and returns the updated GenerationModel entity.
func (gmuo *GenerationModelUpdateOne) Save(ctx context.Context) (*GenerationModel, error) {
	gmuo.defaults()
	return withHooks(ctx, gmuo.sqlSave, gmuo.mutation, gmuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (gmuo *GenerationModelUpdateOne) SaveX(ctx context.Context) *GenerationModel {
	node, err := gmuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (gmuo *GenerationModelUpdateOne) Exec(ctx context.Context) error {
	_, err := gmuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (gmuo *GenerationModelUpdateOne) ExecX(ctx context.Context) {
	if err := gmuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (gmuo *GenerationModelUpdateOne) defaults() {
	if _, ok := gmuo.mutation.UpdatedAt(); !ok {
		v := generationmodel.UpdateDefaultUpdatedAt()
		gmuo.mutation.SetUpdatedAt(v)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (gmuo *GenerationModelUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *GenerationModelUpdateOne {
	gmuo.modifiers = append(gmuo.modifiers, modifiers...)
	return gmuo
}

func (gmuo *GenerationModelUpdateOne) sqlSave(ctx context.Context) (_node *GenerationModel, err error) {
	_spec := sqlgraph.NewUpdateSpec(generationmodel.Table, generationmodel.Columns, sqlgraph.NewFieldSpec(generationmodel.FieldID, field.TypeUUID))
	id, ok := gmuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "GenerationModel.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := gmuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, generationmodel.FieldID)
		for _, f := range fields {
			if !generationmodel.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != generationmodel.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := gmuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := gmuo.mutation.NameInWorker(); ok {
		_spec.SetField(generationmodel.FieldNameInWorker, field.TypeString, value)
	}
	if value, ok := gmuo.mutation.IsActive(); ok {
		_spec.SetField(generationmodel.FieldIsActive, field.TypeBool, value)
	}
	if value, ok := gmuo.mutation.IsDefault(); ok {
		_spec.SetField(generationmodel.FieldIsDefault, field.TypeBool, value)
	}
	if value, ok := gmuo.mutation.IsHidden(); ok {
		_spec.SetField(generationmodel.FieldIsHidden, field.TypeBool, value)
	}
	if value, ok := gmuo.mutation.DisplayWeight(); ok {
		_spec.SetField(generationmodel.FieldDisplayWeight, field.TypeInt32, value)
	}
	if value, ok := gmuo.mutation.AddedDisplayWeight(); ok {
		_spec.AddField(generationmodel.FieldDisplayWeight, field.TypeInt32, value)
	}
	if value, ok := gmuo.mutation.DefaultSchedulerID(); ok {
		_spec.SetField(generationmodel.FieldDefaultSchedulerID, field.TypeUUID, value)
	}
	if gmuo.mutation.DefaultSchedulerIDCleared() {
		_spec.ClearField(generationmodel.FieldDefaultSchedulerID, field.TypeUUID)
	}
	if value, ok := gmuo.mutation.DefaultWidth(); ok {
		_spec.SetField(generationmodel.FieldDefaultWidth, field.TypeInt32, value)
	}
	if value, ok := gmuo.mutation.AddedDefaultWidth(); ok {
		_spec.AddField(generationmodel.FieldDefaultWidth, field.TypeInt32, value)
	}
	if value, ok := gmuo.mutation.DefaultHeight(); ok {
		_spec.SetField(generationmodel.FieldDefaultHeight, field.TypeInt32, value)
	}
	if value, ok := gmuo.mutation.AddedDefaultHeight(); ok {
		_spec.AddField(generationmodel.FieldDefaultHeight, field.TypeInt32, value)
	}
	if value, ok := gmuo.mutation.UpdatedAt(); ok {
		_spec.SetField(generationmodel.FieldUpdatedAt, field.TypeTime, value)
	}
	if gmuo.mutation.GenerationsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := gmuo.mutation.RemovedGenerationsIDs(); len(nodes) > 0 && !gmuo.mutation.GenerationsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := gmuo.mutation.GenerationsIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if gmuo.mutation.SchedulersCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := gmuo.mutation.RemovedSchedulersIDs(); len(nodes) > 0 && !gmuo.mutation.SchedulersCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := gmuo.mutation.SchedulersIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(gmuo.modifiers...)
	_node = &GenerationModel{config: gmuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, gmuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{generationmodel.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	gmuo.mutation.done = true
	return _node, nil
}
