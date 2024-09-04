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
	"github.com/stablecog/sc-go/database/ent/predicate"
	"github.com/stablecog/sc-go/database/ent/upscale"
	"github.com/stablecog/sc-go/database/ent/upscalemodel"
)

// UpscaleModelUpdate is the builder for updating UpscaleModel entities.
type UpscaleModelUpdate struct {
	config
	hooks     []Hook
	mutation  *UpscaleModelMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the UpscaleModelUpdate builder.
func (umu *UpscaleModelUpdate) Where(ps ...predicate.UpscaleModel) *UpscaleModelUpdate {
	umu.mutation.Where(ps...)
	return umu
}

// SetNameInWorker sets the "name_in_worker" field.
func (umu *UpscaleModelUpdate) SetNameInWorker(s string) *UpscaleModelUpdate {
	umu.mutation.SetNameInWorker(s)
	return umu
}

// SetNillableNameInWorker sets the "name_in_worker" field if the given value is not nil.
func (umu *UpscaleModelUpdate) SetNillableNameInWorker(s *string) *UpscaleModelUpdate {
	if s != nil {
		umu.SetNameInWorker(*s)
	}
	return umu
}

// SetIsActive sets the "is_active" field.
func (umu *UpscaleModelUpdate) SetIsActive(b bool) *UpscaleModelUpdate {
	umu.mutation.SetIsActive(b)
	return umu
}

// SetNillableIsActive sets the "is_active" field if the given value is not nil.
func (umu *UpscaleModelUpdate) SetNillableIsActive(b *bool) *UpscaleModelUpdate {
	if b != nil {
		umu.SetIsActive(*b)
	}
	return umu
}

// SetIsDefault sets the "is_default" field.
func (umu *UpscaleModelUpdate) SetIsDefault(b bool) *UpscaleModelUpdate {
	umu.mutation.SetIsDefault(b)
	return umu
}

// SetNillableIsDefault sets the "is_default" field if the given value is not nil.
func (umu *UpscaleModelUpdate) SetNillableIsDefault(b *bool) *UpscaleModelUpdate {
	if b != nil {
		umu.SetIsDefault(*b)
	}
	return umu
}

// SetIsHidden sets the "is_hidden" field.
func (umu *UpscaleModelUpdate) SetIsHidden(b bool) *UpscaleModelUpdate {
	umu.mutation.SetIsHidden(b)
	return umu
}

// SetNillableIsHidden sets the "is_hidden" field if the given value is not nil.
func (umu *UpscaleModelUpdate) SetNillableIsHidden(b *bool) *UpscaleModelUpdate {
	if b != nil {
		umu.SetIsHidden(*b)
	}
	return umu
}

// SetRunpodEndpoint sets the "runpod_endpoint" field.
func (umu *UpscaleModelUpdate) SetRunpodEndpoint(s string) *UpscaleModelUpdate {
	umu.mutation.SetRunpodEndpoint(s)
	return umu
}

// SetNillableRunpodEndpoint sets the "runpod_endpoint" field if the given value is not nil.
func (umu *UpscaleModelUpdate) SetNillableRunpodEndpoint(s *string) *UpscaleModelUpdate {
	if s != nil {
		umu.SetRunpodEndpoint(*s)
	}
	return umu
}

// ClearRunpodEndpoint clears the value of the "runpod_endpoint" field.
func (umu *UpscaleModelUpdate) ClearRunpodEndpoint() *UpscaleModelUpdate {
	umu.mutation.ClearRunpodEndpoint()
	return umu
}

// SetRunpodActive sets the "runpod_active" field.
func (umu *UpscaleModelUpdate) SetRunpodActive(b bool) *UpscaleModelUpdate {
	umu.mutation.SetRunpodActive(b)
	return umu
}

// SetNillableRunpodActive sets the "runpod_active" field if the given value is not nil.
func (umu *UpscaleModelUpdate) SetNillableRunpodActive(b *bool) *UpscaleModelUpdate {
	if b != nil {
		umu.SetRunpodActive(*b)
	}
	return umu
}

// SetUpdatedAt sets the "updated_at" field.
func (umu *UpscaleModelUpdate) SetUpdatedAt(t time.Time) *UpscaleModelUpdate {
	umu.mutation.SetUpdatedAt(t)
	return umu
}

// AddUpscaleIDs adds the "upscales" edge to the Upscale entity by IDs.
func (umu *UpscaleModelUpdate) AddUpscaleIDs(ids ...uuid.UUID) *UpscaleModelUpdate {
	umu.mutation.AddUpscaleIDs(ids...)
	return umu
}

// AddUpscales adds the "upscales" edges to the Upscale entity.
func (umu *UpscaleModelUpdate) AddUpscales(u ...*Upscale) *UpscaleModelUpdate {
	ids := make([]uuid.UUID, len(u))
	for i := range u {
		ids[i] = u[i].ID
	}
	return umu.AddUpscaleIDs(ids...)
}

// Mutation returns the UpscaleModelMutation object of the builder.
func (umu *UpscaleModelUpdate) Mutation() *UpscaleModelMutation {
	return umu.mutation
}

// ClearUpscales clears all "upscales" edges to the Upscale entity.
func (umu *UpscaleModelUpdate) ClearUpscales() *UpscaleModelUpdate {
	umu.mutation.ClearUpscales()
	return umu
}

// RemoveUpscaleIDs removes the "upscales" edge to Upscale entities by IDs.
func (umu *UpscaleModelUpdate) RemoveUpscaleIDs(ids ...uuid.UUID) *UpscaleModelUpdate {
	umu.mutation.RemoveUpscaleIDs(ids...)
	return umu
}

// RemoveUpscales removes "upscales" edges to Upscale entities.
func (umu *UpscaleModelUpdate) RemoveUpscales(u ...*Upscale) *UpscaleModelUpdate {
	ids := make([]uuid.UUID, len(u))
	for i := range u {
		ids[i] = u[i].ID
	}
	return umu.RemoveUpscaleIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (umu *UpscaleModelUpdate) Save(ctx context.Context) (int, error) {
	umu.defaults()
	return withHooks(ctx, umu.sqlSave, umu.mutation, umu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (umu *UpscaleModelUpdate) SaveX(ctx context.Context) int {
	affected, err := umu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (umu *UpscaleModelUpdate) Exec(ctx context.Context) error {
	_, err := umu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (umu *UpscaleModelUpdate) ExecX(ctx context.Context) {
	if err := umu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (umu *UpscaleModelUpdate) defaults() {
	if _, ok := umu.mutation.UpdatedAt(); !ok {
		v := upscalemodel.UpdateDefaultUpdatedAt()
		umu.mutation.SetUpdatedAt(v)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (umu *UpscaleModelUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *UpscaleModelUpdate {
	umu.modifiers = append(umu.modifiers, modifiers...)
	return umu
}

func (umu *UpscaleModelUpdate) sqlSave(ctx context.Context) (n int, err error) {
	_spec := sqlgraph.NewUpdateSpec(upscalemodel.Table, upscalemodel.Columns, sqlgraph.NewFieldSpec(upscalemodel.FieldID, field.TypeUUID))
	if ps := umu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := umu.mutation.NameInWorker(); ok {
		_spec.SetField(upscalemodel.FieldNameInWorker, field.TypeString, value)
	}
	if value, ok := umu.mutation.IsActive(); ok {
		_spec.SetField(upscalemodel.FieldIsActive, field.TypeBool, value)
	}
	if value, ok := umu.mutation.IsDefault(); ok {
		_spec.SetField(upscalemodel.FieldIsDefault, field.TypeBool, value)
	}
	if value, ok := umu.mutation.IsHidden(); ok {
		_spec.SetField(upscalemodel.FieldIsHidden, field.TypeBool, value)
	}
	if value, ok := umu.mutation.RunpodEndpoint(); ok {
		_spec.SetField(upscalemodel.FieldRunpodEndpoint, field.TypeString, value)
	}
	if umu.mutation.RunpodEndpointCleared() {
		_spec.ClearField(upscalemodel.FieldRunpodEndpoint, field.TypeString)
	}
	if value, ok := umu.mutation.RunpodActive(); ok {
		_spec.SetField(upscalemodel.FieldRunpodActive, field.TypeBool, value)
	}
	if value, ok := umu.mutation.UpdatedAt(); ok {
		_spec.SetField(upscalemodel.FieldUpdatedAt, field.TypeTime, value)
	}
	if umu.mutation.UpscalesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   upscalemodel.UpscalesTable,
			Columns: []string{upscalemodel.UpscalesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(upscale.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := umu.mutation.RemovedUpscalesIDs(); len(nodes) > 0 && !umu.mutation.UpscalesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   upscalemodel.UpscalesTable,
			Columns: []string{upscalemodel.UpscalesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(upscale.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := umu.mutation.UpscalesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   upscalemodel.UpscalesTable,
			Columns: []string{upscalemodel.UpscalesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(upscale.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(umu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, umu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{upscalemodel.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	umu.mutation.done = true
	return n, nil
}

// UpscaleModelUpdateOne is the builder for updating a single UpscaleModel entity.
type UpscaleModelUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *UpscaleModelMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetNameInWorker sets the "name_in_worker" field.
func (umuo *UpscaleModelUpdateOne) SetNameInWorker(s string) *UpscaleModelUpdateOne {
	umuo.mutation.SetNameInWorker(s)
	return umuo
}

// SetNillableNameInWorker sets the "name_in_worker" field if the given value is not nil.
func (umuo *UpscaleModelUpdateOne) SetNillableNameInWorker(s *string) *UpscaleModelUpdateOne {
	if s != nil {
		umuo.SetNameInWorker(*s)
	}
	return umuo
}

// SetIsActive sets the "is_active" field.
func (umuo *UpscaleModelUpdateOne) SetIsActive(b bool) *UpscaleModelUpdateOne {
	umuo.mutation.SetIsActive(b)
	return umuo
}

// SetNillableIsActive sets the "is_active" field if the given value is not nil.
func (umuo *UpscaleModelUpdateOne) SetNillableIsActive(b *bool) *UpscaleModelUpdateOne {
	if b != nil {
		umuo.SetIsActive(*b)
	}
	return umuo
}

// SetIsDefault sets the "is_default" field.
func (umuo *UpscaleModelUpdateOne) SetIsDefault(b bool) *UpscaleModelUpdateOne {
	umuo.mutation.SetIsDefault(b)
	return umuo
}

// SetNillableIsDefault sets the "is_default" field if the given value is not nil.
func (umuo *UpscaleModelUpdateOne) SetNillableIsDefault(b *bool) *UpscaleModelUpdateOne {
	if b != nil {
		umuo.SetIsDefault(*b)
	}
	return umuo
}

// SetIsHidden sets the "is_hidden" field.
func (umuo *UpscaleModelUpdateOne) SetIsHidden(b bool) *UpscaleModelUpdateOne {
	umuo.mutation.SetIsHidden(b)
	return umuo
}

// SetNillableIsHidden sets the "is_hidden" field if the given value is not nil.
func (umuo *UpscaleModelUpdateOne) SetNillableIsHidden(b *bool) *UpscaleModelUpdateOne {
	if b != nil {
		umuo.SetIsHidden(*b)
	}
	return umuo
}

// SetRunpodEndpoint sets the "runpod_endpoint" field.
func (umuo *UpscaleModelUpdateOne) SetRunpodEndpoint(s string) *UpscaleModelUpdateOne {
	umuo.mutation.SetRunpodEndpoint(s)
	return umuo
}

// SetNillableRunpodEndpoint sets the "runpod_endpoint" field if the given value is not nil.
func (umuo *UpscaleModelUpdateOne) SetNillableRunpodEndpoint(s *string) *UpscaleModelUpdateOne {
	if s != nil {
		umuo.SetRunpodEndpoint(*s)
	}
	return umuo
}

// ClearRunpodEndpoint clears the value of the "runpod_endpoint" field.
func (umuo *UpscaleModelUpdateOne) ClearRunpodEndpoint() *UpscaleModelUpdateOne {
	umuo.mutation.ClearRunpodEndpoint()
	return umuo
}

// SetRunpodActive sets the "runpod_active" field.
func (umuo *UpscaleModelUpdateOne) SetRunpodActive(b bool) *UpscaleModelUpdateOne {
	umuo.mutation.SetRunpodActive(b)
	return umuo
}

// SetNillableRunpodActive sets the "runpod_active" field if the given value is not nil.
func (umuo *UpscaleModelUpdateOne) SetNillableRunpodActive(b *bool) *UpscaleModelUpdateOne {
	if b != nil {
		umuo.SetRunpodActive(*b)
	}
	return umuo
}

// SetUpdatedAt sets the "updated_at" field.
func (umuo *UpscaleModelUpdateOne) SetUpdatedAt(t time.Time) *UpscaleModelUpdateOne {
	umuo.mutation.SetUpdatedAt(t)
	return umuo
}

// AddUpscaleIDs adds the "upscales" edge to the Upscale entity by IDs.
func (umuo *UpscaleModelUpdateOne) AddUpscaleIDs(ids ...uuid.UUID) *UpscaleModelUpdateOne {
	umuo.mutation.AddUpscaleIDs(ids...)
	return umuo
}

// AddUpscales adds the "upscales" edges to the Upscale entity.
func (umuo *UpscaleModelUpdateOne) AddUpscales(u ...*Upscale) *UpscaleModelUpdateOne {
	ids := make([]uuid.UUID, len(u))
	for i := range u {
		ids[i] = u[i].ID
	}
	return umuo.AddUpscaleIDs(ids...)
}

// Mutation returns the UpscaleModelMutation object of the builder.
func (umuo *UpscaleModelUpdateOne) Mutation() *UpscaleModelMutation {
	return umuo.mutation
}

// ClearUpscales clears all "upscales" edges to the Upscale entity.
func (umuo *UpscaleModelUpdateOne) ClearUpscales() *UpscaleModelUpdateOne {
	umuo.mutation.ClearUpscales()
	return umuo
}

// RemoveUpscaleIDs removes the "upscales" edge to Upscale entities by IDs.
func (umuo *UpscaleModelUpdateOne) RemoveUpscaleIDs(ids ...uuid.UUID) *UpscaleModelUpdateOne {
	umuo.mutation.RemoveUpscaleIDs(ids...)
	return umuo
}

// RemoveUpscales removes "upscales" edges to Upscale entities.
func (umuo *UpscaleModelUpdateOne) RemoveUpscales(u ...*Upscale) *UpscaleModelUpdateOne {
	ids := make([]uuid.UUID, len(u))
	for i := range u {
		ids[i] = u[i].ID
	}
	return umuo.RemoveUpscaleIDs(ids...)
}

// Where appends a list predicates to the UpscaleModelUpdate builder.
func (umuo *UpscaleModelUpdateOne) Where(ps ...predicate.UpscaleModel) *UpscaleModelUpdateOne {
	umuo.mutation.Where(ps...)
	return umuo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (umuo *UpscaleModelUpdateOne) Select(field string, fields ...string) *UpscaleModelUpdateOne {
	umuo.fields = append([]string{field}, fields...)
	return umuo
}

// Save executes the query and returns the updated UpscaleModel entity.
func (umuo *UpscaleModelUpdateOne) Save(ctx context.Context) (*UpscaleModel, error) {
	umuo.defaults()
	return withHooks(ctx, umuo.sqlSave, umuo.mutation, umuo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (umuo *UpscaleModelUpdateOne) SaveX(ctx context.Context) *UpscaleModel {
	node, err := umuo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (umuo *UpscaleModelUpdateOne) Exec(ctx context.Context) error {
	_, err := umuo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (umuo *UpscaleModelUpdateOne) ExecX(ctx context.Context) {
	if err := umuo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (umuo *UpscaleModelUpdateOne) defaults() {
	if _, ok := umuo.mutation.UpdatedAt(); !ok {
		v := upscalemodel.UpdateDefaultUpdatedAt()
		umuo.mutation.SetUpdatedAt(v)
	}
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (umuo *UpscaleModelUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *UpscaleModelUpdateOne {
	umuo.modifiers = append(umuo.modifiers, modifiers...)
	return umuo
}

func (umuo *UpscaleModelUpdateOne) sqlSave(ctx context.Context) (_node *UpscaleModel, err error) {
	_spec := sqlgraph.NewUpdateSpec(upscalemodel.Table, upscalemodel.Columns, sqlgraph.NewFieldSpec(upscalemodel.FieldID, field.TypeUUID))
	id, ok := umuo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "UpscaleModel.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := umuo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, upscalemodel.FieldID)
		for _, f := range fields {
			if !upscalemodel.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != upscalemodel.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := umuo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := umuo.mutation.NameInWorker(); ok {
		_spec.SetField(upscalemodel.FieldNameInWorker, field.TypeString, value)
	}
	if value, ok := umuo.mutation.IsActive(); ok {
		_spec.SetField(upscalemodel.FieldIsActive, field.TypeBool, value)
	}
	if value, ok := umuo.mutation.IsDefault(); ok {
		_spec.SetField(upscalemodel.FieldIsDefault, field.TypeBool, value)
	}
	if value, ok := umuo.mutation.IsHidden(); ok {
		_spec.SetField(upscalemodel.FieldIsHidden, field.TypeBool, value)
	}
	if value, ok := umuo.mutation.RunpodEndpoint(); ok {
		_spec.SetField(upscalemodel.FieldRunpodEndpoint, field.TypeString, value)
	}
	if umuo.mutation.RunpodEndpointCleared() {
		_spec.ClearField(upscalemodel.FieldRunpodEndpoint, field.TypeString)
	}
	if value, ok := umuo.mutation.RunpodActive(); ok {
		_spec.SetField(upscalemodel.FieldRunpodActive, field.TypeBool, value)
	}
	if value, ok := umuo.mutation.UpdatedAt(); ok {
		_spec.SetField(upscalemodel.FieldUpdatedAt, field.TypeTime, value)
	}
	if umuo.mutation.UpscalesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   upscalemodel.UpscalesTable,
			Columns: []string{upscalemodel.UpscalesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(upscale.FieldID, field.TypeUUID),
			},
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := umuo.mutation.RemovedUpscalesIDs(); len(nodes) > 0 && !umuo.mutation.UpscalesCleared() {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   upscalemodel.UpscalesTable,
			Columns: []string{upscalemodel.UpscalesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(upscale.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := umuo.mutation.UpscalesIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   upscalemodel.UpscalesTable,
			Columns: []string{upscalemodel.UpscalesColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: sqlgraph.NewFieldSpec(upscale.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(umuo.modifiers...)
	_node = &UpscaleModel{config: umuo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, umuo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{upscalemodel.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	umuo.mutation.done = true
	return _node, nil
}
