package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

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

// FindEntityByPublicID retrieves an entity by its URL-friendly name.
func (r *Repo) FindEntityByPublicID(ctx context.Context, publicID string, t ds.EntityType) (*ds.Entity, error) {
	_, span := r.tracer.Start(ctx, "FindEntityByPublicID")
	defer span.End()

	ent := new(ds.Entity)
	const query = `
		SELECT * FROM entities 
		WHERE public_id = $1 AND type=$2 AND deleted_at IS NULL`

	err := pgxscan.Get(ctx, r.getDB(ctx), ent, query, publicID, t)
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
		"id":              e.ID,
		"public_id":       e.PublicID,
		"preview_file_id": e.PreviewFileID,
		"owner_id":        e.OwnerID,
		"type":            e.Type,
		"title":           e.Title,
		"description":     e.Description,
		"visibility":      e.Visibility,
		"status":          e.Status,
		"created_at":      e.CreatedAt,
		"updated_at":      e.UpdatedAt,
		"deleted_at":      e.DeletedAt,
	})
}

// UpdateEntity updates both entity and book tables.
func (r *Repo) UpdateEntity(ctx context.Context, e *ds.Entity) error {
	_, span := r.tracer.Start(ctx, "UpdateBook")
	defer span.End()

	err := r.update(ctx, e.ID, "entities", data{
		"title":           e.Title,
		"description":     e.Description,
		"preview_file_id": e.PreviewFileID,
		"visibility":      e.Visibility,
		"updated_at":      time.Now(),
	})
	if err != nil {
		return fmt.Errorf("update entity: %w", err)
	}

	return nil
}
