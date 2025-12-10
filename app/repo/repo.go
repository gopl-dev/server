// Package repo ...
package repo

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"
)

// Repo is the primary struct for database access operations
// All repository methods are attached to this type.
type Repo struct {
	db     *pgxpool.Pool
	tracer trace.Tracer
}

// New is a factory function that creates and returns a new Repo instance.
func New(db *pgxpool.Pool, t trace.Tracer) *Repo {
	return &Repo{
		db:     db,
		tracer: t,
	}
}

func noRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
