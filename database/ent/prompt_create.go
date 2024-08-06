// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"errors"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/sql"
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
	conflict []sql.ConflictOption
}

// SetText sets the "text" field.
func (pc *PromptCreate) SetText(s string) *PromptCreate {
	pc.mutation.SetText(s)
	return pc
}

// SetTranslatedText sets the "translated_text" field.
func (pc *PromptCreate) SetTranslatedText(s string) *PromptCreate {
	pc.mutation.SetTranslatedText(s)
	return pc
}

// SetNillableTranslatedText sets the "translated_text" field if the given value is not nil.
func (pc *PromptCreate) SetNillableTranslatedText(s *string) *PromptCreate {
	if s != nil {
		pc.SetTranslatedText(*s)
	}
	return pc
}

// SetRanTranslation sets the "ran_translation" field.
func (pc *PromptCreate) SetRanTranslation(b bool) *PromptCreate {
	pc.mutation.SetRanTranslation(b)
	return pc
}

// SetNillableRanTranslation sets the "ran_translation" field if the given value is not nil.
func (pc *PromptCreate) SetNillableRanTranslation(b *bool) *PromptCreate {
	if b != nil {
		pc.SetRanTranslation(*b)
	}
	return pc
}

// SetType sets the "type" field.
func (pc *PromptCreate) SetType(pr prompt.Type) *PromptCreate {
	pc.mutation.SetType(pr)
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
	return withHooks(ctx, pc.sqlSave, pc.mutation, pc.hooks)
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
	if _, ok := pc.mutation.RanTranslation(); !ok {
		v := prompt.DefaultRanTranslation
		pc.mutation.SetRanTranslation(v)
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
	if _, ok := pc.mutation.RanTranslation(); !ok {
		return &ValidationError{Name: "ran_translation", err: errors.New(`ent: missing required field "Prompt.ran_translation"`)}
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
		_spec = sqlgraph.NewCreateSpec(prompt.Table, sqlgraph.NewFieldSpec(prompt.FieldID, field.TypeUUID))
	)
	_spec.OnConflict = pc.conflict
	if id, ok := pc.mutation.ID(); ok {
		_node.ID = id
		_spec.ID.Value = &id
	}
	if value, ok := pc.mutation.Text(); ok {
		_spec.SetField(prompt.FieldText, field.TypeString, value)
		_node.Text = value
	}
	if value, ok := pc.mutation.TranslatedText(); ok {
		_spec.SetField(prompt.FieldTranslatedText, field.TypeString, value)
		_node.TranslatedText = &value
	}
	if value, ok := pc.mutation.RanTranslation(); ok {
		_spec.SetField(prompt.FieldRanTranslation, field.TypeBool, value)
		_node.RanTranslation = value
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
				IDSpec: sqlgraph.NewFieldSpec(generation.FieldID, field.TypeUUID),
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
				IDSpec: sqlgraph.NewFieldSpec(voiceover.FieldID, field.TypeUUID),
			},
		}
		for _, k := range nodes {
			edge.Target.Nodes = append(edge.Target.Nodes, k)
		}
		_spec.Edges = append(_spec.Edges, edge)
	}
	return _node, _spec
}

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Prompt.Create().
//		SetText(v).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.PromptUpsert) {
//			SetText(v+v).
//		}).
//		Exec(ctx)
func (pc *PromptCreate) OnConflict(opts ...sql.ConflictOption) *PromptUpsertOne {
	pc.conflict = opts
	return &PromptUpsertOne{
		create: pc,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Prompt.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (pc *PromptCreate) OnConflictColumns(columns ...string) *PromptUpsertOne {
	pc.conflict = append(pc.conflict, sql.ConflictColumns(columns...))
	return &PromptUpsertOne{
		create: pc,
	}
}

type (
	// PromptUpsertOne is the builder for "upsert"-ing
	//  one Prompt node.
	PromptUpsertOne struct {
		create *PromptCreate
	}

	// PromptUpsert is the "OnConflict" setter.
	PromptUpsert struct {
		*sql.UpdateSet
	}
)

// SetText sets the "text" field.
func (u *PromptUpsert) SetText(v string) *PromptUpsert {
	u.Set(prompt.FieldText, v)
	return u
}

// UpdateText sets the "text" field to the value that was provided on create.
func (u *PromptUpsert) UpdateText() *PromptUpsert {
	u.SetExcluded(prompt.FieldText)
	return u
}

// SetTranslatedText sets the "translated_text" field.
func (u *PromptUpsert) SetTranslatedText(v string) *PromptUpsert {
	u.Set(prompt.FieldTranslatedText, v)
	return u
}

// UpdateTranslatedText sets the "translated_text" field to the value that was provided on create.
func (u *PromptUpsert) UpdateTranslatedText() *PromptUpsert {
	u.SetExcluded(prompt.FieldTranslatedText)
	return u
}

// ClearTranslatedText clears the value of the "translated_text" field.
func (u *PromptUpsert) ClearTranslatedText() *PromptUpsert {
	u.SetNull(prompt.FieldTranslatedText)
	return u
}

// SetRanTranslation sets the "ran_translation" field.
func (u *PromptUpsert) SetRanTranslation(v bool) *PromptUpsert {
	u.Set(prompt.FieldRanTranslation, v)
	return u
}

// UpdateRanTranslation sets the "ran_translation" field to the value that was provided on create.
func (u *PromptUpsert) UpdateRanTranslation() *PromptUpsert {
	u.SetExcluded(prompt.FieldRanTranslation)
	return u
}

// SetType sets the "type" field.
func (u *PromptUpsert) SetType(v prompt.Type) *PromptUpsert {
	u.Set(prompt.FieldType, v)
	return u
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *PromptUpsert) UpdateType() *PromptUpsert {
	u.SetExcluded(prompt.FieldType)
	return u
}

// SetUpdatedAt sets the "updated_at" field.
func (u *PromptUpsert) SetUpdatedAt(v time.Time) *PromptUpsert {
	u.Set(prompt.FieldUpdatedAt, v)
	return u
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *PromptUpsert) UpdateUpdatedAt() *PromptUpsert {
	u.SetExcluded(prompt.FieldUpdatedAt)
	return u
}

// UpdateNewValues updates the mutable fields using the new values that were set on create except the ID field.
// Using this option is equivalent to using:
//
//	client.Prompt.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(prompt.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *PromptUpsertOne) UpdateNewValues() *PromptUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		if _, exists := u.create.mutation.ID(); exists {
			s.SetIgnore(prompt.FieldID)
		}
		if _, exists := u.create.mutation.CreatedAt(); exists {
			s.SetIgnore(prompt.FieldCreatedAt)
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Prompt.Create().
//	    OnConflict(sql.ResolveWithIgnore()).
//	    Exec(ctx)
func (u *PromptUpsertOne) Ignore() *PromptUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *PromptUpsertOne) DoNothing() *PromptUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the PromptCreate.OnConflict
// documentation for more info.
func (u *PromptUpsertOne) Update(set func(*PromptUpsert)) *PromptUpsertOne {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&PromptUpsert{UpdateSet: update})
	}))
	return u
}

