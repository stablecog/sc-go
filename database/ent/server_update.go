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
	"github.com/stablecog/go-apps/database/ent/predicate"
	"github.com/stablecog/go-apps/database/ent/server"
)

// ServerUpdate is the builder for updating Server entities.
type ServerUpdate struct {
	config
	hooks    []Hook
	mutation *ServerMutation
}

// Where appends a list predicates to the ServerUpdate builder.
func (su *ServerUpdate) Where(ps ...predicate.Server) *ServerUpdate {
	su.mutation.Where(ps...)
	return su
}

// SetURL sets the "url" field.
func (su *ServerUpdate) SetURL(s string) *ServerUpdate {
	su.mutation.SetURL(s)
	return su
}

// SetHealthy sets the "healthy" field.
func (su *ServerUpdate) SetHealthy(b bool) *ServerUpdate {
	su.mutation.SetHealthy(b)
	return su
}

// SetNillableHealthy sets the "healthy" field if the given value is not nil.
func (su *ServerUpdate) SetNillableHealthy(b *bool) *ServerUpdate {
	if b != nil {
		su.SetHealthy(*b)
	}
	return su
}

// SetEnabled sets the "enabled" field.
func (su *ServerUpdate) SetEnabled(b bool) *ServerUpdate {
	su.mutation.SetEnabled(b)
	return su
}

// SetNillableEnabled sets the "enabled" field if the given value is not nil.
func (su *ServerUpdate) SetNillableEnabled(b *bool) *ServerUpdate {
	if b != nil {
		su.SetEnabled(*b)
	}
	return su
}

// SetFeatures sets the "features" field.
func (su *ServerUpdate) SetFeatures(s struct {
	Name   string   "json:\"name\""
	Values []string "json:\"values,omitempty\""
}) *ServerUpdate {
	su.mutation.SetFeatures(s)
	return su
}

// SetUpdatedAt sets the "updated_at" field.
func (su *ServerUpdate) SetUpdatedAt(t time.Time) *ServerUpdate {
	su.mutation.SetUpdatedAt(t)
	return su
}

// SetUserTier sets the "user_tier" field.
func (su *ServerUpdate) SetUserTier(st server.UserTier) *ServerUpdate {
	su.mutation.SetUserTier(st)
	return su
}

// SetNillableUserTier sets the "user_tier" field if the given value is not nil.
func (su *ServerUpdate) SetNillableUserTier(st *server.UserTier) *ServerUpdate {
	if st != nil {
		su.SetUserTier(*st)
	}
	return su
}

