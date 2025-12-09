package repo

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app/ds"
)

// FindEmailConfirmationByCode retrieves an email confirmation record from the database
// using its unique confirmation code.
//
// If a record is not found, it returns (nil, nil).
func (r *Repo) FindEmailConfirmationByCode(ctx context.Context, code string) (ec *ds.EmailConfirmation, err error) {
	ec = new(ds.EmailConfirmation)

	err = pgxscan.Get(ctx, r.db, ec,
		"SELECT * FROM email_confirmations WHERE code = $1",
		code,
	)
	if noRows(err) {
		ec = nil
		err = nil
	}

	return ec, err
}

// CreateEmailConfirmation creates a new email confirmation record in the database.
func (r *Repo) CreateEmailConfirmation(ctx context.Context, ec *ds.EmailConfirmation) (err error) {
	row := r.db.QueryRow(ctx,
		"INSERT INTO email_confirmations (user_id, code, created_at, expires_at) VALUES ($1, $2, $3, $4) RETURNING id",
		ec.UserID, ec.Code, ec.CreatedAt, ec.ExpiresAt,
	)
	err = row.Scan(&ec.ID)

	return
}

// DeleteEmailConfirmation deletes an email confirmation record from the database
// using its ID.
func (r *Repo) DeleteEmailConfirmation(ctx context.Context, id int64) (err error) {
	_, err = r.db.Exec(ctx, "DELETE FROM email_confirmations WHERE id = $1", id)

	return
}
