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
	"github.com/stablecog/go-apps/database/ent/upscalerealtime"
)

// UpscaleRealtimeCreate is the builder for creating a UpscaleRealtime entity.
type UpscaleRealtimeCreate struct {
	config
	mutation *UpscaleRealtimeMutation
	hooks    []Hook
}

// SetStatus sets the "status" field.
func (urc *UpscaleRealtimeCreate) SetStatus(u upscalerealtime.Status) *UpscaleRealtimeCreate {
	urc.mutation.SetStatus(u)
	return urc
}

// SetCountryCode sets the "country_code" field.
func (urc *UpscaleRealtimeCreate) SetCountryCode(s string) *UpscaleRealtimeCreate {
	urc.mutation.SetCountryCode(s)
	return urc
}

// SetUsesDefaultServer sets the "uses_default_server" field.
func (urc *UpscaleRealtimeCreate) SetUsesDefaultServer(b bool) *UpscaleRealtimeCreate {
	urc.mutation.SetUsesDefaultServer(b)
	return urc
}

// SetWidth sets the "width" field.
func (urc *UpscaleRealtimeCreate) SetWidth(i int) *UpscaleRealtimeCreate {
	urc.mutation.SetWidth(i)
	return urc
}

// SetHeight sets the "height" field.
func (urc *UpscaleRealtimeCreate) SetHeight(i int) *UpscaleRealtimeCreate {
	urc.mutation.SetHeight(i)
	return urc
}

// SetScale sets the "scale" field.
func (urc *UpscaleRealtimeCreate) SetScale(i int) *UpscaleRealtimeCreate {
	urc.mutation.SetScale(i)
	return urc
}