// SetText sets the "text" field.
func (u *PromptUpsertOne) SetText(v string) *PromptUpsertOne {
	return u.Update(func(s *PromptUpsert) {
		s.SetText(v)
	})
}

// UpdateText sets the "text" field to the value that was provided on create.
func (u *PromptUpsertOne) UpdateText() *PromptUpsertOne {
	return u.Update(func(s *PromptUpsert) {
		s.UpdateText()
	})
}

// SetTranslatedText sets the "translated_text" field.
func (u *PromptUpsertOne) SetTranslatedText(v string) *PromptUpsertOne {
	return u.Update(func(s *PromptUpsert) {
		s.SetTranslatedText(v)
	})
}

// UpdateTranslatedText sets the "translated_text" field to the value that was provided on create.
func (u *PromptUpsertOne) UpdateTranslatedText() *PromptUpsertOne {
	return u.Update(func(s *PromptUpsert) {
		s.UpdateTranslatedText()
	})
}

// ClearTranslatedText clears the value of the "translated_text" field.
func (u *PromptUpsertOne) ClearTranslatedText() *PromptUpsertOne {
	return u.Update(func(s *PromptUpsert) {
		s.ClearTranslatedText()
	})
}

// SetRanTranslation sets the "ran_translation" field.
func (u *PromptUpsertOne) SetRanTranslation(v bool) *PromptUpsertOne {
	return u.Update(func(s *PromptUpsert) {
		s.SetRanTranslation(v)
	})
}

// UpdateRanTranslation sets the "ran_translation" field to the value that was provided on create.
func (u *PromptUpsertOne) UpdateRanTranslation() *PromptUpsertOne {
	return u.Update(func(s *PromptUpsert) {
		s.UpdateRanTranslation()
	})
}

