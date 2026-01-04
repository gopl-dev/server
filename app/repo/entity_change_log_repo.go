package repo

import (
	"context"

	"github.com/gopl-dev/server/app/ds"
)

// CreateEntityChangeLog inserts a new audit record.
func (r *Repo) CreateEntityChangeLog(ctx context.Context, log *ds.EntityChangeLog) error {
	_, span := r.tracer.Start(ctx, "CreateEntityChangeLog")
	defer span.End()

	if log.ID.IsNil() {
		log.ID = ds.NewID()
	}

	return r.insert(ctx, "entity_change_logs", data{
		"id":         log.ID,
		"entity_id":  log.EntityID,
		"user_id":    log.UserID,
		"action":     log.Action,
		"created_at": log.CreatedAt,
	})
}
