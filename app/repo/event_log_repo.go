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

// FilterEventLogs retrieves a paginated list of event logs matching the given filter.
func (r *Repo) FilterEventLogs(ctx context.Context, f ds.EventLogsFilter) (logs []ds.EventLog, count int, err error) {
	_, span := r.tracer.Start(ctx, "FilterEventLogs")
	defer span.End()

	b := r.filter("event_logs l", "l").
		join("LEFT JOIN users u ON u.id = l.user_id").
		join("LEFT JOIN entities e ON e.id = l.entity_id").
		columns(`
			l.id as "id",
			l.user_id,
			l.type,
			l.entity_change_id,
			l.message,
			l.is_public,
			l.created_at,
			
			u.username as "user_username",

			e.type as "entity_type",
			e.title as "entity_title",
			e.public_id as "entity_public_id"
`).
		paginate(f.Page, f.PerPage).
		withCount(f.WithCount).
		order("l.created_at", "desc").
		withoutSoftDelete()

	if f.OnlyPublic {
		b.whereRaw("is_public IS TRUE")
	}

	count, err = b.scan(ctx, &logs)
	return
}