// SetType sets the "type" field.
func (u *PromptUpsertOne) SetType(v prompt.Type) *PromptUpsertOne {
	return u.Update(func(s *PromptUpsert) {
		s.SetType(v)
	})
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *PromptUpsertOne) UpdateType() *PromptUpsertOne {
	return u.Update(func(s *PromptUpsert) {
		s.UpdateType()
	})
}

// SetUpdatedAt sets the "updated_at" field.
func (u *PromptUpsertOne) SetUpdatedAt(v time.Time) *PromptUpsertOne {
	return u.Update(func(s *PromptUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *PromptUpsertOne) UpdateUpdatedAt() *PromptUpsertOne {
	return u.Update(func(s *PromptUpsert) {
		s.UpdateUpdatedAt()
	})
}

// Exec executes the query.
func (u *PromptUpsertOne) Exec(ctx context.Context) error {
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for PromptCreate.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *PromptUpsertOne) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}

// Exec executes the UPSERT query and returns the inserted/updated ID.
func (u *PromptUpsertOne) ID(ctx context.Context) (id uuid.UUID, err error) {
	if u.create.driver.Dialect() == dialect.MySQL {
		// In case of "ON CONFLICT", there is no way to get back non-numeric ID
		// fields from the database since MySQL does not support the RETURNING clause.
		return id, errors.New("ent: PromptUpsertOne.ID is not supported by MySQL driver. Use PromptUpsertOne.Exec instead")
	}
	node, err := u.create.Save(ctx)
	if err != nil {
		return id, err
	}
	return node.ID, nil
}

// IDX is like ID, but panics if an error occurs.
func (u *PromptUpsertOne) IDX(ctx context.Context) uuid.UUID {
	id, err := u.ID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// PromptCreateBulk is the builder for creating many Prompt entities in bulk.
type PromptCreateBulk struct {
	config
	err      error
	builders []*PromptCreate
	conflict []sql.ConflictOption
}

// Save creates the Prompt entities in the database.
func (pcb *PromptCreateBulk) Save(ctx context.Context) ([]*Prompt, error) {
	if pcb.err != nil {
		return nil, pcb.err
	}
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
				var err error
				nodes[i], specs[i] = builder.createSpec()
				if i < len(mutators)-1 {
					_, err = mutators[i+1].Mutate(root, pcb.builders[i+1].mutation)
				} else {
					spec := &sqlgraph.BatchCreateSpec{Nodes: specs}
					spec.OnConflict = pcb.conflict
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

// OnConflict allows configuring the `ON CONFLICT` / `ON DUPLICATE KEY` clause
// of the `INSERT` statement. For example:
//
//	client.Prompt.CreateBulk(builders...).
//		OnConflict(
//			// Update the row with the new values
//			// the was proposed for insertion.
//			sql.ResolveWithNewValues(),
//		).
//		// Override some of the fields with custom
//		// update values.
//		Update(func(u *ent.PromptUpsert) {
//			SetText(v+v).
//		}).
//		Exec(ctx)
func (pcb *PromptCreateBulk) OnConflict(opts ...sql.ConflictOption) *PromptUpsertBulk {
	pcb.conflict = opts
	return &PromptUpsertBulk{
		create: pcb,
	}
}

// OnConflictColumns calls `OnConflict` and configures the columns
// as conflict target. Using this option is equivalent to using:
//
//	client.Prompt.Create().
//		OnConflict(sql.ConflictColumns(columns...)).
//		Exec(ctx)
func (pcb *PromptCreateBulk) OnConflictColumns(columns ...string) *PromptUpsertBulk {
	pcb.conflict = append(pcb.conflict, sql.ConflictColumns(columns...))
	return &PromptUpsertBulk{
		create: pcb,
	}
}

// PromptUpsertBulk is the builder for "upsert"-ing
// a bulk of Prompt nodes.
type PromptUpsertBulk struct {
	create *PromptCreateBulk
}

// UpdateNewValues updates the mutable fields using the new values that
// were set on create. Using this option is equivalent to using:
//
//	client.Prompt.Create().
//		OnConflict(
//			sql.ResolveWithNewValues(),
//			sql.ResolveWith(func(u *sql.UpdateSet) {
//				u.SetIgnore(prompt.FieldID)
//			}),
//		).
//		Exec(ctx)
func (u *PromptUpsertBulk) UpdateNewValues() *PromptUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithNewValues())
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(s *sql.UpdateSet) {
		for _, b := range u.create.builders {
			if _, exists := b.mutation.ID(); exists {
				s.SetIgnore(prompt.FieldID)
			}
			if _, exists := b.mutation.CreatedAt(); exists {
				s.SetIgnore(prompt.FieldCreatedAt)
			}
		}
	}))
	return u
}

