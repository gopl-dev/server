package repo

import (
	"context"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

func FindEmailConfirmationByCode(ctx context.Context, code string) (ec *ds.EmailConfirmation, err error) {
	ec = &ds.EmailConfirmation{}
	err = pgxscan.Get(ctx, app.DB(), ec,
		"SELECT * FROM email_confirmations WHERE code = $1",
		code,
	)
	if noRows(err) {
		ec = nil
		err = nil
	}

	return ec, err
}

func CreateEmailConfirmation(ctx context.Context, ec *ds.EmailConfirmation) (err error) {
	r := app.DB().QueryRow(ctx,
		"INSERT INTO email_confirmations (user_id, code, created_at, expires_at) VALUES ($1, $2, $3, $4) RETURNING id",
		ec.UserID, ec.Code, ec.CreatedAt, ec.ExpiresAt,
	)
	err = r.Scan(&ec.ID)

	return
}
