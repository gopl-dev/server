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

	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now()
	}

	query := `
		INSERT INTO user_activity_logs (user_id, action_type, is_public, entity_type, entity_id, meta, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	row := r.db.QueryRow(ctx, query,
		log.UserID, log.ActionType, log.IsPublic,
		log.EntityType, log.EntityID, log.Meta, log.CreatedAt,
	)
	err = row.Scan(&log.ID)

	return
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
	err := pgxscan.Get(ctx, r.db, log, query, userID, t)
	if noRows(err) {
		return nil, ErrActivityLogNotFound
	}

	return log, err
}

// UpdateUserActivityLogPublic updates a user activity log by its ID, setting it to public.
func (r *Repo) UpdateUserActivityLogPublic(ctx context.Context, id ds.ID) (err error) {
	_, span := r.tracer.Start(ctx, "UpdateUserActivityLogPublic")
	defer span.End()

	query := `
		UPDATE user_activity_logs
		SET is_public = true
		WHERE id = $1
	`

	_, err = r.db.Exec(ctx, query, id)
	return
}
