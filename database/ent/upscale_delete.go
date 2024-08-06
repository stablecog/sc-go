// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/stablecog/sc-go/database/ent/predicate"
	"github.com/stablecog/sc-go/database/ent/upscale"
)

// UpscaleDelete is the builder for deleting a Upscale entity.
type UpscaleDelete struct {
	config
	hooks    []Hook
	mutation *UpscaleMutation
}

// Where appends a list predicates to the UpscaleDelete builder.
func (ud *UpscaleDelete) Where(ps ...predicate.Upscale) *UpscaleDelete {
	ud.mutation.Where(ps...)
	return ud
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (ud *UpscaleDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, ud.sqlExec, ud.mutation, ud.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (ud *UpscaleDelete) ExecX(ctx context.Context) int {
	n, err := ud.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (ud *UpscaleDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(upscale.Table, sqlgraph.NewFieldSpec(upscale.FieldID, field.TypeUUID))
	if ps := ud.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, ud.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	ud.mutation.done = true
	return affected, err
}

// UpscaleDeleteOne is the builder for deleting a single Upscale entity.
type UpscaleDeleteOne struct {
	ud *UpscaleDelete
}

// Where appends a list predicates to the UpscaleDelete builder.
func (udo *UpscaleDeleteOne) Where(ps ...predicate.Upscale) *UpscaleDeleteOne {
	udo.ud.mutation.Where(ps...)
	return udo
}

// Exec executes the deletion query.
func (udo *UpscaleDeleteOne) Exec(ctx context.Context) error {
	n, err := udo.ud.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{upscale.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (udo *UpscaleDeleteOne) ExecX(ctx context.Context) {
	if err := udo.Exec(ctx); err != nil {
		panic(err)
	}
}
