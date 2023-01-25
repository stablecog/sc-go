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
	"github.com/stablecog/go-apps/database/ent/generation"
	"github.com/stablecog/go-apps/database/ent/generationoutput"
)

// GenerationOutputCreate is the builder for creating a GenerationOutput entity.
type GenerationOutputCreate struct {
	config
	mutation *GenerationOutputMutation
	hooks    []Hook
}

// SetImageURL sets the "image_url" field.
func (goc *GenerationOutputCreate) SetImageURL(s string) *GenerationOutputCreate {
	goc.mutation.SetImageURL(s)
	return goc
}

// SetUpscaledImageURL sets the "upscaled_image_url" field.
func (goc *GenerationOutputCreate) SetUpscaledImageURL(s string) *GenerationOutputCreate {
	goc.mutation.SetUpscaledImageURL(s)
	return goc
}

// SetNillableUpscaledImageURL sets the "upscaled_image_url" field if the given value is not nil.
func (goc *GenerationOutputCreate) SetNillableUpscaledImageURL(s *string) *GenerationOutputCreate {
	if s != nil {
		goc.SetUpscaledImageURL(*s)
	}
	return goc
}

// SetGenerationID sets the "generation_id" field.
func (goc *GenerationOutputCreate) SetGenerationID(u uuid.UUID) *GenerationOutputCreate {
	goc.mutation.SetGenerationID(u)
	return goc
}

