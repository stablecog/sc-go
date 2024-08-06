// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"

	"entgo.io/ent"
	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/disposableemail"
	"github.com/stablecog/sc-go/database/ent/predicate"
)

// DisposableEmailQuery is the builder for querying DisposableEmail entities.
type DisposableEmailQuery struct {
	config
	ctx        *QueryContext
	order      []disposableemail.OrderOption
	inters     []Interceptor
	predicates []predicate.DisposableEmail
	modifiers  []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the DisposableEmailQuery builder.
func (deq *DisposableEmailQuery) Where(ps ...predicate.DisposableEmail) *DisposableEmailQuery {
	deq.predicates = append(deq.predicates, ps...)
	return deq
}

// Limit the number of records to be returned by this query.
func (deq *DisposableEmailQuery) Limit(limit int) *DisposableEmailQuery {
	deq.ctx.Limit = &limit
	return deq
}

// Offset to start from.
func (deq *DisposableEmailQuery) Offset(offset int) *DisposableEmailQuery {
	deq.ctx.Offset = &offset
	return deq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (deq *DisposableEmailQuery) Unique(unique bool) *DisposableEmailQuery {
	deq.ctx.Unique = &unique
	return deq
}

// Order specifies how the records should be ordered.
func (deq *DisposableEmailQuery) Order(o ...disposableemail.OrderOption) *DisposableEmailQuery {
	deq.order = append(deq.order, o...)
	return deq
}

// First returns the first DisposableEmail entity from the query.
// Returns a *NotFoundError when no DisposableEmail was found.
func (deq *DisposableEmailQuery) First(ctx context.Context) (*DisposableEmail, error) {
	nodes, err := deq.Limit(1).All(setContextOp(ctx, deq.ctx, ent.OpQueryFirst))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{disposableemail.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (deq *DisposableEmailQuery) FirstX(ctx context.Context) *DisposableEmail {
	node, err := deq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first DisposableEmail ID from the query.
// Returns a *NotFoundError when no DisposableEmail ID was found.
func (deq *DisposableEmailQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = deq.Limit(1).IDs(setContextOp(ctx, deq.ctx, ent.OpQueryFirstID)); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{disposableemail.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (deq *DisposableEmailQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := deq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single DisposableEmail entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one DisposableEmail entity is found.
// Returns a *NotFoundError when no DisposableEmail entities are found.
func (deq *DisposableEmailQuery) Only(ctx context.Context) (*DisposableEmail, error) {
	nodes, err := deq.Limit(2).All(setContextOp(ctx, deq.ctx, ent.OpQueryOnly))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{disposableemail.Label}
	default:
		return nil, &NotSingularError{disposableemail.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (deq *DisposableEmailQuery) OnlyX(ctx context.Context) *DisposableEmail {
	node, err := deq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only DisposableEmail ID in the query.
// Returns a *NotSingularError when more than one DisposableEmail ID is found.
// Returns a *NotFoundError when no entities are found.
func (deq *DisposableEmailQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = deq.Limit(2).IDs(setContextOp(ctx, deq.ctx, ent.OpQueryOnlyID)); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{disposableemail.Label}
	default:
		err = &NotSingularError{disposableemail.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (deq *DisposableEmailQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := deq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of DisposableEmails.
func (deq *DisposableEmailQuery) All(ctx context.Context) ([]*DisposableEmail, error) {
	ctx = setContextOp(ctx, deq.ctx, ent.OpQueryAll)
	if err := deq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*DisposableEmail, *DisposableEmailQuery]()
	return withInterceptors[[]*DisposableEmail](ctx, deq, qr, deq.inters)
}

// AllX is like All, but panics if an error occurs.
func (deq *DisposableEmailQuery) AllX(ctx context.Context) []*DisposableEmail {
	nodes, err := deq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of DisposableEmail IDs.
func (deq *DisposableEmailQuery) IDs(ctx context.Context) (ids []uuid.UUID, err error) {
	if deq.ctx.Unique == nil && deq.path != nil {
		deq.Unique(true)
	}
	ctx = setContextOp(ctx, deq.ctx, ent.OpQueryIDs)
	if err = deq.Select(disposableemail.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (deq *DisposableEmailQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := deq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (deq *DisposableEmailQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, deq.ctx, ent.OpQueryCount)
	if err := deq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, deq, querierCount[*DisposableEmailQuery](), deq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (deq *DisposableEmailQuery) CountX(ctx context.Context) int {
	count, err := deq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (deq *DisposableEmailQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, deq.ctx, ent.OpQueryExist)
	switch _, err := deq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (deq *DisposableEmailQuery) ExistX(ctx context.Context) bool {
	exist, err := deq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the DisposableEmailQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (deq *DisposableEmailQuery) Clone() *DisposableEmailQuery {
	if deq == nil {
		return nil
	}
	return &DisposableEmailQuery{
		config:     deq.config,
		ctx:        deq.ctx.Clone(),
		order:      append([]disposableemail.OrderOption{}, deq.order...),
		inters:     append([]Interceptor{}, deq.inters...),
		predicates: append([]predicate.DisposableEmail{}, deq.predicates...),
		// clone intermediate query.
		sql:  deq.sql.Clone(),
		path: deq.path,
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Domain string `json:"domain,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.DisposableEmail.Query().
//		GroupBy(disposableemail.FieldDomain).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (deq *DisposableEmailQuery) GroupBy(field string, fields ...string) *DisposableEmailGroupBy {
	deq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &DisposableEmailGroupBy{build: deq}
	grbuild.flds = &deq.ctx.Fields
	grbuild.label = disposableemail.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Domain string `json:"domain,omitempty"`
//	}
//
//	client.DisposableEmail.Query().
//		Select(disposableemail.FieldDomain).
//		Scan(ctx, &v)
func (deq *DisposableEmailQuery) Select(fields ...string) *DisposableEmailSelect {
	deq.ctx.Fields = append(deq.ctx.Fields, fields...)
	sbuild := &DisposableEmailSelect{DisposableEmailQuery: deq}
	sbuild.label = disposableemail.Label
	sbuild.flds, sbuild.scan = &deq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a DisposableEmailSelect configured with the given aggregations.
func (deq *DisposableEmailQuery) Aggregate(fns ...AggregateFunc) *DisposableEmailSelect {
	return deq.Select().Aggregate(fns...)
}

func (deq *DisposableEmailQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range deq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, deq); err != nil {
				return err
			}
		}
	}
	for _, f := range deq.ctx.Fields {
		if !disposableemail.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if deq.path != nil {
		prev, err := deq.path(ctx)
		if err != nil {
			return err
		}
		deq.sql = prev
	}
	return nil
}

func (deq *DisposableEmailQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*DisposableEmail, error) {
	var (
		nodes = []*DisposableEmail{}
		_spec = deq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*DisposableEmail).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &DisposableEmail{config: deq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	if len(deq.modifiers) > 0 {
		_spec.Modifiers = deq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, deq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	return nodes, nil
}

func (deq *DisposableEmailQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := deq.querySpec()
	if len(deq.modifiers) > 0 {
		_spec.Modifiers = deq.modifiers
	}
	_spec.Node.Columns = deq.ctx.Fields
	if len(deq.ctx.Fields) > 0 {
		_spec.Unique = deq.ctx.Unique != nil && *deq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, deq.driver, _spec)
}

func (deq *DisposableEmailQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(disposableemail.Table, disposableemail.Columns, sqlgraph.NewFieldSpec(disposableemail.FieldID, field.TypeUUID))
	_spec.From = deq.sql
	if unique := deq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if deq.path != nil {
		_spec.Unique = true
	}
	if fields := deq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, disposableemail.FieldID)
		for i := range fields {
			if fields[i] != disposableemail.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := deq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := deq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := deq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := deq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (deq *DisposableEmailQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(deq.driver.Dialect())
	t1 := builder.Table(disposableemail.Table)
	columns := deq.ctx.Fields
	if len(columns) == 0 {
		columns = disposableemail.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if deq.sql != nil {
		selector = deq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if deq.ctx.Unique != nil && *deq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range deq.modifiers {
		m(selector)
	}
	for _, p := range deq.predicates {
		p(selector)
	}
	for _, p := range deq.order {
		p(selector)
	}
	if offset := deq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := deq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (deq *DisposableEmailQuery) Modify(modifiers ...func(s *sql.Selector)) *DisposableEmailSelect {
	deq.modifiers = append(deq.modifiers, modifiers...)
	return deq.Select()
}

// DisposableEmailGroupBy is the group-by builder for DisposableEmail entities.
type DisposableEmailGroupBy struct {
	selector
	build *DisposableEmailQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (degb *DisposableEmailGroupBy) Aggregate(fns ...AggregateFunc) *DisposableEmailGroupBy {
	degb.fns = append(degb.fns, fns...)
	return degb
}

// Scan applies the selector query and scans the result into the given value.
func (degb *DisposableEmailGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, degb.build.ctx, ent.OpQueryGroupBy)
	if err := degb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*DisposableEmailQuery, *DisposableEmailGroupBy](ctx, degb.build, degb, degb.build.inters, v)
}

func (degb *DisposableEmailGroupBy) sqlScan(ctx context.Context, root *DisposableEmailQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(degb.fns))
	for _, fn := range degb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*degb.flds)+len(degb.fns))
		for _, f := range *degb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*degb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := degb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// DisposableEmailSelect is the builder for selecting fields of DisposableEmail entities.
type DisposableEmailSelect struct {
	*DisposableEmailQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (des *DisposableEmailSelect) Aggregate(fns ...AggregateFunc) *DisposableEmailSelect {
	des.fns = append(des.fns, fns...)
	return des
}

// Scan applies the selector query and scans the result into the given value.
func (des *DisposableEmailSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, des.ctx, ent.OpQuerySelect)
	if err := des.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*DisposableEmailQuery, *DisposableEmailSelect](ctx, des.DisposableEmailQuery, des, des.inters, v)
}

func (des *DisposableEmailSelect) sqlScan(ctx context.Context, root *DisposableEmailQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(des.fns))
	for _, fn := range des.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*des.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := des.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (des *DisposableEmailSelect) Modify(modifiers ...func(s *sql.Selector)) *DisposableEmailSelect {
	des.modifiers = append(des.modifiers, modifiers...)
	return des
}
