// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/stablecog/sc-go/database/ent/predicate"
	"github.com/stablecog/sc-go/database/ent/upscaleoutput"
)

// UpscaleOutputDelete is the builder for deleting a UpscaleOutput entity.
type UpscaleOutputDelete struct {
	config
	hooks    []Hook
	mutation *UpscaleOutputMutation
}

// Where appends a list predicates to the UpscaleOutputDelete builder.
func (uod *UpscaleOutputDelete) Where(ps ...predicate.UpscaleOutput) *UpscaleOutputDelete {
	uod.mutation.Where(ps...)
	return uod
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (uod *UpscaleOutputDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, uod.sqlExec, uod.mutation, uod.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (uod *UpscaleOutputDelete) ExecX(ctx context.Context) int {
	n, err := uod.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (uod *UpscaleOutputDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(upscaleoutput.Table, sqlgraph.NewFieldSpec(upscaleoutput.FieldID, field.TypeUUID))
	if ps := uod.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, uod.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	uod.mutation.done = true
	return affected, err
}

// UpscaleOutputDeleteOne is the builder for deleting a single UpscaleOutput entity.
type UpscaleOutputDeleteOne struct {
	uod *UpscaleOutputDelete
}

// Where appends a list predicates to the UpscaleOutputDelete builder.
func (uodo *UpscaleOutputDeleteOne) Where(ps ...predicate.UpscaleOutput) *UpscaleOutputDeleteOne {
	uodo.uod.mutation.Where(ps...)
	return uodo
}

// Exec executes the deletion query.
func (uodo *UpscaleOutputDeleteOne) Exec(ctx context.Context) error {
	n, err := uodo.uod.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{upscaleoutput.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (uodo *UpscaleOutputDeleteOne) ExecX(ctx context.Context) {
	if err := uodo.Exec(ctx); err != nil {
		panic(err)
	}
}
