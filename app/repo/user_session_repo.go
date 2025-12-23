package repo

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

// CreateUserSession inserts a new user session record into the database.
func (r *Repo) CreateUserSession(ctx context.Context, s *ds.UserSession) (err error) {
	_, span := r.tracer.Start(ctx, "CreateUserSession")
	defer span.End()

	_, err = r.db.Exec(ctx, `INSERT INTO user_sessions (id, user_id, created_at, expires_at) VALUES ($1, $2, $3, $4)`,
		s.ID, s.UserID, s.CreatedAt, s.ExpiresAt)

	return
}

// FindUserSessionByID retrieves a user session record from the database using its unique ID.
func (r *Repo) FindUserSessionByID(ctx context.Context, id uuid.UUID) (sess *ds.UserSession, err error) {
	_, span := r.tracer.Start(ctx, "FindUserSessionByID")
	defer span.End()

	sess = new(ds.UserSession)
	err = pgxscan.Get(ctx, r.db, sess, `SELECT * FROM user_sessions WHERE id = $1`, id)
	if noRows(err) {
		sess = nil
		err = nil
	}

	return
}

// ProlongUserSession updates the expiration timestamp of an existing user session.
func (r *Repo) ProlongUserSession(ctx context.Context, id uuid.UUID) (err error) {
	_, span := r.tracer.Start(ctx, "ProlongUserSession")
	defer span.End()

	_, err = r.db.Exec(ctx, `UPDATE user_sessions SET expires_at = $1 WHERE id = $2`,
		time.Now().Add(time.Hour*time.Duration(app.Config().Session.DurationHours)), id)

	return
}

// DeleteUserSession removes a user session record from the database using its unique ID.
func (r *Repo) DeleteUserSession(ctx context.Context, id uuid.UUID) (err error) {
	_, span := r.tracer.Start(ctx, "DeleteUserSession")
	defer span.End()

	_, err = r.db.Exec(ctx, `DELETE FROM user_sessions WHERE id = $1`, id)
	return
}
