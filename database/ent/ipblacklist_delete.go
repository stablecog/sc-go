// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/stablecog/sc-go/database/ent/ipblacklist"
	"github.com/stablecog/sc-go/database/ent/predicate"
)

// IPBlackListDelete is the builder for deleting a IPBlackList entity.
type IPBlackListDelete struct {
	config
	hooks    []Hook
	mutation *IPBlackListMutation
}

// Where appends a list predicates to the IPBlackListDelete builder.
func (ibld *IPBlackListDelete) Where(ps ...predicate.IPBlackList) *IPBlackListDelete {
	ibld.mutation.Where(ps...)
	return ibld
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (ibld *IPBlackListDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, ibld.sqlExec, ibld.mutation, ibld.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (ibld *IPBlackListDelete) ExecX(ctx context.Context) int {
	n, err := ibld.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (ibld *IPBlackListDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(ipblacklist.Table, sqlgraph.NewFieldSpec(ipblacklist.FieldID, field.TypeUUID))
	if ps := ibld.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, ibld.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	ibld.mutation.done = true
	return affected, err
}

// IPBlackListDeleteOne is the builder for deleting a single IPBlackList entity.
type IPBlackListDeleteOne struct {
	ibld *IPBlackListDelete
}

// Where appends a list predicates to the IPBlackListDelete builder.
func (ibldo *IPBlackListDeleteOne) Where(ps ...predicate.IPBlackList) *IPBlackListDeleteOne {
	ibldo.ibld.mutation.Where(ps...)
	return ibldo
}

// Exec executes the deletion query.
func (ibldo *IPBlackListDeleteOne) Exec(ctx context.Context) error {
	n, err := ibldo.ibld.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{ipblacklist.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (ibldo *IPBlackListDeleteOne) ExecX(ctx context.Context) {
	if err := ibldo.Exec(ctx); err != nil {
		panic(err)
	}
}
