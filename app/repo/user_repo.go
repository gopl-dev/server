package repo

import (
	"context"
	"errors"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app/ds"
)

var (
	// ErrUserNotFound is returned when lookup method fails to find a user.
	ErrUserNotFound = errors.New("user not found")
)

// FindUserByEmail retrieves a user from the database by their email address.
func (r *Repo) FindUserByEmail(ctx context.Context, email string) (*ds.User, error) {
	_, span := r.tracer.Start(ctx, "FindUserByEmail")
	defer span.End()

	user := new(ds.User)
	err := pgxscan.Get(ctx, r.db, user, `SELECT * FROM users WHERE email = $1`, email)
	if noRows(err) {
		return nil, ErrUserNotFound
	}

	return user, err
}

// FindUserByUsername retrieves a user from the database by their username.
func (r *Repo) FindUserByUsername(ctx context.Context, username string) (*ds.User, error) {
	_, span := r.tracer.Start(ctx, "FindUserByUsername")
	defer span.End()

	user := new(ds.User)
	err := pgxscan.Get(ctx, r.db, user, `SELECT * FROM users WHERE username = $1`, username)
	if noRows(err) {
		return nil, ErrUserNotFound
	}

	return user, err
}

// FindUserByID retrieves a user from the database by their ID.
func (r *Repo) FindUserByID(ctx context.Context, id int64) (*ds.User, error) {
	_, span := r.tracer.Start(ctx, "FindUserByID")
	defer span.End()

	user := new(ds.User)
	err := pgxscan.Get(ctx, r.db, user, `SELECT * FROM users WHERE id = $1`, id)
	if noRows(err) {
		user = nil
		err = ErrUserNotFound
	}

	return user, err
}

// CreateUser inserts a new user record into the database.
func (r *Repo) CreateUser(ctx context.Context, user *ds.User) (err error) {
	_, span := r.tracer.Start(ctx, "CreateUser")
	defer span.End()

	row := r.db.QueryRow(ctx,
		"INSERT INTO users (username, email, email_confirmed, password, created_at) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		user.Username, user.Email, user.EmailConfirmed, user.Password, user.CreatedAt)
	err = row.Scan(&user.ID)

	return
}

// SetUserEmailConfirmed updates a user's record to set the email_confirmed flag to true
// and updates the updated_at timestamp.
func (r *Repo) SetUserEmailConfirmed(ctx context.Context, userID int64) (err error) {
	_, span := r.tracer.Start(ctx, "SetUserEmailConfirmed")
	defer span.End()

	_, err = r.db.Exec(ctx, "UPDATE users SET email_confirmed = true, updated_at = NOW() WHERE id = $1", userID)
	return
}
