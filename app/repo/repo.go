package repo

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repo struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repo {
	return &Repo{
		db: db,
	}
}

func noRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
