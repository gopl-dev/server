// Package repo ...
package repo

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repo is the primary struct for database access operations
// All repository methods are attached to this type.
type Repo struct {
	db *pgxpool.Pool
}

// New is a factory function that creates and returns a new Repo instance.
func New(db *pgxpool.Pool) *Repo {
	return &Repo{
		db: db,
	}
}

func noRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
