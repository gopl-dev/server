package repo

import (
	"context"
	"time"

	"github.com/gopl-dev/server/app/ds"
)

// CreateEventLog persists an EventLog entry to the database.
func (r *Repo) CreateEventLog(ctx context.Context, log *ds.EventLog) error {
	_, span := r.tracer.Start(ctx, "CreateEventLog")
	defer span.End()

	if log.ID.IsNil() {
		log.ID = ds.NewID()
	}

	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	return r.insert(ctx, "event_logs", data{
		"id":               log.ID,
		"user_id":          log.UserID,
		"entity_id":        log.EntityID,
		"entity_change_id": log.EntityChangeID,
		"type":             log.Type,
		"message":          log.Message,
		"meta":             log.Meta,
		"is_public":        log.IsPublic,
		"created_at":       log.CreatedAt,
	})
}
