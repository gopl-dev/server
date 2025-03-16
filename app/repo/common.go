package repo

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

func noRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}
