package repo

import (
	"context"
	"time"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app/ds"
)

// FindEmailConfirmationByCode retrieves an email confirmation record from the database
// using its unique confirmation code.
// If a record is not found, it returns (nil, nil).
func (r *Repo) FindEmailConfirmationByCode(ctx context.Context, code string) (ec *ds.EmailConfirmation, err error) {
	_, span := r.tracer.Start(ctx, "FindEmailConfirmationByCode")
	defer span.End()

	ec = new(ds.EmailConfirmation)
	err = pgxscan.Get(ctx, r.getDB(ctx), ec,
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
	_, span := r.tracer.Start(ctx, "CreateEmailConfirmation")
	defer span.End()

	if ec.ID.IsNil() {
		ec.ID = ds.NewID()
	}

	if ec.CreatedAt.IsZero() {
		ec.CreatedAt = time.Now()
	}

	return r.insert(ctx, "email_confirmations", data{
		"id":         ec.ID,
		"user_id":    ec.UserID,
		"code":       ec.Code,
		"created_at": ec.CreatedAt,
		"expires_at": ec.ExpiresAt,
	})
}

// DeleteEmailConfirmation deletes an email confirmation record from the database
// using its ID.
func (r *Repo) DeleteEmailConfirmation(ctx context.Context, id ds.ID) (err error) {
	_, span := r.tracer.Start(ctx, "DeleteEmailConfirmation")
	defer span.End()

	return r.hardDelete(ctx, "email_confirmations", id)
}

// DeleteEmailConfirmationByUser deletes an email confirmations that belong to specific user.
func (r *Repo) DeleteEmailConfirmationByUser(ctx context.Context, userID ds.ID) (err error) {
	_, span := r.tracer.Start(ctx, "DeleteEmailConfirmationByUser")
	defer span.End()

	return r.exec(ctx, "DELETE FROM email_confirmations WHERE user_id = $1", userID)
}
