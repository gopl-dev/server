package repo

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

func (r *Repo) CreateUserSession(ctx context.Context, s *ds.UserSession) (err error) {
	_, err = r.db.Exec(ctx, `INSERT INTO user_sessions (id, user_id, created_at, expires_at) VALUES ($1, $2, $3, $4)`,
		s.ID, s.UserID, s.CreatedAt, s.ExpiresAt)

	return
}

func (r *Repo) FindUserSessionByID(ctx context.Context, id string) (sess *ds.UserSession, err error) {
	sess = &ds.UserSession{}
	err = pgxscan.Get(ctx, r.db, sess, `SELECT * FROM user_sessions WHERE id = $1`, id)
	if noRows(err) {
		sess = nil
		err = nil
	}

	return
}

func (r *Repo) ProlongUserSession(ctx context.Context, id uuid.UUID) (err error) {
	_, err = r.db.Exec(ctx, `UPDATE user_sessions SET expires_at = $1 WHERE id = $2`,
		time.Now().Add(time.Hour*time.Duration(app.Config().Session.DurationHours)), id)

	return
}

func (r *Repo) DeleteUserSession(ctx context.Context, id uuid.UUID) (err error) {
	_, err = r.db.Exec(ctx, `DELETE FROM user_sessions WHERE id = $1`, id)

	return
}
