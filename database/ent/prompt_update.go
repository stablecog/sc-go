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
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/generation"
	"github.com/stablecog/sc-go/database/ent/predicate"
	"github.com/stablecog/sc-go/database/ent/prompt"
	"github.com/stablecog/sc-go/database/ent/voiceover"
)

// PromptUpdate is the builder for updating Prompt entities.
type PromptUpdate struct {
	config
	hooks     []Hook
	mutation  *PromptMutation
	modifiers []func(*sql.UpdateBuilder)
}

// Where appends a list predicates to the PromptUpdate builder.
func (pu *PromptUpdate) Where(ps ...predicate.Prompt) *PromptUpdate {
	pu.mutation.Where(ps...)
	return pu
}

// SetText sets the "text" field.
func (pu *PromptUpdate) SetText(s string) *PromptUpdate {
	pu.mutation.SetText(s)
	return pu
}

// SetNillableText sets the "text" field if the given value is not nil.
func (pu *PromptUpdate) SetNillableText(s *string) *PromptUpdate {
	if s != nil {
		pu.SetText(*s)
	}
	return pu
}

// SetTranslatedText sets the "translated_text" field.
func (pu *PromptUpdate) SetTranslatedText(s string) *PromptUpdate {
	pu.mutation.SetTranslatedText(s)
	return pu
}

// SetNillableTranslatedText sets the "translated_text" field if the given value is not nil.
func (pu *PromptUpdate) SetNillableTranslatedText(s *string) *PromptUpdate {
	if s != nil {
		pu.SetTranslatedText(*s)
	}
	return pu
}

// ClearTranslatedText clears the value of the "translated_text" field.
func (pu *PromptUpdate) ClearTranslatedText() *PromptUpdate {
	pu.mutation.ClearTranslatedText()
	return pu
}

// SetRanTranslation sets the "ran_translation" field.
func (pu *PromptUpdate) SetRanTranslation(b bool) *PromptUpdate {
	pu.mutation.SetRanTranslation(b)
	return pu
}

// SetNillableRanTranslation sets the "ran_translation" field if the given value is not nil.
func (pu *PromptUpdate) SetNillableRanTranslation(b *bool) *PromptUpdate {
	if b != nil {
		pu.SetRanTranslation(*b)
	}
	return pu
}

// SetType sets the "type" field.
func (pu *PromptUpdate) SetType(pr prompt.Type) *PromptUpdate {
	pu.mutation.SetType(pr)
	return pu
}

// SetNillableType sets the "type" field if the given value is not nil.
func (pu *PromptUpdate) SetNillableType(pr *prompt.Type) *PromptUpdate {
	if pr != nil {
		pu.SetType(*pr)
	}
	return pu
}

// SetUpdatedAt sets the "updated_at" field.
func (pu *PromptUpdate) SetUpdatedAt(t time.Time) *PromptUpdate {
	pu.mutation.SetUpdatedAt(t)
	return pu
}

// AddGenerationIDs adds the "generations" edge to the Generation entity by IDs.
func (pu *PromptUpdate) AddGenerationIDs(ids ...uuid.UUID) *PromptUpdate {
	pu.mutation.AddGenerationIDs(ids...)
	return pu
}

