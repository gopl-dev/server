package repo

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app/ds"
)

// CreateChangeEmailRequest inserts a new change email request into the database.
func (r *Repo) CreateChangeEmailRequest(ctx context.Context, req *ds.ChangeEmailRequest) error {
	_, span := r.tracer.Start(ctx, "CreateChangeEmailRequest")
	defer span.End()

	query := `
		INSERT INTO change_email_requests (user_id, new_email, token, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	return r.db.QueryRow(ctx, query, req.UserID, req.NewEmail, req.Token, req.ExpiresAt, req.CreatedAt).Scan(&req.ID)
}

// FindChangeEmailRequestByToken retrieves a change email request from the database by its token.
// If the token is not found, it returns ErrChangeEmailRequestNotFound.
func (r *Repo) FindChangeEmailRequestByToken(ctx context.Context, token string) (*ds.ChangeEmailRequest, error) {
	_, span := r.tracer.Start(ctx, "FindChangeEmailRequestByToken")
	defer span.End()

	req := new(ds.ChangeEmailRequest)
	err := pgxscan.Get(ctx, r.db, req, `SELECT * FROM change_email_requests WHERE token = $1`, token)
	if noRows(err) {
		return nil, ErrChangeEmailRequestNotFound
	}
	return req, err
}

// DeleteChangeEmailRequest removes a change email request from the database by its ID.
func (r *Repo) DeleteChangeEmailRequest(ctx context.Context, id ds.ID) error {
	_, span := r.tracer.Start(ctx, "DeleteChangeEmailRequest")
	defer span.End()

	_, err := r.db.Exec(ctx, `DELETE FROM change_email_requests WHERE id = $1`, id)
	return err
}

// DeleteChangeEmailRequestsByUser removes a change email request for specific user.
func (r *Repo) DeleteChangeEmailRequestsByUser(ctx context.Context, userID ds.ID) error {
	_, span := r.tracer.Start(ctx, "DeleteChangeEmailRequest")
	defer span.End()

	_, err := r.db.Exec(ctx, `DELETE FROM change_email_requests WHERE user_id = $1`, userID)
	return err
}
