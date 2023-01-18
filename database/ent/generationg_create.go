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
	"github.com/stablecog/go-apps/database/ent/generationg"
	"github.com/stablecog/go-apps/database/ent/model"
	"github.com/stablecog/go-apps/database/ent/negativeprompt"
	"github.com/stablecog/go-apps/database/ent/prompt"
	"github.com/stablecog/go-apps/database/ent/scheduler"
	"github.com/stablecog/go-apps/database/ent/user"
	"github.com/stablecog/go-apps/database/enttypes"
)

// GenerationGCreate is the builder for creating a GenerationG entity.
type GenerationGCreate struct {
	config
	mutation *GenerationGMutation
	hooks    []Hook
}

// SetPromptID sets the "prompt_id" field.
func (gg *GenerationGCreate) SetPromptID(u uuid.UUID) *GenerationGCreate {
	gg.mutation.SetPromptID(u)
	return gg
}

// SetNegativePromptID sets the "negative_prompt_id" field.
func (gg *GenerationGCreate) SetNegativePromptID(u uuid.UUID) *GenerationGCreate {
	gg.mutation.SetNegativePromptID(u)
	return gg
}

// SetModelID sets the "model_id" field.
func (gg *GenerationGCreate) SetModelID(u uuid.UUID) *GenerationGCreate {
	gg.mutation.SetModelID(u)
	return gg
}

// SetImageID sets the "image_id" field.
func (gg *GenerationGCreate) SetImageID(s string) *GenerationGCreate {
	gg.mutation.SetImageID(s)
	return gg
}

// SetWidth sets the "width" field.
func (gg *GenerationGCreate) SetWidth(i int) *GenerationGCreate {
	gg.mutation.SetWidth(i)
	return gg
}

// SetHeight sets the "height" field.
func (gg *GenerationGCreate) SetHeight(i int) *GenerationGCreate {
	gg.mutation.SetHeight(i)
	return gg
}

// SetSeed sets the "seed" field.
func (gg *GenerationGCreate) SetSeed(ei enttypes.BigInt) *GenerationGCreate {
	gg.mutation.SetSeed(ei)
	return gg
}

// SetNillableSeed sets the "seed" field if the given value is not nil.
func (gg *GenerationGCreate) SetNillableSeed(ei *enttypes.BigInt) *GenerationGCreate {
	if ei != nil {
		gg.SetSeed(*ei)
	}
	return gg
}

// SetNumInferenceSteps sets the "num_inference_steps" field.
func (gg *GenerationGCreate) SetNumInferenceSteps(i int) *GenerationGCreate {
	gg.mutation.SetNumInferenceSteps(i)
	return gg
}

// SetGuidanceScale sets the "guidance_scale" field.
func (gg *GenerationGCreate) SetGuidanceScale(f float64) *GenerationGCreate {
	gg.mutation.SetGuidanceScale(f)
	return gg
}

// SetHidden sets the "hidden" field.
func (gg *GenerationGCreate) SetHidden(b bool) *GenerationGCreate {
	gg.mutation.SetHidden(b)
	return gg
}

// SetNillableHidden sets the "hidden" field if the given value is not nil.
func (gg *GenerationGCreate) SetNillableHidden(b *bool) *GenerationGCreate {
	if b != nil {
		gg.SetHidden(*b)
	}
	return gg
}

// SetSchedulerID sets the "scheduler_id" field.
func (gg *GenerationGCreate) SetSchedulerID(u uuid.UUID) *GenerationGCreate {
	gg.mutation.SetSchedulerID(u)
	return gg
}

// SetUserID sets the "user_id" field.
func (gg *GenerationGCreate) SetUserID(u uuid.UUID) *GenerationGCreate {
	gg.mutation.SetUserID(u)
	return gg
}

// SetUserTier sets the "user_tier" field.
func (gg *GenerationGCreate) SetUserTier(gt generationg.UserTier) *GenerationGCreate {
	gg.mutation.SetUserTier(gt)
	return gg
}

// SetNillableUserTier sets the "user_tier" field if the given value is not nil.
func (gg *GenerationGCreate) SetNillableUserTier(gt *generationg.UserTier) *GenerationGCreate {
	if gt != nil {
		gg.SetUserTier(*gt)
	}
	return gg
}

// SetCreatedAt sets the "created_at" field.
func (gg *GenerationGCreate) SetCreatedAt(t time.Time) *GenerationGCreate {
	gg.mutation.SetCreatedAt(t)
	return gg
}

