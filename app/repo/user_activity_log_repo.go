package repo

import (
	"context"
	"errors"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app/ds"
	useractivity "github.com/gopl-dev/server/app/ds/user_activity"
)

// ErrActivityLogNotFound is returned when lookup method fails to find a row.
var ErrActivityLogNotFound = errors.New("user activity log not found")

// CreateUserActivityLog inserts a new user activity log record into the database.
func (r *Repo) CreateUserActivityLog(ctx context.Context, log *ds.UserActivityLog) (err error) {
	_, span := r.tracer.Start(ctx, "CreateUserActivityLog")
	defer span.End()

	if log.ID.IsNil() {
		log.ID = ds.NewID()
	}

	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	return r.insert(ctx, "user_activity_logs", data{
		"id":          log.ID,
		"user_id":     log.UserID,
		"action_type": log.ActionType,
		"is_public":   log.IsPublic,
		"entity_type": log.EntityType,
		"entity_id":   log.EntityID,
		"meta":        log.Meta,
		"created_at":  log.CreatedAt,
	})
}

// FindUserActivityLogByUserAndType finds the latest user activity log for a given user and action type.
func (r *Repo) FindUserActivityLogByUserAndType(
	ctx context.Context, userID ds.ID, t useractivity.Type) (*ds.UserActivityLog, error) {
	_, span := r.tracer.Start(ctx, "FindUserActivityLogByUserAndType")
	defer span.End()

	log := new(ds.UserActivityLog)
	query := `
		SELECT *
		FROM user_activity_logs
		WHERE user_id = $1 AND action_type = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	err := pgxscan.Get(ctx, r.getDB(ctx), log, query, userID, t)
	if noRows(err) {
		return nil, ErrActivityLogNotFound
	}

	return log, err
}

// UpdateUserActivityLogPublic updates a user activity log by its ID, setting it to public.
func (r *Repo) UpdateUserActivityLogPublic(ctx context.Context, id ds.ID) (err error) {
	_, span := r.tracer.Start(ctx, "UpdateUserActivityLogPublic")
	defer span.End()

	return r.exec(ctx, "UPDATE user_activity_logs SET is_public = true WHERE id = $1", id)
}
