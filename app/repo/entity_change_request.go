package repo

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app/ds"
)

// FindPendingEntityChangeRequest ...
func (r *Repo) FindPendingEntityChangeRequest(ctx context.Context, entityID, userID ds.ID) (*ds.EntityChangeRequest, error) {
	ctx, span := r.tracer.Start(ctx, "FindPendingEntityChangeRequest")
	defer span.End()

	const query = `SELECT * FROM entity_change_requests WHERE entity_id = $1 AND user_id = $2 AND status = $3 ORDER BY updated_at DESC NULLS LAST, created_at DESC LIMIT 1`

	req := new(ds.EntityChangeRequest)
	err := pgxscan.Get(ctx, r.getDB(ctx), req, query, entityID, userID, ds.EntityChangePending)
	if noRows(err) {
		return nil, nil
	}

	return req, err
}

// CreateEntityChangeRequest creates a new entity change request record.
func (r *Repo) CreateEntityChangeRequest(ctx context.Context, m *ds.EntityChangeRequest) error {
	ctx, span := r.tracer.Start(ctx, "CreateEntityChangeRequest")
	defer span.End()

	err := r.insert(ctx, "entity_change_requests", data{
		"id":          m.ID,
		"entity_id":   m.EntityID,
		"user_id":     m.UserID,
		"status":      m.Status,
		"diff":        m.Diff,
		"message":     m.Message,
		"revision":    m.Revision,
		"reviewer_id": m.ReviewerID,
		"reviewed_at": m.ReviewedAt,
		"review_note": m.ReviewNote,
		"created_at":  m.CreatedAt,
		"updated_at":  m.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("create entity change request: %w", err)
	}

	return nil
}

// UpdateEntityChangeRequest updates an existing entity change request record.
func (r *Repo) UpdateEntityChangeRequest(ctx context.Context, m *ds.EntityChangeRequest) error {
	ctx, span := r.tracer.Start(ctx, "UpdateEntityChangeRequest")
	defer span.End()

	err := r.update(ctx, m.ID, "entity_change_requests", data{
		"diff":       m.Diff,
		"message":    m.Message,
		"revision":   m.Revision,
		"updated_at": m.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("update entity change request: %w", err)
	}

	return nil
}