// Mutation returns the ServerMutation object of the builder.
func (su *ServerUpdate) Mutation() *ServerMutation {
	return su.mutation
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (su *ServerUpdate) Save(ctx context.Context) (int, error) {
	su.defaults()
	return withHooks[int, ServerMutation](ctx, su.sqlSave, su.mutation, su.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (su *ServerUpdate) SaveX(ctx context.Context) int {
	affected, err := su.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (su *ServerUpdate) Exec(ctx context.Context) error {
	_, err := su.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (su *ServerUpdate) ExecX(ctx context.Context) {
	if err := su.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (su *ServerUpdate) defaults() {
	if _, ok := su.mutation.UpdatedAt(); !ok {
		v := server.UpdateDefaultUpdatedAt()
		su.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (su *ServerUpdate) check() error {
	if v, ok := su.mutation.UserTier(); ok {
		if err := server.UserTierValidator(v); err != nil {
			return &ValidationError{Name: "user_tier", err: fmt.Errorf(`ent: validator failed for field "Server.user_tier": %w`, err)}
		}
	}
	return nil
}

func (su *ServerUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := su.check(); err != nil {
		return n, err
	}
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   server.Table,
			Columns: server.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: server.FieldID,
			},
		},
	}
	if ps := su.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := su.mutation.URL(); ok {
		_spec.SetField(server.FieldURL, field.TypeString, value)
	}
	if value, ok := su.mutation.Healthy(); ok {
		_spec.SetField(server.FieldHealthy, field.TypeBool, value)
	}
	if value, ok := su.mutation.Enabled(); ok {
		_spec.SetField(server.FieldEnabled, field.TypeBool, value)
	}
	if value, ok := su.mutation.Features(); ok {
		_spec.SetField(server.FieldFeatures, field.TypeJSON, value)
	}
	if value, ok := su.mutation.UpdatedAt(); ok {
		_spec.SetField(server.FieldUpdatedAt, field.TypeTime, value)
	}
	if value, ok := su.mutation.UserTier(); ok {
		_spec.SetField(server.FieldUserTier, field.TypeEnum, value)
	}
	if n, err = sqlgraph.UpdateNodes(ctx, su.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{server.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	su.mutation.done = true
	return n, nil
}

// ServerUpdateOne is the builder for updating a single Server entity.
type ServerUpdateOne struct {
	config
	fields   []string
	hooks    []Hook
	mutation *ServerMutation
}

// SetURL sets the "url" field.
func (suo *ServerUpdateOne) SetURL(s string) *ServerUpdateOne {
	suo.mutation.SetURL(s)
	return suo
}

// SetHealthy sets the "healthy" field.
func (suo *ServerUpdateOne) SetHealthy(b bool) *ServerUpdateOne {
	suo.mutation.SetHealthy(b)
	return suo
}

// SetNillableHealthy sets the "healthy" field if the given value is not nil.
func (suo *ServerUpdateOne) SetNillableHealthy(b *bool) *ServerUpdateOne {
	if b != nil {
		suo.SetHealthy(*b)
	}
	return suo
}

// SetEnabled sets the "enabled" field.
func (suo *ServerUpdateOne) SetEnabled(b bool) *ServerUpdateOne {
	suo.mutation.SetEnabled(b)
	return suo
}

// SetNillableEnabled sets the "enabled" field if the given value is not nil.
func (suo *ServerUpdateOne) SetNillableEnabled(b *bool) *ServerUpdateOne {
	if b != nil {
		suo.SetEnabled(*b)
	}
	return suo
}

// SetFeatures sets the "features" field.
func (suo *ServerUpdateOne) SetFeatures(s struct {
	Name   string   "json:\"name\""
	Values []string "json:\"values,omitempty\""
}) *ServerUpdateOne {
	suo.mutation.SetFeatures(s)
	return suo
}

// SetUpdatedAt sets the "updated_at" field.
func (suo *ServerUpdateOne) SetUpdatedAt(t time.Time) *ServerUpdateOne {
	suo.mutation.SetUpdatedAt(t)
	return suo
}

// SetUserTier sets the "user_tier" field.
func (suo *ServerUpdateOne) SetUserTier(st server.UserTier) *ServerUpdateOne {
	suo.mutation.SetUserTier(st)
	return suo
}

// SetNillableUserTier sets the "user_tier" field if the given value is not nil.
func (suo *ServerUpdateOne) SetNillableUserTier(st *server.UserTier) *ServerUpdateOne {
	if st != nil {
		suo.SetUserTier(*st)
	}
	return suo
}

// Mutation returns the ServerMutation object of the builder.
func (suo *ServerUpdateOne) Mutation() *ServerMutation {
	return suo.mutation
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (suo *ServerUpdateOne) Select(field string, fields ...string) *ServerUpdateOne {
	suo.fields = append([]string{field}, fields...)
	return suo
}

// Save executes the query and returns the updated Server entity.
func (suo *ServerUpdateOne) Save(ctx context.Context) (*Server, error) {
	suo.defaults()
	return withHooks[*Server, ServerMutation](ctx, suo.sqlSave, suo.mutation, suo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (suo *ServerUpdateOne) SaveX(ctx context.Context) *Server {
	node, err := suo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (suo *ServerUpdateOne) Exec(ctx context.Context) error {
	_, err := suo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (suo *ServerUpdateOne) ExecX(ctx context.Context) {
	if err := suo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (suo *ServerUpdateOne) defaults() {
	if _, ok := suo.mutation.UpdatedAt(); !ok {
		v := server.UpdateDefaultUpdatedAt()
		suo.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (suo *ServerUpdateOne) check() error {
	if v, ok := suo.mutation.UserTier(); ok {
		if err := server.UserTierValidator(v); err != nil {
			return &ValidationError{Name: "user_tier", err: fmt.Errorf(`ent: validator failed for field "Server.user_tier": %w`, err)}
		}
	}
	return nil
}

func (suo *ServerUpdateOne) sqlSave(ctx context.Context) (_node *Server, err error) {
	if err := suo.check(); err != nil {
		return _node, err
	}
	_spec := &sqlgraph.UpdateSpec{
		Node: &sqlgraph.NodeSpec{
			Table:   server.Table,
			Columns: server.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: server.FieldID,
			},
		},
	}
	id, ok := suo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Server.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := suo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, server.FieldID)
		for _, f := range fields {
			if !server.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != server.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := suo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := suo.mutation.URL(); ok {
		_spec.SetField(server.FieldURL, field.TypeString, value)
	}
	if value, ok := suo.mutation.Healthy(); ok {
		_spec.SetField(server.FieldHealthy, field.TypeBool, value)
	}
	if value, ok := suo.mutation.Enabled(); ok {
		_spec.SetField(server.FieldEnabled, field.TypeBool, value)
	}
	if value, ok := suo.mutation.Features(); ok {
		_spec.SetField(server.FieldFeatures, field.TypeJSON, value)
	}
	if value, ok := suo.mutation.UpdatedAt(); ok {
		_spec.SetField(server.FieldUpdatedAt, field.TypeTime, value)
	}
	if value, ok := suo.mutation.UserTier(); ok {
		_spec.SetField(server.FieldUserTier, field.TypeEnum, value)
	}
	_node = &Server{config: suo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, suo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{server.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	suo.mutation.done = true
	return _node, nil
}
