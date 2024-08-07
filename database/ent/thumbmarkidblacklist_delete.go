// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/stablecog/sc-go/database/ent/predicate"
	"github.com/stablecog/sc-go/database/ent/thumbmarkidblacklist"
)

// ThumbmarkIdBlackListDelete is the builder for deleting a ThumbmarkIdBlackList entity.
type ThumbmarkIdBlackListDelete struct {
	config
	hooks    []Hook
	mutation *ThumbmarkIdBlackListMutation
}

// Where appends a list predicates to the ThumbmarkIdBlackListDelete builder.
func (tibld *ThumbmarkIdBlackListDelete) Where(ps ...predicate.ThumbmarkIdBlackList) *ThumbmarkIdBlackListDelete {
	tibld.mutation.Where(ps...)
	return tibld
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (tibld *ThumbmarkIdBlackListDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, tibld.sqlExec, tibld.mutation, tibld.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (tibld *ThumbmarkIdBlackListDelete) ExecX(ctx context.Context) int {
	n, err := tibld.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (tibld *ThumbmarkIdBlackListDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(thumbmarkidblacklist.Table, sqlgraph.NewFieldSpec(thumbmarkidblacklist.FieldID, field.TypeUUID))
	if ps := tibld.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, tibld.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	tibld.mutation.done = true
	return affected, err
}

// ThumbmarkIdBlackListDeleteOne is the builder for deleting a single ThumbmarkIdBlackList entity.
type ThumbmarkIdBlackListDeleteOne struct {
	tibld *ThumbmarkIdBlackListDelete
}

// Where appends a list predicates to the ThumbmarkIdBlackListDelete builder.
func (tibldo *ThumbmarkIdBlackListDeleteOne) Where(ps ...predicate.ThumbmarkIdBlackList) *ThumbmarkIdBlackListDeleteOne {
	tibldo.tibld.mutation.Where(ps...)
	return tibldo
}

// Exec executes the deletion query.
func (tibldo *ThumbmarkIdBlackListDeleteOne) Exec(ctx context.Context) error {
	n, err := tibldo.tibld.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{thumbmarkidblacklist.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (tibldo *ThumbmarkIdBlackListDeleteOne) ExecX(ctx context.Context) {
	if err := tibldo.Exec(ctx); err != nil {
		panic(err)
	}
}
