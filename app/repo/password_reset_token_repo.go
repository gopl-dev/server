package repo

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app/ds"
)

// CreatePasswordResetToken inserts a new password reset token into the database.
func (r *Repo) CreatePasswordResetToken(ctx context.Context, token *ds.PasswordResetToken) error {
	_, span := r.tracer.Start(ctx, "CreatePasswordResetToken")
	defer span.End()

	query := `
		INSERT INTO password_reset_tokens (user_id, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt).Scan(&token.ID)
}

// FindPasswordResetToken retrieves a password reset token from the database by the token string.
// If the token is not found, it returns ErrPasswordResetTokenNotFound.
func (r *Repo) FindPasswordResetToken(ctx context.Context, token string) (*ds.PasswordResetToken, error) {
	_, span := r.tracer.Start(ctx, "FindPasswordResetToken")
	defer span.End()

	prt := new(ds.PasswordResetToken)
	err := pgxscan.Get(ctx, r.db, prt, `SELECT * FROM password_reset_tokens WHERE token = $1`, token)
	if noRows(err) {
		return nil, ErrPasswordResetTokenNotFound
	}
	return prt, err
}

// DeletePasswordResetToken removes a password reset token from the database by its ID.
func (r *Repo) DeletePasswordResetToken(ctx context.Context, id int64) error {
	_, span := r.tracer.Start(ctx, "DeletePasswordResetToken")
	defer span.End()

	_, err := r.db.Exec(ctx, `DELETE FROM password_reset_tokens WHERE id = $1`, id)
	return err
}
