package repo

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app/ds"
)

var (
	// ErrEntityNotFound is a sentinel error returned when entity is not found.
	ErrEntityNotFound = errors.New("entity not found")
)

// DeleteEntity marks an entity as deleted.
func (r *Repo) DeleteEntity(ctx context.Context, id ds.ID) error {
	return r.delete(ctx, "entities", id)
}

// FindEntityByURLName retrieves a entity by its URL-friendly name.
func (r *Repo) FindEntityByURLName(ctx context.Context, name string) (*ds.Entity, error) {
	_, span := r.tracer.Start(ctx, "FindEntityByURLName")
	defer span.End()

	ent := new(ds.Entity)
	const query = `
		SELECT * FROM entities 
		WHERE url_name = $1 AND deleted_at IS NULL`

	err := pgxscan.Get(ctx, r.getDB(ctx), ent, query, name)
	if noRows(err) {
		return nil, ErrEntityNotFound
	}

	return ent, err
}

// CreateEntity inserts entity.
func (r *Repo) CreateEntity(ctx context.Context, e *ds.Entity) error {
	_, span := r.tracer.Start(ctx, "CreateEntity")
	defer span.End()

	return r.insert(ctx, "entities", data{
		"id":         e.ID,
		"owner_id":   e.OwnerID,
		"type":       e.Type,
		"url_name":   e.URLName,
		"title":      e.Title,
		"visibility": e.Visibility,
		"status":     e.Status,
		"created_at": e.CreatedAt,
	})
}
