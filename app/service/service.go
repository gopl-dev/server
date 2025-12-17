package service

import (
	"github.com/gopl-dev/server/app/repo"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"
)

// Service holds dependencies required for the application's business logic layer.
type Service struct {
	db     *repo.Repo
	tracer trace.Tracer
}

// New is a factory function that creates and returns a new Service instance.
func New(db *pgxpool.Pool, t trace.Tracer) *Service {
	return &Service{
		db:     repo.New(db, t),
		tracer: t,
	}
}
