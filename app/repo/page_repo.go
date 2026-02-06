package repo

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

var (
	// ErrPageNotFound is a sentinel error returned when page not found.
	ErrPageNotFound = app.ErrNotFound("page not found")
)

// GetPageByPublicID retrieves a page by its public ID.
func (r *Repo) GetPageByPublicID(ctx context.Context, publicID string) (*ds.Page, error) {
	_, span := r.tracer.Start(ctx, "GetPageByPublicID")
	defer span.End()

	page := new(ds.Page)
	const query = `SELECT * FROM entities e JOIN pages p USING (id) WHERE e.public_id = $1 AND e.type = $2 AND e.deleted_at IS NULL LIMIT 1`

	err := pgxscan.Get(ctx, r.getDB(ctx), page, query, publicID, ds.EntityTypePage)
	if noRows(err) {
		return nil, ErrPageNotFound
	}

	return page, err
}

// GetPageByID retrieves a page by its ID.
func (r *Repo) GetPageByID(ctx context.Context, id ds.ID) (*ds.Page, error) {
	_, span := r.tracer.Start(ctx, "GetPageByID")
	defer span.End()

	page := new(ds.Page)
	const query = `SELECT * FROM entities e JOIN pages p USING (id) WHERE e.id = $1 AND e.type = $2 AND e.deleted_at IS NULL LIMIT 1`

	err := pgxscan.Get(ctx, r.getDB(ctx), page, query, id, ds.EntityTypePage)
	if noRows(err) {
		return nil, ErrPageNotFound
	}

	return page, err
}

// CreatePage inserts a new page record into the database.
func (r *Repo) CreatePage(ctx context.Context, p *ds.Page) error {
	_, span := r.tracer.Start(ctx, "CreatePage")
	defer span.End()

	return r.insert(ctx, "pages", data{
		"id":          p.ID,
		"content_raw": p.ContentRaw,
		"content":     p.Content,
	})
}

// UpdatePage updates the stored content of an existing page.
func (r *Repo) UpdatePage(ctx context.Context, p *ds.Page) error {
	_, span := r.tracer.Start(ctx, "UpdatePage")
	defer span.End()

	err := r.update(ctx, p.ID, "pages", data{
		"content_raw": p.ContentRaw,
		"content":     p.Content,
	})
	if err != nil {
		return fmt.Errorf("update page: %w", err)
	}

	return nil
}
