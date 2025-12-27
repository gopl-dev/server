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
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

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

func (r *Repo) insert(ctx context.Context, table string, data map[string]any) (row pgx.Row, err error) {
	sql, args, err := sq.Insert(table).
		SetMap(data).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		return
	}

	row = r.db.QueryRow(ctx, sql, args...)
	return
}

// filterBuilder wraps squirrel.SelectBuilder and provides common helper to handle most gotchas with selecting data.
type filterBuilder struct {
	qb             sq.SelectBuilder
	db             *app.DB
	columnsSet     bool
	whereDeleted   bool
	whereDeletedAt bool
	selectCount    bool
}

func (r *Repo) filter(table string) *filterBuilder {
	q := sq.Select().From(table).PlaceholderFormat(sq.Dollar)

	return &filterBuilder{
		qb: q,
		db: r.db,
	}
}

func (b *filterBuilder) columns(columns ...string) *filterBuilder { //nolint:unparam
	b.qb = b.qb.Columns(columns...)
	b.columnsSet = true

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
		return b.atDT("deleted_at", dt.DT)
	}

	return b.dtRange("deleted_at", dt)
}

func (b *filterBuilder) createdAt(dt *ds.FilterDT) *filterBuilder {
	if dt == nil {
		return b
	}

	if dt.DT != nil {
		b.atDT("created_at", dt.DT)
		return b
	}

	b.dtRange("created_at", dt)
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

	if strings.ToLower(direction) != "asc" {
		direction = "DESC"
	}

	b.qb = b.qb.OrderBy(column + " " + direction)
	return b
}

func (b *filterBuilder) apply(filters ...filterFn) *filterBuilder {
	for _, fn := range filters {
		b.qb = fn(b.qb)
	}
	return b
}

func (b *filterBuilder) withCount(flag bool) *filterBuilder {
	b.selectCount = flag
	return b
}

func (b *filterBuilder) sql() (sql string, args []any, err error) {
	lb := *b
	if !lb.columnsSet {
		lb.columns("*")
	}

	if !lb.whereDeletedAt {
		if lb.whereDeleted {
			lb.qb = lb.qb.Where("deleted_at IS NOT NULL")
		} else {
			lb.qb = lb.qb.Where("deleted_at IS NULL")
		}
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

	return lb.qb.
		RemoveLimit().
		RemoveOffset().
		OrderBy().
		ToSql()
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

type filterFn func(sq.SelectBuilder) sq.SelectBuilder

func noRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
