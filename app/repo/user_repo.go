package repo

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

// FindUserByEmail retrieves a user from the database by their email address.
// If no user is found, it returns (nil, nil).
func FindUserByEmail(ctx context.Context, email string) (user *ds.User, err error) {
	user = &ds.User{}
	err = pgxscan.Get(ctx, app.DB(), user, `SELECT * FROM users WHERE email = $1`, email)
	if noRows(err) {
		user = nil
		err = nil
	}

	return user, err
}

// FindUserByID retrieves a user from the database by their ID.
func FindUserByID(ctx context.Context, id int64) (user *ds.User, err error) {
	user = &ds.User{}
	err = pgxscan.Get(ctx, app.DB(), user, `SELECT * FROM users WHERE id = $1`, id)
	if noRows(err) {
		user = nil
		err = ErrUserNotFound
	}

	return user, err
}

// FindUserByUsername retrieves a user from the database by their username.
// If no user is found, it returns (nil, nil).
func FindUserByUsername(ctx context.Context, username string) (user *ds.User, err error) {
	user = &ds.User{}
	err = pgxscan.Get(ctx, app.DB(), user, `SELECT * FROM users WHERE username = $1`, username)
	if noRows(err) {
		user = nil
		err = nil
	}

	return user, err
}

// CreateUser inserts a new user record into the database.
// It populates the newly created user's ID upon success.
func CreateUser(ctx context.Context, user *ds.User) (err error) {
	r := app.DB().QueryRow(ctx, "INSERT INTO users (username, email, email_confirmed, password, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		user.Username, user.Email, user.EmailConfirmed, user.Password, user.CreatedAt)
	err = r.Scan(&user.ID)

	return
}

// SetUserEmailConfirmed sets the email_confirmed flag for a user.
func SetUserEmailConfirmed(ctx context.Context, userID int64) (err error) {
	_, err = app.DB().Exec(ctx, "UPDATE users SET email_confirmed = true, updated_at=NOW() WHERE id = $1", userID)

	return
}