// AddGenerations adds the "generations" edges to the Generation entity.
func (pu *PromptUpdate) AddGenerations(g ...*Generation) *PromptUpdate {
	ids := make([]uuid.UUID, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return pu.AddGenerationIDs(ids...)
}

// AddVoiceoverIDs adds the "voiceovers" edge to the Voiceover entity by IDs.
func (pu *PromptUpdate) AddVoiceoverIDs(ids ...uuid.UUID) *PromptUpdate {
	pu.mutation.AddVoiceoverIDs(ids...)
	return pu
}

// AddVoiceovers adds the "voiceovers" edges to the Voiceover entity.
func (pu *PromptUpdate) AddVoiceovers(v ...*Voiceover) *PromptUpdate {
	ids := make([]uuid.UUID, len(v))
	for i := range v {
		ids[i] = v[i].ID
	}
	return pu.AddVoiceoverIDs(ids...)
}

// Mutation returns the PromptMutation object of the builder.
func (pu *PromptUpdate) Mutation() *PromptMutation {
	return pu.mutation
}

// ClearGenerations clears all "generations" edges to the Generation entity.
func (pu *PromptUpdate) ClearGenerations() *PromptUpdate {
	pu.mutation.ClearGenerations()
	return pu
}

// RemoveGenerationIDs removes the "generations" edge to Generation entities by IDs.
func (pu *PromptUpdate) RemoveGenerationIDs(ids ...uuid.UUID) *PromptUpdate {
	pu.mutation.RemoveGenerationIDs(ids...)
	return pu
}

// RemoveGenerations removes "generations" edges to Generation entities.
func (pu *PromptUpdate) RemoveGenerations(g ...*Generation) *PromptUpdate {
	ids := make([]uuid.UUID, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return pu.RemoveGenerationIDs(ids...)
}

// ClearVoiceovers clears all "voiceovers" edges to the Voiceover entity.
func (pu *PromptUpdate) ClearVoiceovers() *PromptUpdate {
	pu.mutation.ClearVoiceovers()
	return pu
}

// RemoveVoiceoverIDs removes the "voiceovers" edge to Voiceover entities by IDs.
func (pu *PromptUpdate) RemoveVoiceoverIDs(ids ...uuid.UUID) *PromptUpdate {
	pu.mutation.RemoveVoiceoverIDs(ids...)
	return pu
}

// RemoveVoiceovers removes "voiceovers" edges to Voiceover entities.
func (pu *PromptUpdate) RemoveVoiceovers(v ...*Voiceover) *PromptUpdate {
	ids := make([]uuid.UUID, len(v))
	for i := range v {
		ids[i] = v[i].ID
	}
	return pu.RemoveVoiceoverIDs(ids...)
}

// Save executes the query and returns the number of nodes affected by the update operation.
func (pu *PromptUpdate) Save(ctx context.Context) (int, error) {
	pu.defaults()
	return withHooks(ctx, pu.sqlSave, pu.mutation, pu.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (pu *PromptUpdate) SaveX(ctx context.Context) int {
	affected, err := pu.Save(ctx)
	if err != nil {
		panic(err)
	}
	return affected
}

// Exec executes the query.
func (pu *PromptUpdate) Exec(ctx context.Context) error {
	_, err := pu.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (pu *PromptUpdate) ExecX(ctx context.Context) {
	if err := pu.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (pu *PromptUpdate) defaults() {
	if _, ok := pu.mutation.UpdatedAt(); !ok {
		v := prompt.UpdateDefaultUpdatedAt()
		pu.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (pu *PromptUpdate) check() error {
	if v, ok := pu.mutation.GetType(); ok {
		if err := prompt.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "Prompt.type": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (pu *PromptUpdate) Modify(modifiers ...func(u *sql.UpdateBuilder)) *PromptUpdate {
	pu.modifiers = append(pu.modifiers, modifiers...)
	return pu
}

func (pu *PromptUpdate) sqlSave(ctx context.Context) (n int, err error) {
	if err := pu.check(); err != nil {
		return n, err
	}
	_spec := sqlgraph.NewUpdateSpec(prompt.Table, prompt.Columns, sqlgraph.NewFieldSpec(prompt.FieldID, field.TypeUUID))
	if ps := pu.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := pu.mutation.Text(); ok {
		_spec.SetField(prompt.FieldText, field.TypeString, value)
	}
	if value, ok := pu.mutation.TranslatedText(); ok {
		_spec.SetField(prompt.FieldTranslatedText, field.TypeString, value)
	}
	if pu.mutation.TranslatedTextCleared() {
		_spec.ClearField(prompt.FieldTranslatedText, field.TypeString)
	}
	if value, ok := pu.mutation.RanTranslation(); ok {
		_spec.SetField(prompt.FieldRanTranslation, field.TypeBool, value)
	}
	if value, ok := pu.mutation.GetType(); ok {
		_spec.SetField(prompt.FieldType, field.TypeEnum, value)
	}
	if value, ok := pu.mutation.UpdatedAt(); ok {
		_spec.SetField(prompt.FieldUpdatedAt, field.TypeTime, value)
	}
	if pu.mutation.GenerationsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := pu.mutation.RemovedGenerationsIDs(); len(nodes) > 0 && !pu.mutation.GenerationsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := pu.mutation.GenerationsIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if pu.mutation.VoiceoversCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := pu.mutation.RemovedVoiceoversIDs(); len(nodes) > 0 && !pu.mutation.VoiceoversCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := pu.mutation.VoiceoversIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(pu.modifiers...)
	if n, err = sqlgraph.UpdateNodes(ctx, pu.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{prompt.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return 0, err
	}
	pu.mutation.done = true
	return n, nil
}

// PromptUpdateOne is the builder for updating a single Prompt entity.
type PromptUpdateOne struct {
	config
	fields    []string
	hooks     []Hook
	mutation  *PromptMutation
	modifiers []func(*sql.UpdateBuilder)
}

// SetText sets the "text" field.
func (puo *PromptUpdateOne) SetText(s string) *PromptUpdateOne {
	puo.mutation.SetText(s)
	return puo
}

// SetNillableText sets the "text" field if the given value is not nil.
func (puo *PromptUpdateOne) SetNillableText(s *string) *PromptUpdateOne {
	if s != nil {
		puo.SetText(*s)
	}
	return puo
}

// SetTranslatedText sets the "translated_text" field.
func (puo *PromptUpdateOne) SetTranslatedText(s string) *PromptUpdateOne {
	puo.mutation.SetTranslatedText(s)
	return puo
}

// SetNillableTranslatedText sets the "translated_text" field if the given value is not nil.
func (puo *PromptUpdateOne) SetNillableTranslatedText(s *string) *PromptUpdateOne {
	if s != nil {
		puo.SetTranslatedText(*s)
	}
	return puo
}

// ClearTranslatedText clears the value of the "translated_text" field.
func (puo *PromptUpdateOne) ClearTranslatedText() *PromptUpdateOne {
	puo.mutation.ClearTranslatedText()
	return puo
}

// SetRanTranslation sets the "ran_translation" field.
func (puo *PromptUpdateOne) SetRanTranslation(b bool) *PromptUpdateOne {
	puo.mutation.SetRanTranslation(b)
	return puo
}

// SetNillableRanTranslation sets the "ran_translation" field if the given value is not nil.
func (puo *PromptUpdateOne) SetNillableRanTranslation(b *bool) *PromptUpdateOne {
	if b != nil {
		puo.SetRanTranslation(*b)
	}
	return puo
}

// SetType sets the "type" field.
func (puo *PromptUpdateOne) SetType(pr prompt.Type) *PromptUpdateOne {
	puo.mutation.SetType(pr)
	return puo
}

// SetNillableType sets the "type" field if the given value is not nil.
func (puo *PromptUpdateOne) SetNillableType(pr *prompt.Type) *PromptUpdateOne {
	if pr != nil {
		puo.SetType(*pr)
	}
	return puo
}

// SetUpdatedAt sets the "updated_at" field.
func (puo *PromptUpdateOne) SetUpdatedAt(t time.Time) *PromptUpdateOne {
	puo.mutation.SetUpdatedAt(t)
	return puo
}

// AddGenerationIDs adds the "generations" edge to the Generation entity by IDs.
func (puo *PromptUpdateOne) AddGenerationIDs(ids ...uuid.UUID) *PromptUpdateOne {
	puo.mutation.AddGenerationIDs(ids...)
	return puo
}

// AddGenerations adds the "generations" edges to the Generation entity.
func (puo *PromptUpdateOne) AddGenerations(g ...*Generation) *PromptUpdateOne {
	ids := make([]uuid.UUID, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return puo.AddGenerationIDs(ids...)
}

// AddVoiceoverIDs adds the "voiceovers" edge to the Voiceover entity by IDs.
func (puo *PromptUpdateOne) AddVoiceoverIDs(ids ...uuid.UUID) *PromptUpdateOne {
	puo.mutation.AddVoiceoverIDs(ids...)
	return puo
}

// AddVoiceovers adds the "voiceovers" edges to the Voiceover entity.
func (puo *PromptUpdateOne) AddVoiceovers(v ...*Voiceover) *PromptUpdateOne {
	ids := make([]uuid.UUID, len(v))
	for i := range v {
		ids[i] = v[i].ID
	}
	return puo.AddVoiceoverIDs(ids...)
}

// Mutation returns the PromptMutation object of the builder.
func (puo *PromptUpdateOne) Mutation() *PromptMutation {
	return puo.mutation
}

// ClearGenerations clears all "generations" edges to the Generation entity.
func (puo *PromptUpdateOne) ClearGenerations() *PromptUpdateOne {
	puo.mutation.ClearGenerations()
	return puo
}

// RemoveGenerationIDs removes the "generations" edge to Generation entities by IDs.
func (puo *PromptUpdateOne) RemoveGenerationIDs(ids ...uuid.UUID) *PromptUpdateOne {
	puo.mutation.RemoveGenerationIDs(ids...)
	return puo
}

// RemoveGenerations removes "generations" edges to Generation entities.
func (puo *PromptUpdateOne) RemoveGenerations(g ...*Generation) *PromptUpdateOne {
	ids := make([]uuid.UUID, len(g))
	for i := range g {
		ids[i] = g[i].ID
	}
	return puo.RemoveGenerationIDs(ids...)
}

// ClearVoiceovers clears all "voiceovers" edges to the Voiceover entity.
func (puo *PromptUpdateOne) ClearVoiceovers() *PromptUpdateOne {
	puo.mutation.ClearVoiceovers()
	return puo
}

// RemoveVoiceoverIDs removes the "voiceovers" edge to Voiceover entities by IDs.
func (puo *PromptUpdateOne) RemoveVoiceoverIDs(ids ...uuid.UUID) *PromptUpdateOne {
	puo.mutation.RemoveVoiceoverIDs(ids...)
	return puo
}

// RemoveVoiceovers removes "voiceovers" edges to Voiceover entities.
func (puo *PromptUpdateOne) RemoveVoiceovers(v ...*Voiceover) *PromptUpdateOne {
	ids := make([]uuid.UUID, len(v))
	for i := range v {
		ids[i] = v[i].ID
	}
	return puo.RemoveVoiceoverIDs(ids...)
}

// Where appends a list predicates to the PromptUpdate builder.
func (puo *PromptUpdateOne) Where(ps ...predicate.Prompt) *PromptUpdateOne {
	puo.mutation.Where(ps...)
	return puo
}

// Select allows selecting one or more fields (columns) of the returned entity.
// The default is selecting all fields defined in the entity schema.
func (puo *PromptUpdateOne) Select(field string, fields ...string) *PromptUpdateOne {
	puo.fields = append([]string{field}, fields...)
	return puo
}

// Save executes the query and returns the updated Prompt entity.
func (puo *PromptUpdateOne) Save(ctx context.Context) (*Prompt, error) {
	puo.defaults()
	return withHooks(ctx, puo.sqlSave, puo.mutation, puo.hooks)
}

// SaveX is like Save, but panics if an error occurs.
func (puo *PromptUpdateOne) SaveX(ctx context.Context) *Prompt {
	node, err := puo.Save(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// Exec executes the query on the entity.
func (puo *PromptUpdateOne) Exec(ctx context.Context) error {
	_, err := puo.Save(ctx)
	return err
}

// ExecX is like Exec, but panics if an error occurs.
func (puo *PromptUpdateOne) ExecX(ctx context.Context) {
	if err := puo.Exec(ctx); err != nil {
		panic(err)
	}
}

// defaults sets the default values of the builder before save.
func (puo *PromptUpdateOne) defaults() {
	if _, ok := puo.mutation.UpdatedAt(); !ok {
		v := prompt.UpdateDefaultUpdatedAt()
		puo.mutation.SetUpdatedAt(v)
	}
}

// check runs all checks and user-defined validators on the builder.
func (puo *PromptUpdateOne) check() error {
	if v, ok := puo.mutation.GetType(); ok {
		if err := prompt.TypeValidator(v); err != nil {
			return &ValidationError{Name: "type", err: fmt.Errorf(`ent: validator failed for field "Prompt.type": %w`, err)}
		}
	}
	return nil
}

// Modify adds a statement modifier for attaching custom logic to the UPDATE statement.
func (puo *PromptUpdateOne) Modify(modifiers ...func(u *sql.UpdateBuilder)) *PromptUpdateOne {
	puo.modifiers = append(puo.modifiers, modifiers...)
	return puo
}

func (puo *PromptUpdateOne) sqlSave(ctx context.Context) (_node *Prompt, err error) {
	if err := puo.check(); err != nil {
		return _node, err
	}
	_spec := sqlgraph.NewUpdateSpec(prompt.Table, prompt.Columns, sqlgraph.NewFieldSpec(prompt.FieldID, field.TypeUUID))
	id, ok := puo.mutation.ID()
	if !ok {
		return nil, &ValidationError{Name: "id", err: errors.New(`ent: missing "Prompt.id" for update`)}
	}
	_spec.Node.ID.Value = id
	if fields := puo.fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, prompt.FieldID)
		for _, f := range fields {
			if !prompt.ValidColumn(f) {
				return nil, &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
			}
			if f != prompt.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, f)
			}
		}
	}
	if ps := puo.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if value, ok := puo.mutation.Text(); ok {
		_spec.SetField(prompt.FieldText, field.TypeString, value)
	}
	if value, ok := puo.mutation.TranslatedText(); ok {
		_spec.SetField(prompt.FieldTranslatedText, field.TypeString, value)
	}
	if puo.mutation.TranslatedTextCleared() {
		_spec.ClearField(prompt.FieldTranslatedText, field.TypeString)
	}
	if value, ok := puo.mutation.RanTranslation(); ok {
		_spec.SetField(prompt.FieldRanTranslation, field.TypeBool, value)
	}
	if value, ok := puo.mutation.GetType(); ok {
		_spec.SetField(prompt.FieldType, field.TypeEnum, value)
	}
	if value, ok := puo.mutation.UpdatedAt(); ok {
		_spec.SetField(prompt.FieldUpdatedAt, field.TypeTime, value)
	}
	if puo.mutation.GenerationsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := puo.mutation.RemovedGenerationsIDs(); len(nodes) > 0 && !puo.mutation.GenerationsCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := puo.mutation.GenerationsIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	if puo.mutation.VoiceoversCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := puo.mutation.RemovedVoiceoversIDs(); len(nodes) > 0 && !puo.mutation.VoiceoversCleared() {
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
		_spec.Edges.Clear = append(_spec.Edges.Clear, edge)
	}
	if nodes := puo.mutation.VoiceoversIDs(); len(nodes) > 0 {
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
		_spec.Edges.Add = append(_spec.Edges.Add, edge)
	}
	_spec.AddModifiers(puo.modifiers...)
	_node = &Prompt{config: puo.config}
	_spec.Assign = _node.assignValues
	_spec.ScanValues = _node.scanValues
	if err = sqlgraph.UpdateNode(ctx, puo.driver, _spec); err != nil {
		if _, ok := err.(*sqlgraph.NotFoundError); ok {
			err = &NotFoundError{prompt.Label}
		} else if sqlgraph.IsConstraintError(err) {
			err = &ConstraintError{msg: err.Error(), wrap: err}
		}
		return nil, err
	}
	puo.mutation.done = true
	return _node, nil
}
