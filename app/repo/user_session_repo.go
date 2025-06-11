package repo

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/google/uuid"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

func CreateUserSession(ctx context.Context, s *ds.UserSession) (err error) {
	_, err = app.DB().Exec(ctx, `INSERT INTO user_sessions (id, user_id, created_at, expires_at) VALUES ($1, $2, $3, $4)`,
		s.ID, s.UserID, s.CreatedAt, s.ExpiresAt)

	return
}

func FindUserSessionByID(ctx context.Context, id string) (sess *ds.UserSession, err error) {
	sess = &ds.UserSession{}
	err = pgxscan.Get(ctx, app.DB(), sess, `SELECT * FROM user_sessions WHERE id = $1`, id)
	if noRows(err) {
		sess = nil
		err = nil
	}

	return
}

func ProlongUserSession(ctx context.Context, id uuid.UUID) (err error) {
	_, err = app.DB().Exec(ctx, `UPDATE user_sessions SET expires_at = $1 WHERE id = $2`,
		time.Now().Add(time.Hour*time.Duration(app.Config().Session.DurationHours)), id)

	return
}

func DeleteUserSession(ctx context.Context, id uuid.UUID) (err error) {
	_, err = app.DB().Exec(ctx, `DELETE FROM user_sessions WHERE id = $1`, id)

	return
}
