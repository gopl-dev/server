// Package repo ...
package repo

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

type data map[string]any

// dbKey is a private type to avoid collisions in context.WithValue.
type dbKey struct{}

// DBI matches the common methods between pgxpool.Pool and pgx.Tx.
// This allows repository methods to work whether they are in a transaction or not.
type DBI interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Repo is the primary struct for database access operations
// All repository methods are attached to this type.
type Repo struct {
	db     *app.DB
	tracer trace.Tracer
}

// New is a factory function that creates and returns a new Repo instance.
func New(db *app.DB, t trace.Tracer) *Repo {
	return &Repo{
		db:     db,
		tracer: t,
	}
}

// WithTx wraps app.RunInTx and puts the transaction into the context.
func (r *Repo) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return app.RunInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		// We store the transaction in the context
		ctx = context.WithValue(ctx, dbKey{}, tx)
		return fn(ctx)
	})
}

// getDB returns the transaction from context if it exists, otherwise returns the pool.
func (r *Repo) getDB(ctx context.Context) DBI {
	if tx, ok := ctx.Value(dbKey{}).(pgx.Tx); ok {
		return tx
	}

	return r.db
}

// insert inserts a data map into the DB.
// (If another method like insertSomething is needed later, rename this to insertMap;
// until then, it remains simply insert).
func (r *Repo) insert(ctx context.Context, table string, values data) (err error) {
	sql, args, err := sq.Insert(table).
		SetMap(values).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return
	}

	_, err = r.getDB(ctx).Exec(ctx, sql, args...)
	return
}

func (r *Repo) update(ctx context.Context, id ds.ID, table string, values data) (err error) {
	sql, args, err := sq.Update(table).
		SetMap(values).
		Where("id = ?", id).
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return
	}

	_, err = r.getDB(ctx).Exec(ctx, sql, args...)
	return
}

// delete performs a soft delete by setting the deleted_at timestamp to the current time.
// It marks a record as deleted without actually removing it from the database.
func (r *Repo) delete(ctx context.Context, table string, id ds.ID) (err error) {
	sql, args, err := sq.Update(table).
		PlaceholderFormat(sq.Dollar).
		Set("deleted_at", "NOW()").
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return
	}
	_, err = r.getDB(ctx).Exec(ctx, sql, args...)
	return
}

// hardDelete permanently removes a record from the database table by its ID.
// Unlike a soft delete, this action cannot be undone.
func (r *Repo) hardDelete(ctx context.Context, table string, id ds.ID) (err error) {
	sql, args, err := sq.Delete(table).
		PlaceholderFormat(sq.Dollar).
		Where("id = ?", id).
		ToSql()
	if err != nil {
		return
	}
	_, err = r.getDB(ctx).Exec(ctx, sql, args...)
	return
}

// insert inserts a data map into the DB.
// (If another method like insertSomething is needed later, rename this to insertMap;
// until then, it remains simply insert).
func (r *Repo) exec(ctx context.Context, query string, args ...any) (err error) {
	_, err = r.getDB(ctx).Exec(ctx, query, args...)
	return
}

// filterBuilder wraps squirrel.SelectBuilder and provides common helper to handle most gotchas with selecting data.
type filterBuilder struct {
	qb             sq.SelectBuilder
	db             *app.DB
	selector       string
	columnsSet     bool
	whereDeleted   bool
	whereDeletedAt bool
	selectCount    bool
	orderBy        string
}

func (r *Repo) filter(table string, selectorOpt ...string) *filterBuilder {
	q := sq.Select().From(table).PlaceholderFormat(sq.Dollar)

	b := &filterBuilder{
		qb: q,
		db: r.db,
	}

	if len(selectorOpt) > 0 {
		b.selector = selectorOpt[0] + "."
	}

	return b
}

func (b *filterBuilder) columns(columns ...string) *filterBuilder {
	b.qb = b.qb.Columns(columns...)
	b.columnsSet = true

	return b
}

func (b *filterBuilder) join(clause string, args ...any) *filterBuilder {
	b.qb = b.qb.JoinClause(clause, args...)

	return b
}

func (b *filterBuilder) paginate(page, perPage int) *filterBuilder {
	if perPage <= 0 {
		perPage = ds.PerPageDefault
	}
	if perPage > ds.PerPageMax {
		perPage = ds.PerPageMax
	}

	if page < 1 {
		page = 1
	}

	// page 1 -> offset 0; page 2 -> offset 25
	offset := (page - 1) * perPage

	b.qb = b.qb.Limit(uint64(perPage)) //nolint:gosec

	if offset > 0 {
		b.qb = b.qb.Offset(uint64(offset))
	}

	return b
}

func (b *filterBuilder) deleted(flag bool) *filterBuilder {
	b.whereDeleted = flag
	return b
}

