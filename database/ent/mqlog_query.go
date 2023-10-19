// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"github.com/stablecog/sc-go/database/ent/mqlog"
	"github.com/stablecog/sc-go/database/ent/predicate"
)

// MqLogQuery is the builder for querying MqLog entities.
type MqLogQuery struct {
	config
	ctx        *QueryContext
	order      []OrderFunc
	inters     []Interceptor
	predicates []predicate.MqLog
	modifiers  []func(*sql.Selector)
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the MqLogQuery builder.
func (mlq *MqLogQuery) Where(ps ...predicate.MqLog) *MqLogQuery {
	mlq.predicates = append(mlq.predicates, ps...)
	return mlq
}

// Limit the number of records to be returned by this query.
func (mlq *MqLogQuery) Limit(limit int) *MqLogQuery {
	mlq.ctx.Limit = &limit
	return mlq
}

// Offset to start from.
func (mlq *MqLogQuery) Offset(offset int) *MqLogQuery {
	mlq.ctx.Offset = &offset
	return mlq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (mlq *MqLogQuery) Unique(unique bool) *MqLogQuery {
	mlq.ctx.Unique = &unique
	return mlq
}

// Order specifies how the records should be ordered.
func (mlq *MqLogQuery) Order(o ...OrderFunc) *MqLogQuery {
	mlq.order = append(mlq.order, o...)
	return mlq
}

// First returns the first MqLog entity from the query.
// Returns a *NotFoundError when no MqLog was found.
func (mlq *MqLogQuery) First(ctx context.Context) (*MqLog, error) {
	nodes, err := mlq.Limit(1).All(setContextOp(ctx, mlq.ctx, "First"))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{mqlog.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (mlq *MqLogQuery) FirstX(ctx context.Context) *MqLog {
	node, err := mlq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first MqLog ID from the query.
// Returns a *NotFoundError when no MqLog ID was found.
func (mlq *MqLogQuery) FirstID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = mlq.Limit(1).IDs(setContextOp(ctx, mlq.ctx, "FirstID")); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{mqlog.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (mlq *MqLogQuery) FirstIDX(ctx context.Context) uuid.UUID {
	id, err := mlq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single MqLog entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one MqLog entity is found.
// Returns a *NotFoundError when no MqLog entities are found.
func (mlq *MqLogQuery) Only(ctx context.Context) (*MqLog, error) {
	nodes, err := mlq.Limit(2).All(setContextOp(ctx, mlq.ctx, "Only"))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{mqlog.Label}
	default:
		return nil, &NotSingularError{mqlog.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (mlq *MqLogQuery) OnlyX(ctx context.Context) *MqLog {
	node, err := mlq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only MqLog ID in the query.
// Returns a *NotSingularError when more than one MqLog ID is found.
// Returns a *NotFoundError when no entities are found.
func (mlq *MqLogQuery) OnlyID(ctx context.Context) (id uuid.UUID, err error) {
	var ids []uuid.UUID
	if ids, err = mlq.Limit(2).IDs(setContextOp(ctx, mlq.ctx, "OnlyID")); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{mqlog.Label}
	default:
		err = &NotSingularError{mqlog.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (mlq *MqLogQuery) OnlyIDX(ctx context.Context) uuid.UUID {
	id, err := mlq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of MqLogs.
func (mlq *MqLogQuery) All(ctx context.Context) ([]*MqLog, error) {
	ctx = setContextOp(ctx, mlq.ctx, "All")
	if err := mlq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*MqLog, *MqLogQuery]()
	return withInterceptors[[]*MqLog](ctx, mlq, qr, mlq.inters)
}

// AllX is like All, but panics if an error occurs.
func (mlq *MqLogQuery) AllX(ctx context.Context) []*MqLog {
	nodes, err := mlq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of MqLog IDs.
func (mlq *MqLogQuery) IDs(ctx context.Context) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	ctx = setContextOp(ctx, mlq.ctx, "IDs")
	if err := mlq.Select(mqlog.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (mlq *MqLogQuery) IDsX(ctx context.Context) []uuid.UUID {
	ids, err := mlq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (mlq *MqLogQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, mlq.ctx, "Count")
	if err := mlq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, mlq, querierCount[*MqLogQuery](), mlq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (mlq *MqLogQuery) CountX(ctx context.Context) int {
	count, err := mlq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (mlq *MqLogQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, mlq.ctx, "Exist")
	switch _, err := mlq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (mlq *MqLogQuery) ExistX(ctx context.Context) bool {
	exist, err := mlq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the MqLogQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (mlq *MqLogQuery) Clone() *MqLogQuery {
	if mlq == nil {
		return nil
	}
	return &MqLogQuery{
		config:     mlq.config,
		ctx:        mlq.ctx.Clone(),
		order:      append([]OrderFunc{}, mlq.order...),
		inters:     append([]Interceptor{}, mlq.inters...),
		predicates: append([]predicate.MqLog{}, mlq.predicates...),
		// clone intermediate query.
		sql:  mlq.sql.Clone(),
		path: mlq.path,
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		MessageID string `json:"message_id,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.MqLog.Query().
//		GroupBy(mqlog.FieldMessageID).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (mlq *MqLogQuery) GroupBy(field string, fields ...string) *MqLogGroupBy {
	mlq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &MqLogGroupBy{build: mlq}
	grbuild.flds = &mlq.ctx.Fields
	grbuild.label = mqlog.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		MessageID string `json:"message_id,omitempty"`
//	}
//
//	client.MqLog.Query().
//		Select(mqlog.FieldMessageID).
//		Scan(ctx, &v)
func (mlq *MqLogQuery) Select(fields ...string) *MqLogSelect {
	mlq.ctx.Fields = append(mlq.ctx.Fields, fields...)
	sbuild := &MqLogSelect{MqLogQuery: mlq}
	sbuild.label = mqlog.Label
	sbuild.flds, sbuild.scan = &mlq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a MqLogSelect configured with the given aggregations.
func (mlq *MqLogQuery) Aggregate(fns ...AggregateFunc) *MqLogSelect {
	return mlq.Select().Aggregate(fns...)
}

func (mlq *MqLogQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range mlq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, mlq); err != nil {
				return err
			}
		}
	}
	for _, f := range mlq.ctx.Fields {
		if !mqlog.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if mlq.path != nil {
		prev, err := mlq.path(ctx)
		if err != nil {
			return err
		}
		mlq.sql = prev
	}
	return nil
}

func (mlq *MqLogQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*MqLog, error) {
	var (
		nodes = []*MqLog{}
		_spec = mlq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*MqLog).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &MqLog{config: mlq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	if len(mlq.modifiers) > 0 {
		_spec.Modifiers = mlq.modifiers
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, mlq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	return nodes, nil
}

func (mlq *MqLogQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := mlq.querySpec()
	if len(mlq.modifiers) > 0 {
		_spec.Modifiers = mlq.modifiers
	}
	_spec.Node.Columns = mlq.ctx.Fields
	if len(mlq.ctx.Fields) > 0 {
		_spec.Unique = mlq.ctx.Unique != nil && *mlq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, mlq.driver, _spec)
}

func (mlq *MqLogQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := &sqlgraph.QuerySpec{
		Node: &sqlgraph.NodeSpec{
			Table:   mqlog.Table,
			Columns: mqlog.Columns,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeUUID,
				Column: mqlog.FieldID,
			},
		},
		From:   mlq.sql,
		Unique: true,
	}
	if unique := mlq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	}
	if fields := mlq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, mqlog.FieldID)
		for i := range fields {
			if fields[i] != mqlog.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := mlq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := mlq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := mlq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := mlq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (mlq *MqLogQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(mlq.driver.Dialect())
	t1 := builder.Table(mqlog.Table)
	columns := mlq.ctx.Fields
	if len(columns) == 0 {
		columns = mqlog.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if mlq.sql != nil {
		selector = mlq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if mlq.ctx.Unique != nil && *mlq.ctx.Unique {
		selector.Distinct()
	}
	for _, m := range mlq.modifiers {
		m(selector)
	}
	for _, p := range mlq.predicates {
		p(selector)
	}
	for _, p := range mlq.order {
		p(selector)
	}
	if offset := mlq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := mlq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// Modify adds a query modifier for attaching custom logic to queries.
func (mlq *MqLogQuery) Modify(modifiers ...func(s *sql.Selector)) *MqLogSelect {
	mlq.modifiers = append(mlq.modifiers, modifiers...)
	return mlq.Select()
}

// MqLogGroupBy is the group-by builder for MqLog entities.
type MqLogGroupBy struct {
	selector
	build *MqLogQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (mlgb *MqLogGroupBy) Aggregate(fns ...AggregateFunc) *MqLogGroupBy {
	mlgb.fns = append(mlgb.fns, fns...)
	return mlgb
}

// Scan applies the selector query and scans the result into the given value.
func (mlgb *MqLogGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, mlgb.build.ctx, "GroupBy")
	if err := mlgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*MqLogQuery, *MqLogGroupBy](ctx, mlgb.build, mlgb, mlgb.build.inters, v)
}

func (mlgb *MqLogGroupBy) sqlScan(ctx context.Context, root *MqLogQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(mlgb.fns))
	for _, fn := range mlgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*mlgb.flds)+len(mlgb.fns))
		for _, f := range *mlgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*mlgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := mlgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// MqLogSelect is the builder for selecting fields of MqLog entities.
type MqLogSelect struct {
	*MqLogQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (mls *MqLogSelect) Aggregate(fns ...AggregateFunc) *MqLogSelect {
	mls.fns = append(mls.fns, fns...)
	return mls
}

// Scan applies the selector query and scans the result into the given value.
func (mls *MqLogSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, mls.ctx, "Select")
	if err := mls.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*MqLogQuery, *MqLogSelect](ctx, mls.MqLogQuery, mls, mls.inters, v)
}

func (mls *MqLogSelect) sqlScan(ctx context.Context, root *MqLogQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(mls.fns))
	for _, fn := range mls.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*mls.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := mls.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// Modify adds a query modifier for attaching custom logic to queries.
func (mls *MqLogSelect) Modify(modifiers ...func(s *sql.Selector)) *MqLogSelect {
	mls.modifiers = append(mls.modifiers, modifiers...)
	return mls
}