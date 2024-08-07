// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/stablecog/sc-go/database/ent/generationmodel"
	"github.com/stablecog/sc-go/database/ent/predicate"
)

// GenerationModelDelete is the builder for deleting a GenerationModel entity.
type GenerationModelDelete struct {
	config
	hooks    []Hook
	mutation *GenerationModelMutation
}

// Where appends a list predicates to the GenerationModelDelete builder.
func (gmd *GenerationModelDelete) Where(ps ...predicate.GenerationModel) *GenerationModelDelete {
	gmd.mutation.Where(ps...)
	return gmd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (gmd *GenerationModelDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, gmd.sqlExec, gmd.mutation, gmd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (gmd *GenerationModelDelete) ExecX(ctx context.Context) int {
	n, err := gmd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (gmd *GenerationModelDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(generationmodel.Table, sqlgraph.NewFieldSpec(generationmodel.FieldID, field.TypeUUID))
	if ps := gmd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, gmd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	gmd.mutation.done = true
	return affected, err
}

// GenerationModelDeleteOne is the builder for deleting a single GenerationModel entity.
type GenerationModelDeleteOne struct {
	gmd *GenerationModelDelete
}

// Where appends a list predicates to the GenerationModelDelete builder.
func (gmdo *GenerationModelDeleteOne) Where(ps ...predicate.GenerationModel) *GenerationModelDeleteOne {
	gmdo.gmd.mutation.Where(ps...)
	return gmdo
}

// Exec executes the deletion query.
func (gmdo *GenerationModelDeleteOne) Exec(ctx context.Context) error {
	n, err := gmdo.gmd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{generationmodel.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (gmdo *GenerationModelDeleteOne) ExecX(ctx context.Context) {
	if err := gmdo.Exec(ctx); err != nil {
		panic(err)
	}
}
