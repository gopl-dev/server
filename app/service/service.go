package service

import (
	"github.com/gopl-dev/server/app/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

// Service holds dependencies required for the application's business logic layer.
type Service struct {
	db *repo.Repo
}

// New is a factory function that creates and returns a new Service instance.
func New(db *pgxpool.Pool) *Service {
	return &Service{
		db: repo.New(db),
	}
}
