// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/stablecog/sc-go/database/ent/predicate"
	"github.com/stablecog/sc-go/database/ent/voiceoverspeaker"
)

// VoiceoverSpeakerDelete is the builder for deleting a VoiceoverSpeaker entity.
type VoiceoverSpeakerDelete struct {
	config
	hooks    []Hook
	mutation *VoiceoverSpeakerMutation
}

// Where appends a list predicates to the VoiceoverSpeakerDelete builder.
func (vsd *VoiceoverSpeakerDelete) Where(ps ...predicate.VoiceoverSpeaker) *VoiceoverSpeakerDelete {
	vsd.mutation.Where(ps...)
	return vsd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (vsd *VoiceoverSpeakerDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, vsd.sqlExec, vsd.mutation, vsd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (vsd *VoiceoverSpeakerDelete) ExecX(ctx context.Context) int {
	n, err := vsd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (vsd *VoiceoverSpeakerDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(voiceoverspeaker.Table, sqlgraph.NewFieldSpec(voiceoverspeaker.FieldID, field.TypeUUID))
	if ps := vsd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, vsd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	vsd.mutation.done = true
	return affected, err
}

// VoiceoverSpeakerDeleteOne is the builder for deleting a single VoiceoverSpeaker entity.
type VoiceoverSpeakerDeleteOne struct {
	vsd *VoiceoverSpeakerDelete
}

// Where appends a list predicates to the VoiceoverSpeakerDelete builder.
func (vsdo *VoiceoverSpeakerDeleteOne) Where(ps ...predicate.VoiceoverSpeaker) *VoiceoverSpeakerDeleteOne {
	vsdo.vsd.mutation.Where(ps...)
	return vsdo
}

// Exec executes the deletion query.
func (vsdo *VoiceoverSpeakerDeleteOne) Exec(ctx context.Context) error {
	n, err := vsdo.vsd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{voiceoverspeaker.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (vsdo *VoiceoverSpeakerDeleteOne) ExecX(ctx context.Context) {
	if err := vsdo.Exec(ctx); err != nil {
		panic(err)
	}
}