// Ignore sets each column to itself in case of conflict.
// Using this option is equivalent to using:
//
//	client.Prompt.Create().
//		OnConflict(sql.ResolveWithIgnore()).
//		Exec(ctx)
func (u *PromptUpsertBulk) Ignore() *PromptUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWithIgnore())
	return u
}

// DoNothing configures the conflict_action to `DO NOTHING`.
// Supported only by SQLite and PostgreSQL.
func (u *PromptUpsertBulk) DoNothing() *PromptUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.DoNothing())
	return u
}

// Update allows overriding fields `UPDATE` values. See the PromptCreateBulk.OnConflict
// documentation for more info.
func (u *PromptUpsertBulk) Update(set func(*PromptUpsert)) *PromptUpsertBulk {
	u.create.conflict = append(u.create.conflict, sql.ResolveWith(func(update *sql.UpdateSet) {
		set(&PromptUpsert{UpdateSet: update})
	}))
	return u
}

// SetText sets the "text" field.
func (u *PromptUpsertBulk) SetText(v string) *PromptUpsertBulk {
	return u.Update(func(s *PromptUpsert) {
		s.SetText(v)
	})
}

// UpdateText sets the "text" field to the value that was provided on create.
func (u *PromptUpsertBulk) UpdateText() *PromptUpsertBulk {
	return u.Update(func(s *PromptUpsert) {
		s.UpdateText()
	})
}

// SetTranslatedText sets the "translated_text" field.
func (u *PromptUpsertBulk) SetTranslatedText(v string) *PromptUpsertBulk {
	return u.Update(func(s *PromptUpsert) {
		s.SetTranslatedText(v)
	})
}

// UpdateTranslatedText sets the "translated_text" field to the value that was provided on create.
func (u *PromptUpsertBulk) UpdateTranslatedText() *PromptUpsertBulk {
	return u.Update(func(s *PromptUpsert) {
		s.UpdateTranslatedText()
	})
}

// ClearTranslatedText clears the value of the "translated_text" field.
func (u *PromptUpsertBulk) ClearTranslatedText() *PromptUpsertBulk {
	return u.Update(func(s *PromptUpsert) {
		s.ClearTranslatedText()
	})
}

// SetRanTranslation sets the "ran_translation" field.
func (u *PromptUpsertBulk) SetRanTranslation(v bool) *PromptUpsertBulk {
	return u.Update(func(s *PromptUpsert) {
		s.SetRanTranslation(v)
	})
}

// UpdateRanTranslation sets the "ran_translation" field to the value that was provided on create.
func (u *PromptUpsertBulk) UpdateRanTranslation() *PromptUpsertBulk {
	return u.Update(func(s *PromptUpsert) {
		s.UpdateRanTranslation()
	})
}

// SetType sets the "type" field.
func (u *PromptUpsertBulk) SetType(v prompt.Type) *PromptUpsertBulk {
	return u.Update(func(s *PromptUpsert) {
		s.SetType(v)
	})
}

// UpdateType sets the "type" field to the value that was provided on create.
func (u *PromptUpsertBulk) UpdateType() *PromptUpsertBulk {
	return u.Update(func(s *PromptUpsert) {
		s.UpdateType()
	})
}

// SetUpdatedAt sets the "updated_at" field.
func (u *PromptUpsertBulk) SetUpdatedAt(v time.Time) *PromptUpsertBulk {
	return u.Update(func(s *PromptUpsert) {
		s.SetUpdatedAt(v)
	})
}

// UpdateUpdatedAt sets the "updated_at" field to the value that was provided on create.
func (u *PromptUpsertBulk) UpdateUpdatedAt() *PromptUpsertBulk {
	return u.Update(func(s *PromptUpsert) {
		s.UpdateUpdatedAt()
	})
}

// Exec executes the query.
func (u *PromptUpsertBulk) Exec(ctx context.Context) error {
	if u.create.err != nil {
		return u.create.err
	}
	for i, b := range u.create.builders {
		if len(b.conflict) != 0 {
			return fmt.Errorf("ent: OnConflict was set for builder %d. Set it on the PromptCreateBulk instead", i)
		}
	}
	if len(u.create.conflict) == 0 {
		return errors.New("ent: missing options for PromptCreateBulk.OnConflict")
	}
	return u.create.Exec(ctx)
}

// ExecX is like Exec, but panics if an error occurs.
func (u *PromptUpsertBulk) ExecX(ctx context.Context) {
	if err := u.create.Exec(ctx); err != nil {
		panic(err)
	}
}