func (b *filterBuilder) deletedAt(dt *ds.FilterDT) *filterBuilder {
	if dt == nil {
		return b
	}

	b.whereDeletedAt = true

	if dt.DT != nil {
		return b.atDT(b.selector+"deleted_at", dt.DT)
	}

	return b.dtRange(b.selector+"deleted_at", dt)
}

func (b *filterBuilder) createdAt(dt *ds.FilterDT) *filterBuilder {
	if dt == nil {
		return b
	}

	if dt.DT != nil {
		b.atDT(b.selector+"created_at", dt.DT)
		return b
	}

	b.dtRange(b.selector+"created_at", dt)
	return b
}

func (b *filterBuilder) dtRange(column string, dt *ds.FilterDT) *filterBuilder {
	if dt == nil {
		return b
	}

	if dt.To != nil {
		b.qb = b.qb.Where(column+" < ?", dt.To)
	}
	if dt.From != nil {
		b.qb = b.qb.Where(column+" > ?", dt.From)
	}

	return b
}

func (b *filterBuilder) atDT(column string, dt *time.Time) *filterBuilder {
	if dt == nil {
		return b
	}

	b.qb = b.qb.Where(column+"::datetime = ?", dt)
	return b
}

func (b *filterBuilder) order(column string, direction string) *filterBuilder {
	if column == "" {
		return b
	}

	if !strings.Contains(column, ".") {
		column = b.selector + column
	}

	if strings.ToLower(direction) != "asc" {
		direction = "DESC"
	}

	b.orderBy = column + " " + direction
	return b
}

func (b *filterBuilder) where(column string, val any) *filterBuilder {
	b.qb = b.qb.Where(sq.Eq{column: val})

	return b
}

func (b *filterBuilder) withCount(ok bool) *filterBuilder {
	b.selectCount = ok
	return b
}

func (b *filterBuilder) sql() (sql string, args []any, err error) {
	lb := *b
	if !lb.columnsSet {
		lb.columns("*")
	}

	lb.applyDeletedFilter()

	if lb.orderBy != "" {
		lb.qb = lb.qb.OrderBy(b.orderBy)
	}

	sql, args, err = lb.qb.ToSql()
	if err != nil {
		err = fmt.Errorf("failed to build SQL: %w", err)
	}

	return
}

func (b *filterBuilder) countSQL() (sql string, args []any, err error) {
	lb := *b

	lb.qb = lb.qb.RemoveColumns()
	lb.columns("COUNT(*)")

	lb.applyDeletedFilter()

	return lb.qb.
		RemoveLimit().
		RemoveOffset().
		ToSql()
}

// applyDeletedFilter applies a default soft-delete condition.
//
// If no explicit `deleted_at` condition was added earlier, it applies a filter
// based on the `whereDeleted` flag:
//   - when `whereDeleted` is true, only soft-deleted records are selected
//     (`deleted_at IS NOT NULL`);
//   - otherwise, only non-deleted records are selected
//     (`deleted_at IS NULL`).
func (b *filterBuilder) applyDeletedFilter() {
	if !b.whereDeletedAt {
		if b.whereDeleted {
			b.qb = b.qb.Where(b.selector + "deleted_at IS NOT NULL")
			return
		}

		b.qb = b.qb.Where(b.selector + "deleted_at IS NULL")
	}
}

func (b *filterBuilder) scan(ctx context.Context, desc any) (count int, err error) {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		sql, args, err := b.sql()
		if err != nil {
			return err
		}

		return pgxscan.Select(ctx, b.db, desc, sql, args...)
	})

	if b.selectCount {
		eg.Go(func() error {
			sql, args, err := b.countSQL()
			if err != nil {
				return err
			}

			return b.db.QueryRow(ctx, sql, args...).Scan(&count)
		})
	}

	err = eg.Wait()
	return
}

func (b *filterBuilder) apply(filters ...filterFn) *filterBuilder {
	for _, fn := range filters {
		b.qb = fn(b.qb)
	}

	return b
}

type filterFn func(sq.SelectBuilder) sq.SelectBuilder

// whereIn builds a filter function that conditionally applies an equality or IN
// clause for the given column based on the number of values provided.
//
// Behavior:
//   - len(val) == 0: no condition is applied (the filter is skipped)
//   - len(val) == 1: applies `column = val[0]`
//   - len(val) > 1: applies `column IN (val...)`
//
// This helper exists as a standalone function because Go currently does not
// allow type parameters on methods. Using a generic free function keeps the
// filterBuilder API non-generic while still providing type-safe slice handling.
//
// This helper is intended to avoid generating `IN ()` clauses (which Squirrel
// turns into `(1=0)`) and to keep slice-specific logic out of the query builder.
func whereIn[T any](column string, val []T) filterFn {
	return func(sb sq.SelectBuilder) sq.SelectBuilder {
		if len(val) == 0 {
			return sb
		}

		if len(val) == 1 {
			sb = sb.Where(sq.Eq{column: val[0]})
			return sb
		}

		sb = sb.Where(sq.Eq{column: val})
		return sb
	}
}

func noRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