// SetCreatedAt sets the "created_at" field.
func (goc *GenerationOutputCreate) SetCreatedAt(t time.Time) *GenerationOutputCreate {
	goc.mutation.SetCreatedAt(t)
	return goc
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (goc *GenerationOutputCreate) SetNillableCreatedAt(t *time.Time) *GenerationOutputCreate {
	if t != nil {
		goc.SetCreatedAt(*t)
	}
	return goc
}

// SetUpdatedAt sets the "updated_at" field.
func (goc *GenerationOutputCreate) SetUpdatedAt(t time.Time) *GenerationOutputCreate {
	goc.mutation.SetUpdatedAt(t)
	return goc
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (goc *GenerationOutputCreate) SetNillableUpdatedAt(t *time.Time) *GenerationOutputCreate {
	if t != nil {
		goc.SetUpdatedAt(*t)
	}
	return goc
}

// SetID sets the "id" field.
func (goc *GenerationOutputCreate) SetID(u uuid.UUID) *GenerationOutputCreate {
	goc.mutation.SetID(u)
	return goc
}

// SetNillableID sets the "id" field if the given value is not nil.
func (goc *GenerationOutputCreate) SetNillableID(u *uuid.UUID) *GenerationOutputCreate {
	if u != nil {
		goc.SetID(*u)
	}
	return goc
}

// SetGenerationsID sets the "generations" edge to the Generation entity by ID.
func (goc *GenerationOutputCreate) SetGenerationsID(id uuid.UUID) *GenerationOutputCreate {
	goc.mutation.SetGenerationsID(id)
	return goc
}

// SetGenerations sets the "generations" edge to the Generation entity.
func (goc *GenerationOutputCreate) SetGenerations(g *Generation) *GenerationOutputCreate {
	return goc.SetGenerationsID(g.ID)
}

// Mutation returns the GenerationOutputMutation object of the builder.
func (goc *GenerationOutputCreate) Mutation() *GenerationOutputMutation {
	return goc.mutation
}

// Save creates the GenerationOutput in the database.
func (goc *GenerationOutputCreate) Save(ctx context.Context) (*GenerationOutput, error) {
	goc.defaults()
	return withHooks[*GenerationOutput, GenerationOutputMutation](ctx, goc.sqlSave, goc.mutation, goc.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (goc *GenerationOutputCreate) SaveX(ctx context.Context) *GenerationOutput {
	v, err := goc.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (goc *GenerationOutputCreate) Exec(ctx context.Context) error {
	_, err := goc.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (goc *GenerationOutputCreate) ExecX(ctx context.Context) {
	if err := goc.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (goc *GenerationOutputCreate) defaults() {
	if _, ok := goc.mutation.CreatedAt(); !ok {
		v := generationoutput.DefaultCreatedAt()
		goc.mutation.SetCreatedAt(v)
	}
	if _, ok := goc.mutation.UpdatedAt(); !ok {
		v := generationoutput.DefaultUpdatedAt()
		goc.mutation.SetUpdatedAt(v)
	}
	if _, ok := goc.mutation.ID(); !ok {
		v := generationoutput.DefaultID()
		goc.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (goc *GenerationOutputCreate) check() error {
	if _, ok := goc.mutation.ImageURL(); !ok {
		return &ValidationError{Name: "image_url", err: errors.New(`ent: missing required field "GenerationOutput.image_url"`)}
	}
	if _, ok := goc.mutation.GenerationID(); !ok {
		return &ValidationError{Name: "generation_id", err: errors.New(`ent: missing required field "GenerationOutput.generation_id"`)}
	}
	if _, ok := goc.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "GenerationOutput.created_at"`)}
	}
	if _, ok := goc.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "GenerationOutput.updated_at"`)}
	}
	if _, ok := goc.mutation.GenerationsID(); !ok {
		return &ValidationError{Name: "generations", err: errors.New(`ent: missing required edge "GenerationOutput.generations"`)}
	}
	return nil
}

func (goc *GenerationOutputCreate) sqlSave(ctx context.Context) (*GenerationOutput, error) {
	if err := goc.check(); err != nil {
		return nil, err
	}
	_node, _spec := goc.createSpec()
	if err := sqlgraph.CreateNode(ctx, goc.driver, _spec); err != nil {
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
	goc.mutation.id = &_node.ID
	goc.mutation.done = true
	return _node, nil
}

func (goc *GenerationOutputCreate) createSpec() (*GenerationOutput, *sqlgraph.CreateSpec) {
	var (
		_node = &GenerationOutput{config: goc.config}
		_spec = &sqlgraph.CreateSpec{
			Table: generationoutput.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: generationoutput.FieldID,
			},
		}
	)
	if id, ok := goc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := goc.mutation.ImageURL(); ok {
		_spec.SetField(generationoutput.FieldImageURL, field.TypeString, value)
		_node.ImageURL = value
	}
	if value, ok := goc.mutation.UpscaledImageURL(); ok {
		_spec.SetField(generationoutput.FieldUpscaledImageURL, field.TypeString, value)
		_node.UpscaledImageURL = &value
	}
	if value, ok := goc.mutation.CreatedAt(); ok {
		_spec.SetField(generationoutput.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := goc.mutation.UpdatedAt(); ok {
		_spec.SetField(generationoutput.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if nodes := goc.mutation.GenerationsIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   generationoutput.GenerationsTable,
			Columns: []string{generationoutput.GenerationsColumn},
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
		_node.GenerationID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// GenerationOutputCreateBulk is the builder for creating many GenerationOutput entities in bulk.
type GenerationOutputCreateBulk struct {
	config
	builders []*GenerationOutputCreate
}

// Save creates the GenerationOutput entities in the database.
func (gocb *GenerationOutputCreateBulk) Save(ctx context.Context) ([]*GenerationOutput, error) {
	specs := make([]*sqlgraph.CreateSpec, len(gocb.builders))
	nodes := make([]*GenerationOutput, len(gocb.builders))
	mutators := make([]Mutator, len(gocb.builders))
	for i := range gocb.builders {
		func(i int, root context.Context) {
			builder := gocb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*GenerationOutputMutation)
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
					_, err = mutators[i+1].Mutate(root, gocb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, gocb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, gocb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (gocb *GenerationOutputCreateBulk) SaveX(ctx context.Context) []*GenerationOutput {
	v, err := gocb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (gocb *GenerationOutputCreateBulk) Exec(ctx context.Context) error {
	_, err := gocb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (gocb *GenerationOutputCreateBulk) ExecX(ctx context.Context) {
	if err := gocb.Exec(ctx); err != nil {
		panic(err)
	}
}
