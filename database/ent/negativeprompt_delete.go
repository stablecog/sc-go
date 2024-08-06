// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/stablecog/sc-go/database/ent/negativeprompt"
	"github.com/stablecog/sc-go/database/ent/predicate"
)

// NegativePromptDelete is the builder for deleting a NegativePrompt entity.
type NegativePromptDelete struct {
	config
	hooks    []Hook
	mutation *NegativePromptMutation
}

// Where appends a list predicates to the NegativePromptDelete builder.
func (npd *NegativePromptDelete) Where(ps ...predicate.NegativePrompt) *NegativePromptDelete {
	npd.mutation.Where(ps...)
	return npd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (npd *NegativePromptDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, npd.sqlExec, npd.mutation, npd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (npd *NegativePromptDelete) ExecX(ctx context.Context) int {
	n, err := npd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (npd *NegativePromptDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(negativeprompt.Table, sqlgraph.NewFieldSpec(negativeprompt.FieldID, field.TypeUUID))
	if ps := npd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, npd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	npd.mutation.done = true
	return affected, err
}

// NegativePromptDeleteOne is the builder for deleting a single NegativePrompt entity.
type NegativePromptDeleteOne struct {
	npd *NegativePromptDelete
}

// Where appends a list predicates to the NegativePromptDelete builder.
func (npdo *NegativePromptDeleteOne) Where(ps ...predicate.NegativePrompt) *NegativePromptDeleteOne {
	npdo.npd.mutation.Where(ps...)
	return npdo
}

// Exec executes the deletion query.
func (npdo *NegativePromptDeleteOne) Exec(ctx context.Context) error {
	n, err := npdo.npd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{negativeprompt.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (npdo *NegativePromptDeleteOne) ExecX(ctx context.Context) {
	if err := npdo.Exec(ctx); err != nil {
		panic(err)
	}
}
