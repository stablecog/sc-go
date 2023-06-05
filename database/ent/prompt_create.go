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
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/database/ent/voiceover"
)

// PromptCreate is the builder for creating a Prompt entity.
type PromptCreate struct {
	config
	mutation *PromptMutation
	hooks    []Hook
}

// SetText sets the "text" field.
func (pc *PromptCreate) SetText(s string) *PromptCreate {
	pc.mutation.SetText(s)
	return pc
}

// SetType sets the "type" field.
func (pc *PromptCreate) SetType(pr prompt.Type) *PromptCreate {
	pc.mutation.SetType(pr)
	return pc
}

// SetNillableType sets the "type" field if the given value is not nil.
func (pc *PromptCreate) SetNillableType(pr *prompt.Type) *PromptCreate {
	if pr != nil {
		pc.SetType(*pr)
	}
	return pc
}

// SetCreatedAt sets the "created_at" field.
func (pc *PromptCreate) SetCreatedAt(t time.Time) *PromptCreate {
	pc.mutation.SetCreatedAt(t)
	return pc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (pc *PromptCreate) SetNillableCreatedAt(t *time.Time) *PromptCreate {
	if t != nil {
		pc.SetCreatedAt(*t)
	}
	return pc
}

// SetUpdatedAt sets the "updated_at" field.
func (pc *PromptCreate) SetUpdatedAt(t time.Time) *PromptCreate {
	pc.mutation.SetUpdatedAt(t)
	return pc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (pc *PromptCreate) SetNillableUpdatedAt(t *time.Time) *PromptCreate {
	if t != nil {
		pc.SetUpdatedAt(*t)
	}
	return pc
}

// SetID sets the "id" field.
func (pc *PromptCreate) SetID(u uuid.UUID) *PromptCreate {
	pc.mutation.SetID(u)
	return pc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (pc *PromptCreate) SetNillableID(u *uuid.UUID) *PromptCreate {
	if u != nil {
		pc.SetID(*u)
	}
	return pc
}

// AddGenerationIDs adds the "generations" edge to the Generation entity by IDs.
func (pc *PromptCreate) AddGenerationIDs(ids ...uuid.UUID) *PromptCreate {
	pc.mutation.AddGenerationIDs(ids...)
	return pc
}

// AddGenerations adds the "generations" edges to the Generation entity.
func (pc *PromptCreate) AddGenerations(g ...*Generation) *PromptCreate {
	ids := make([]uuid.UUID, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return pc.AddGenerationIDs(ids...)
}

// AddVoiceoverIDs adds the "voiceovers" edge to the Voiceover entity by IDs.
func (pc *PromptCreate) AddVoiceoverIDs(ids ...uuid.UUID) *PromptCreate {
	pc.mutation.AddVoiceoverIDs(ids...)
	return pc
}

// AddVoiceovers adds the "voiceovers" edges to the Voiceover entity.
func (pc *PromptCreate) AddVoiceovers(v ...*Voiceover) *PromptCreate {
	ids := make([]uuid.UUID, len(v))
	for i := range v {
		ids[i] = v[i].ID
	}
	return pc.AddVoiceoverIDs(ids...)
}

// Mutation returns the PromptMutation object of the builder.
func (pc *PromptCreate) Mutation() *PromptMutation {
	return pc.mutation
}

// Save creates the Prompt in the database.
func (pc *PromptCreate) Save(ctx context.Context) (*Prompt, error) {
	pc.defaults()
	return withHooks[*Prompt, PromptMutation](ctx, pc.sqlSave, pc.mutation, pc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (pc *PromptCreate) SaveX(ctx context.Context) *Prompt {
	v, err := pc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (pc *PromptCreate) Exec(ctx context.Context) error {
	_, err := pc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pc *PromptCreate) ExecX(ctx context.Context) {
	if err := pc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (pc *PromptCreate) defaults() {
	if _, ok := pc.mutation.GetType(); !ok {
		v := prompt.DefaultType
		pc.mutation.SetType(v)
	}
	if _, ok := pc.mutation.CreatedAt(); !ok {
		v := prompt.DefaultCreatedAt()
		pc.mutation.SetCreatedAt(v)
	}
	if _, ok := pc.mutation.UpdatedAt(); !ok {
		v := prompt.DefaultUpdatedAt()
		pc.mutation.SetUpdatedAt(v)
	}
	if _, ok := pc.mutation.ID(); !ok {
		v := prompt.DefaultID()
		pc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (pc *PromptCreate) check() error {
	if _, ok := pc.mutation.Text(); !ok {
		return &ValidationError{Name: "text", err: errors.New(`ent: missing required field "Prompt.text"`)}
	}
	if _, ok := pc.mutation.GetType(); !ok {
		return &ValidationError{Name: "type", err: errors.New(`ent: missing required field "Prompt.type"`)}
	}
	if v, ok := pc.mutation.GetType(); ok {
		if err := prompt.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "Prompt.type": %w`, err)}
		}
	}
	if _, ok := pc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "Prompt.created_at"`)}
	}
	if _, ok := pc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "Prompt.updated_at"`)}
	}
	return nil
}

func (pc *PromptCreate) sqlSave(ctx context.Context) (*Prompt, error) {
	if err := pc.check(); err != nil {
		return nil, err
	}
	_node, _spec := pc.createSpec()
	if err := sqlgraph.CreateNode(ctx, pc.driver, _spec); err != nil {
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
	pc.mutation.id = &_node.ID
	pc.mutation.done = true
	return _node, nil
}

func (pc *PromptCreate) createSpec() (*Prompt, *sqlgraph.CreateSpec) {
	var (
		_node = &Prompt{config: pc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: prompt.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: prompt.FieldID,
			},
		}
	)
	if id, ok := pc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := pc.mutation.Text(); ok {
		_spec.SetField(prompt.FieldText, field.TypeString, value)
		_node.Text = value
	}
	if value, ok := pc.mutation.GetType(); ok {
		_spec.SetField(prompt.FieldType, field.TypeEnum, value)
		_node.Type = value
	}
	if value, ok := pc.mutation.CreatedAt(); ok {
		_spec.SetField(prompt.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := pc.mutation.UpdatedAt(); ok {
		_spec.SetField(prompt.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if nodes := pc.mutation.GenerationsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   prompt.GenerationsTable,
			Columns: []string{prompt.GenerationsColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: generation.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := pc.mutation.VoiceoversIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.O2M,
			Inverse: false,
			Table:   prompt.VoiceoversTable,
			Columns: []string{prompt.VoiceoversColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: voiceover.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// PromptCreateBulk is the builder for creating many Prompt entities in bulk.
type PromptCreateBulk struct {
	config
	builders []*PromptCreate
}

// Save creates the Prompt entities in the database.
func (pcb *PromptCreateBulk) Save(ctx context.Context) ([]*Prompt, error) {
	specs := make([]*sqlgraph.CreateSpec, len(pcb.builders))
	nodes := make([]*Prompt, len(pcb.builders))
	mutators := make([]Mutator, len(pcb.builders))
	for i := range pcb.builders {
		func(i int, root context.Context) {
			builder := pcb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*PromptMutation)
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
					_, err = mutators[i+1].Mutate(root, pcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, pcb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, pcb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (pcb *PromptCreateBulk) SaveX(ctx context.Context) []*Prompt {
	v, err := pcb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (pcb *PromptCreateBulk) Exec(ctx context.Context) error {
	_, err := pcb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pcb *PromptCreateBulk) ExecX(ctx context.Context) {
	if err := pcb.Exec(ctx); err != nil {
		panic(err)
	}
}