// SetCreatedAt sets the "created_at" field.
func (urc *UpscaleRealtimeCreate) SetCreatedAt(t time.Time) *UpscaleRealtimeCreate {
	urc.mutation.SetCreatedAt(t)
	return urc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (urc *UpscaleRealtimeCreate) SetNillableCreatedAt(t *time.Time) *UpscaleRealtimeCreate {
	if t != nil {
		urc.SetCreatedAt(*t)
	}
	return urc
}

// SetUpdatedAt sets the "updated_at" field.
func (urc *UpscaleRealtimeCreate) SetUpdatedAt(t time.Time) *UpscaleRealtimeCreate {
	urc.mutation.SetUpdatedAt(t)
	return urc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (urc *UpscaleRealtimeCreate) SetNillableUpdatedAt(t *time.Time) *UpscaleRealtimeCreate {
	if t != nil {
		urc.SetUpdatedAt(*t)
	}
	return urc
}

// SetUserTier sets the "user_tier" field.
func (urc *UpscaleRealtimeCreate) SetUserTier(ut upscalerealtime.UserTier) *UpscaleRealtimeCreate {
	urc.mutation.SetUserTier(ut)
	return urc
}

// SetNillableUserTier sets the "user_tier" field if the given value is not nil.
func (urc *UpscaleRealtimeCreate) SetNillableUserTier(ut *upscalerealtime.UserTier) *UpscaleRealtimeCreate {
	if ut != nil {
		urc.SetUserTier(*ut)
	}
	return urc
}

// SetID sets the "id" field.
func (urc *UpscaleRealtimeCreate) SetID(u uuid.UUID) *UpscaleRealtimeCreate {
	urc.mutation.SetID(u)
	return urc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (urc *UpscaleRealtimeCreate) SetNillableID(u *uuid.UUID) *UpscaleRealtimeCreate {
	if u != nil {
		urc.SetID(*u)
	}
	return urc
}

// Mutation returns the UpscaleRealtimeMutation object of the builder.
func (urc *UpscaleRealtimeCreate) Mutation() *UpscaleRealtimeMutation {
	return urc.mutation
}

// Save creates the UpscaleRealtime in the database.
func (urc *UpscaleRealtimeCreate) Save(ctx context.Context) (*UpscaleRealtime, error) {
	urc.defaults()
	return withHooks[*UpscaleRealtime, UpscaleRealtimeMutation](ctx, urc.sqlSave, urc.mutation, urc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (urc *UpscaleRealtimeCreate) SaveX(ctx context.Context) *UpscaleRealtime {
	v, err := urc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (urc *UpscaleRealtimeCreate) Exec(ctx context.Context) error {
	_, err := urc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (urc *UpscaleRealtimeCreate) ExecX(ctx context.Context) {
	if err := urc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (urc *UpscaleRealtimeCreate) defaults() {
	if _, ok := urc.mutation.CreatedAt(); !ok {
		v := upscalerealtime.DefaultCreatedAt()
		urc.mutation.SetCreatedAt(v)
	}
	if _, ok := urc.mutation.UpdatedAt(); !ok {
		v := upscalerealtime.DefaultUpdatedAt()
		urc.mutation.SetUpdatedAt(v)
	}
	if _, ok := urc.mutation.UserTier(); !ok {
		v := upscalerealtime.DefaultUserTier
		urc.mutation.SetUserTier(v)
	}
	if _, ok := urc.mutation.ID(); !ok {
		v := upscalerealtime.DefaultID()
		urc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (urc *UpscaleRealtimeCreate) check() error {
	if _, ok := urc.mutation.Status(); !ok {
		return &ValidationError{Name: "status", err: errors.New(`ent: missing required field "UpscaleRealtime.status"`)}
	}
	if v, ok := urc.mutation.Status(); ok {
		if err := upscalerealtime.StatusValidator(v); err != nil {
			return &ValidationError{Name: "status", err: fmt.Errorf(`ent: validator failed for field "UpscaleRealtime.status": %w`, err)}
		}
	}
	if _, ok := urc.mutation.CountryCode(); !ok {
		return &ValidationError{Name: "country_code", err: errors.New(`ent: missing required field "UpscaleRealtime.country_code"`)}
	}
	if _, ok := urc.mutation.UsesDefaultServer(); !ok {
		return &ValidationError{Name: "uses_default_server", err: errors.New(`ent: missing required field "UpscaleRealtime.uses_default_server"`)}
	}
	if _, ok := urc.mutation.Width(); !ok {
		return &ValidationError{Name: "width", err: errors.New(`ent: missing required field "UpscaleRealtime.width"`)}
	}
	if _, ok := urc.mutation.Height(); !ok {
		return &ValidationError{Name: "height", err: errors.New(`ent: missing required field "UpscaleRealtime.height"`)}
	}
	if _, ok := urc.mutation.Scale(); !ok {
		return &ValidationError{Name: "scale", err: errors.New(`ent: missing required field "UpscaleRealtime.scale"`)}
	}
	if _, ok := urc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "UpscaleRealtime.created_at"`)}
	}
	if _, ok := urc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "UpscaleRealtime.updated_at"`)}
	}
	if _, ok := urc.mutation.UserTier(); !ok {
		return &ValidationError{Name: "user_tier", err: errors.New(`ent: missing required field "UpscaleRealtime.user_tier"`)}
	}
	if v, ok := urc.mutation.UserTier(); ok {
		if err := upscalerealtime.UserTierValidator(v); err != nil {
			return &ValidationError{Name: "user_tier", err: fmt.Errorf(`ent: validator failed for field "UpscaleRealtime.user_tier": %w`, err)}
		}
	}
	return nil
}

func (urc *UpscaleRealtimeCreate) sqlSave(ctx context.Context) (*UpscaleRealtime, error) {
	if err := urc.check(); err != nil {
		return nil, err
	}
	_node, _spec := urc.createSpec()
	if err := sqlgraph.CreateNode(ctx, urc.driver, _spec); err != nil {
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
	urc.mutation.id = &_node.ID
	urc.mutation.done = true
	return _node, nil
}

func (urc *UpscaleRealtimeCreate) createSpec() (*UpscaleRealtime, *sqlgraph.CreateSpec) {
	var (
		_node = &UpscaleRealtime{config: urc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: upscalerealtime.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: upscalerealtime.FieldID,
			},
		}
	)
	if id, ok := urc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := urc.mutation.Status(); ok {
		_spec.SetField(upscalerealtime.FieldStatus, field.TypeEnum, value)
		_node.Status = value
	}
	if value, ok := urc.mutation.CountryCode(); ok {
		_spec.SetField(upscalerealtime.FieldCountryCode, field.TypeString, value)
		_node.CountryCode = &value
	}
	if value, ok := urc.mutation.UsesDefaultServer(); ok {
		_spec.SetField(upscalerealtime.FieldUsesDefaultServer, field.TypeBool, value)
		_node.UsesDefaultServer = value
	}
	if value, ok := urc.mutation.Width(); ok {
		_spec.SetField(upscalerealtime.FieldWidth, field.TypeInt, value)
		_node.Width = &value
	}
	if value, ok := urc.mutation.Height(); ok {
		_spec.SetField(upscalerealtime.FieldHeight, field.TypeInt, value)
		_node.Height = &value
	}
	if value, ok := urc.mutation.Scale(); ok {
		_spec.SetField(upscalerealtime.FieldScale, field.TypeInt, value)
		_node.Scale = &value
	}
	if value, ok := urc.mutation.CreatedAt(); ok {
		_spec.SetField(upscalerealtime.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := urc.mutation.UpdatedAt(); ok {
		_spec.SetField(upscalerealtime.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if value, ok := urc.mutation.UserTier(); ok {
		_spec.SetField(upscalerealtime.FieldUserTier, field.TypeEnum, value)
		_node.UserTier = value
	}
	return _node, _spec
}

// UpscaleRealtimeCreateBulk is the builder for creating many UpscaleRealtime entities in bulk.
type UpscaleRealtimeCreateBulk struct {
	config
	builders []*UpscaleRealtimeCreate
}

// Save creates the UpscaleRealtime entities in the database.
func (urcb *UpscaleRealtimeCreateBulk) Save(ctx context.Context) ([]*UpscaleRealtime, error) {
	specs := make([]*sqlgraph.CreateSpec, len(urcb.builders))
	nodes := make([]*UpscaleRealtime, len(urcb.builders))
	mutators := make([]Mutator, len(urcb.builders))
	for i := range urcb.builders {
		func(i int, root context.Context) {
			builder := urcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*UpscaleRealtimeMutation)
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
					_, err = mutators[i+1].Mutate(root, urcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, urcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, urcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (urcb *UpscaleRealtimeCreateBulk) SaveX(ctx context.Context) []*UpscaleRealtime {
	v, err := urcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (urcb *UpscaleRealtimeCreateBulk) Exec(ctx context.Context) error {
	_, err := urcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (urcb *UpscaleRealtimeCreateBulk) ExecX(ctx context.Context) {
	if err := urcb.Exec(ctx); err != nil {
		panic(err)
	}
}