// SetNillableCreatedAt sets the "created_at" field if the given value is not nil.
func (gg *GenerationGCreate) SetNillableCreatedAt(t *time.Time) *GenerationGCreate {
	if t != nil {
		gg.SetCreatedAt(*t)
	}
	return gg
}

// SetUpdatedAt sets the "updated_at" field.
func (gg *GenerationGCreate) SetUpdatedAt(t time.Time) *GenerationGCreate {
	gg.mutation.SetUpdatedAt(t)
	return gg
}

// SetNillableUpdatedAt sets the "updated_at" field if the given value is not nil.
func (gg *GenerationGCreate) SetNillableUpdatedAt(t *time.Time) *GenerationGCreate {
	if t != nil {
		gg.SetUpdatedAt(*t)
	}
	return gg
}

// SetID sets the "id" field.
func (gg *GenerationGCreate) SetID(u uuid.UUID) *GenerationGCreate {
	gg.mutation.SetID(u)
	return gg
}

// SetNillableID sets the "id" field if the given value is not nil.
func (gg *GenerationGCreate) SetNillableID(u *uuid.UUID) *GenerationGCreate {
	if u != nil {
		gg.SetID(*u)
	}
	return gg
}

// SetUser sets the "user" edge to the User entity.
func (gg *GenerationGCreate) SetUser(u *User) *GenerationGCreate {
	return gg.SetUserID(u.ID)
}

// SetModel sets the "model" edge to the Model entity.
func (gg *GenerationGCreate) SetModel(m *Model) *GenerationGCreate {
	return gg.SetModelID(m.ID)
}

// SetPrompt sets the "prompt" edge to the Prompt entity.
func (gg *GenerationGCreate) SetPrompt(p *Prompt) *GenerationGCreate {
	return gg.SetPromptID(p.ID)
}

// SetNegativePrompt sets the "negative_prompt" edge to the NegativePrompt entity.
func (gg *GenerationGCreate) SetNegativePrompt(n *NegativePrompt) *GenerationGCreate {
	return gg.SetNegativePromptID(n.ID)
}

// SetScheduler sets the "scheduler" edge to the Scheduler entity.
func (gg *GenerationGCreate) SetScheduler(s *Scheduler) *GenerationGCreate {
	return gg.SetSchedulerID(s.ID)
}

// Mutation returns the GenerationGMutation object of the builder.
func (gg *GenerationGCreate) Mutation() *GenerationGMutation {
	return gg.mutation
}

// Save creates the GenerationG in the database.
func (gg *GenerationGCreate) Save(ctx context.Context) (*GenerationG, error) {
	gg.defaults()
	return withHooks[*GenerationG, GenerationGMutation](ctx, gg.sqlSave, gg.mutation, gg.hooks)
}

