package repo

import (
	"context"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

var (
	// ErrEntityChangeRequestNotFound is a sentinel error returned when change request not found.
	ErrEntityChangeRequestNotFound = app.ErrNotFound("change request not found")
)

// GetPendingChangeRequest retrieves the most recent pending change request for a specific entity and user.
func (r *Repo) GetPendingChangeRequest(ctx context.Context, entityID, userID ds.ID) (*ds.EntityChangeRequest, error) {
	ctx, span := r.tracer.Start(ctx, "GetPendingChangeRequest")
	defer span.End()

	const query = `SELECT * FROM entity_change_requests WHERE entity_id = $1 AND user_id = $2 AND status = $3 ORDER BY updated_at DESC NULLS LAST, created_at DESC LIMIT 1`

	req := new(ds.EntityChangeRequest)
	err := pgxscan.Get(ctx, r.getDB(ctx), req, query, entityID, userID, ds.EntityChangePending)
	if noRows(err) {
		return nil, ErrEntityChangeRequestNotFound
	}

	return req, err
}

// CreateChangeRequest creates a new entity change request record.
func (r *Repo) CreateChangeRequest(ctx context.Context, m *ds.EntityChangeRequest) error {
	ctx, span := r.tracer.Start(ctx, "CreateChangeRequest")
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

// UpdateChangeRequest updates an existing entity change request record.
func (r *Repo) UpdateChangeRequest(ctx context.Context, m *ds.EntityChangeRequest) error {
	ctx, span := r.tracer.Start(ctx, "UpdateChangeRequest")
	defer span.End()

	err := r.update(ctx, m.ID, "entity_change_requests", data{
		"diff":       m.Diff,
		"message":    m.Message,
		"revision":   m.Revision,
		"updated_at": m.UpdatedAt,
	})
	if err != nil {
		return fmt.Errorf("update change request: %w", err)
	}

	return nil
}

// FilterChangeRequests retrieves a paginated list of change requests matching the given filter.
func (r *Repo) FilterChangeRequests(ctx context.Context, f ds.ChangeRequestsFilter) (reqs []ds.EntityChangeRequest, count int, err error) {
	_, span := r.tracer.Start(ctx, "FilterChangeRequests")
	defer span.End()

	b := r.filter("entity_change_requests r", "r").
		join("JOIN users u ON u.id = r.user_id").
		join("JOIN entities e ON e.id = r.entity_id").
		columns(`
			r.id as "id",
			r.status,
			r.created_at,

			u.username as "username",

			e.type as "entity_type",
			e.title as "entity_title",
			e.public_id as "entity_public_id"
`).
		paginate(f.Page, f.PerPage).
		where("r.status", f.Status).
		whereRaw("e.deleted_at IS NULL").
		withCount(f.WithCount).
		order("r.created_at", "asc").
		withoutSoftDelete()

	count, err = b.scan(ctx, &reqs)
	return
}

// GetChangeRequestByID retrieves an entity change request by its ID.
func (r *Repo) GetChangeRequestByID(ctx context.Context, id ds.ID) (req *ds.EntityChangeRequest, err error) {
	_, span := r.tracer.Start(ctx, "GetChangeRequestByID")
	defer span.End()

	req = new(ds.EntityChangeRequest)
	const query = `SELECT 
    		r.id as "id",
    		r.entity_id,
    		r.user_id,
			r.status,
			r.diff,
			r.created_at,

			e.type as "entity_type"
    FROM entity_change_requests r 
    JOIN entities e ON r.entity_id = e.id
    WHERE r.id = $1`

	err = pgxscan.Get(ctx, r.getDB(ctx), req, query, id)
	if noRows(err) {
		return nil, ErrEntityChangeRequestNotFound
	}

	return req, err
}

// CommitChangeRequest marks a change request as committed.
func (r *Repo) CommitChangeRequest(ctx context.Context, req *ds.EntityChangeRequest) error {
	_, span := r.tracer.Start(ctx, "CommitChangeRequest")
	defer span.End()

	err := r.update(ctx, req.ID, "entity_change_requests", data{
		"status":      ds.EntityChangeCommitted,
		"reviewer_id": req.ReviewerID,
		"reviewed_at": time.Now(),
		"updated_at":  time.Now(),
	})
	if err != nil {
		return fmt.Errorf("commit change request: %w", err)
	}

	return nil
}

// RejectChangeRequest marks a change request as rejected.
func (r *Repo) RejectChangeRequest(ctx context.Context, id, reviewerID ds.ID, note string) (err error) {
	_, span := r.tracer.Start(ctx, "RejectChangeRequest")
	defer span.End()

	err = r.update(ctx, id, "entity_change_requests", data{
		"status":      ds.EntityChangeRejected,
		"reviewer_id": reviewerID,
		"review_note": note,
		"reviewed_at": time.Now(),
		"updated_at":  time.Now(),
	})
	if err != nil {
		return fmt.Errorf("reject change request: %w", err)
	}

	return nil
}
