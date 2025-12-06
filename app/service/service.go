package service

import (
	"github.com/gopl-dev/server/app/repo"
	"github.com/jackc/pgx/v5/pgxpool"
)

type contextKey string

type Service struct {
	db *repo.Repo
}

func New(db *pgxpool.Pool) *Service {
	return &Service{
		db: repo.New(db),
	}
}