// SaveX calls Save and panics if Save returns an error.
func (gg *GenerationGCreate) SaveX(ctx context.Context) *GenerationG {
	v, err := gg.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (gg *GenerationGCreate) Exec(ctx context.Context) error {
	_, err := gg.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (gg *GenerationGCreate) ExecX(ctx context.Context) {
	if err := gg.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (gg *GenerationGCreate) defaults() {
	if _, ok := gg.mutation.Hidden(); !ok {
		v := generationg.DefaultHidden
		gg.mutation.SetHidden(v)
	}
	if _, ok := gg.mutation.UserTier(); !ok {
		v := generationg.DefaultUserTier
		gg.mutation.SetUserTier(v)
	}
	if _, ok := gg.mutation.CreatedAt(); !ok {
		v := generationg.DefaultCreatedAt()
		gg.mutation.SetCreatedAt(v)
	}
	if _, ok := gg.mutation.UpdatedAt(); !ok {
		v := generationg.DefaultUpdatedAt()
		gg.mutation.SetUpdatedAt(v)
	}
	if _, ok := gg.mutation.ID(); !ok {
		v := generationg.DefaultID()
		gg.mutation.SetID(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (gg *GenerationGCreate) check() error {
	if _, ok := gg.mutation.PromptID(); !ok {
		return &ValidationError{Name: "prompt_id", err: errors.New(`ent: missing required field "GenerationG.prompt_id"`)}
	}
	if _, ok := gg.mutation.NegativePromptID(); !ok {
		return &ValidationError{Name: "negative_prompt_id", err: errors.New(`ent: missing required field "GenerationG.negative_prompt_id"`)}
	}
	if _, ok := gg.mutation.ModelID(); !ok {
		return &ValidationError{Name: "model_id", err: errors.New(`ent: missing required field "GenerationG.model_id"`)}
	}
	if _, ok := gg.mutation.ImageID(); !ok {
		return &ValidationError{Name: "image_id", err: errors.New(`ent: missing required field "GenerationG.image_id"`)}
	}
	if _, ok := gg.mutation.Width(); !ok {
		return &ValidationError{Name: "width", err: errors.New(`ent: missing required field "GenerationG.width"`)}
	}
	if _, ok := gg.mutation.Height(); !ok {
		return &ValidationError{Name: "height", err: errors.New(`ent: missing required field "GenerationG.height"`)}
	}
	if _, ok := gg.mutation.NumInferenceSteps(); !ok {
		return &ValidationError{Name: "num_inference_steps", err: errors.New(`ent: missing required field "GenerationG.num_inference_steps"`)}
	}
	if _, ok := gg.mutation.GuidanceScale(); !ok {
		return &ValidationError{Name: "guidance_scale", err: errors.New(`ent: missing required field "GenerationG.guidance_scale"`)}
	}
	if _, ok := gg.mutation.Hidden(); !ok {
		return &ValidationError{Name: "hidden", err: errors.New(`ent: missing required field "GenerationG.hidden"`)}
	}
	if _, ok := gg.mutation.SchedulerID(); !ok {
		return &ValidationError{Name: "scheduler_id", err: errors.New(`ent: missing required field "GenerationG.scheduler_id"`)}
	}
	if _, ok := gg.mutation.UserID(); !ok {
		return &ValidationError{Name: "user_id", err: errors.New(`ent: missing required field "GenerationG.user_id"`)}
	}
	if _, ok := gg.mutation.UserTier(); !ok {
		return &ValidationError{Name: "user_tier", err: errors.New(`ent: missing required field "GenerationG.user_tier"`)}
	}
	if v, ok := gg.mutation.UserTier(); ok {
		if err := generationg.UserTierValidator(v); err != nil {
			return &ValidationError{Name: "user_tier", err: fmt.Errorf(`ent: validator failed for field "GenerationG.user_tier": %w`, err)}
		}
	}
	if _, ok := gg.mutation.CreatedAt(); !ok {
		return &ValidationError{Name: "created_at", err: errors.New(`ent: missing required field "GenerationG.created_at"`)}
	}
	if _, ok := gg.mutation.UpdatedAt(); !ok {
		return &ValidationError{Name: "updated_at", err: errors.New(`ent: missing required field "GenerationG.updated_at"`)}
	}
	if _, ok := gg.mutation.UserID(); !ok {
		return &ValidationError{Name: "user", err: errors.New(`ent: missing required edge "GenerationG.user"`)}
	}
	if _, ok := gg.mutation.ModelID(); !ok {
		return &ValidationError{Name: "model", err: errors.New(`ent: missing required edge "GenerationG.model"`)}
	}
	if _, ok := gg.mutation.PromptID(); !ok {
		return &ValidationError{Name: "prompt", err: errors.New(`ent: missing required edge "GenerationG.prompt"`)}
	}
	if _, ok := gg.mutation.NegativePromptID(); !ok {
		return &ValidationError{Name: "negative_prompt", err: errors.New(`ent: missing required edge "GenerationG.negative_prompt"`)}
	}
	if _, ok := gg.mutation.SchedulerID(); !ok {
		return &ValidationError{Name: "scheduler", err: errors.New(`ent: missing required edge "GenerationG.scheduler"`)}
	}
	return nil
}

func (gg *GenerationGCreate) sqlSave(ctx context.Context) (*GenerationG, error) {
	if err := gg.check(); err != nil {
		return nil, err
	}
	_node, _spec := gg.createSpec()
	if err := sqlgraph.CreateNode(ctx, gg.driver, _spec); err != nil {
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
	gg.mutation.id = &_node.ID
	gg.mutation.done = true
	return _node, nil
}

func (gg *GenerationGCreate) createSpec() (*GenerationG, *sqlgraph.CreateSpec) {
	var (
		_node = &GenerationG{config: gg.config}
		_spec = &sqlgraph.CreateSpec{
			Table: generationg.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: generationg.FieldID,
			},
		}
	)
	if id, ok := gg.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := gg.mutation.ImageID(); ok {
		_spec.SetField(generationg.FieldImageID, field.TypeString, value)
		_node.ImageID = value
	}
	if value, ok := gg.mutation.Width(); ok {
		_spec.SetField(generationg.FieldWidth, field.TypeInt, value)
		_node.Width = value
	}
	if value, ok := gg.mutation.Height(); ok {
		_spec.SetField(generationg.FieldHeight, field.TypeInt, value)
		_node.Height = value
	}
	if value, ok := gg.mutation.Seed(); ok {
		_spec.SetField(generationg.FieldSeed, field.TypeInt, value)
		_node.Seed = &value
	}
	if value, ok := gg.mutation.NumInferenceSteps(); ok {
		_spec.SetField(generationg.FieldNumInferenceSteps, field.TypeInt, value)
		_node.NumInferenceSteps = &value
	}
	if value, ok := gg.mutation.GuidanceScale(); ok {
		_spec.SetField(generationg.FieldGuidanceScale, field.TypeFloat64, value)
		_node.GuidanceScale = value
	}
	if value, ok := gg.mutation.Hidden(); ok {
		_spec.SetField(generationg.FieldHidden, field.TypeBool, value)
		_node.Hidden = value
	}
	if value, ok := gg.mutation.UserTier(); ok {
		_spec.SetField(generationg.FieldUserTier, field.TypeEnum, value)
		_node.UserTier = value
	}
	if value, ok := gg.mutation.CreatedAt(); ok {
		_spec.SetField(generationg.FieldCreatedAt, field.TypeTime, value)
		_node.CreatedAt = value
	}
	if value, ok := gg.mutation.UpdatedAt(); ok {
		_spec.SetField(generationg.FieldUpdatedAt, field.TypeTime, value)
		_node.UpdatedAt = value
	}
	if nodes := gg.mutation.UserIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   generationg.UserTable,
			Columns: []string{generationg.UserColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: user.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.UserID = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := gg.mutation.ModelIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   generationg.ModelTable,
			Columns: []string{generationg.ModelColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: model.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.ModelID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := gg.mutation.PromptIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   generationg.PromptTable,
			Columns: []string{generationg.PromptColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: prompt.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.PromptID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := gg.mutation.NegativePromptIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   generationg.NegativePromptTable,
			Columns: []string{generationg.NegativePromptColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: negativeprompt.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.NegativePromptID = &nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	if nodes := gg.mutation.SchedulerIDs(); len(nodes) > 0 {
		edge := &sqlgraph.EdgeSpec{
			Rel:     sqlgraph.M2O,
			Inverse: true,
			Table:   generationg.SchedulerTable,
			Columns: []string{generationg.SchedulerColumn},
			Bidi:    false,
			Target: &sqlgraph.EdgeTarget{
				IDSpec: &sqlgraph.FieldSpec{
					Type:   field.TypeUUID,
					Column: scheduler.FieldID,
				},
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_node.SchedulerID = nodes[0]
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// GenerationGCreateBulk is the builder for creating many GenerationG entities in bulk.
type GenerationGCreateBulk struct {
	config
	builders []*GenerationGCreate
}

// Save creates the GenerationG entities in the database.
func (ggb *GenerationGCreateBulk) Save(ctx context.Context) ([]*GenerationG, error) {
	specs := make([]*sqlgraph.CreateSpec, len(ggb.builders))
	nodes := make([]*GenerationG, len(ggb.builders))
	mutators := make([]Mutator, len(ggb.builders))
	for i := range ggb.builders {
		func(i int, root context.Context) {
			builder := ggb.builders[i]
			builder.defaults()
			var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
				mutation, ok := m.(*GenerationGMutation)
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
					_, err = mutators[i+1].Mutate(root, ggb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					// Invoke the actual operation on the latest mutation in the chain.
					if err = sqlgraph.BatchCreate(ctx, ggb.driver, spec); err != nil {
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
		if _, err := mutators[0].Mutate(ctx, ggb.builders[0].mutation); err != nil {
			return nil, err
		}
	}
	return nodes, nil
}

// SaveX is like Save, but panics if an error occurs.
func (ggb *GenerationGCreateBulk) SaveX(ctx context.Context) []*GenerationG {
	v, err := ggb.Save(ctx)
	if err != nil {
		panic(err)
	}
	return v
}

// Exec executes the query.
func (ggb *GenerationGCreateBulk) Exec(ctx context.Context) error {
	_, err := ggb.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (ggb *GenerationGCreateBulk) ExecX(ctx context.Context) {
	if err := ggb.Exec(ctx); err != nil {
		panic(err)
	}
}
