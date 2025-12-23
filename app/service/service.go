// Package service ...
package service

import (
	"fmt"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/repo"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"
)

const (
	// UsernameMinLen defines the minimum allowed length, in characters, for a user's username.
	UsernameMinLen = 2

	// UsernameMaxLen defines the maximum allowed length, in characters, for a user's username.
	UsernameMaxLen = 30

	// UserPasswordMinLen defines the minimum allowed length, in characters, for a user's password.
	UserPasswordMinLen = 6
)

var (
	newPasswordInputRules = z.String().Min(UserPasswordMinLen,
		z.Message("Password must be at least 6 characters")).
		Required(z.Message("Password is required"))
	userIDInputRules = z.Int64().Required(z.Message("userID is required"))
	emailInputRules  = z.String().Email().Required(z.Message("Email is required"))
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

func validateInput(rules z.Shape, data any) (err error) {
	// Zod panics if struct is missing rules key
	// we don't want that
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
				return
			}

			err = fmt.Errorf("%v", r) //nolint:err113
			return
		}
	}()

	issueMap := z.Struct(rules).Validate(data)
	if len(issueMap) == 0 {
		return nil
	}

	ie := app.NewInputError()
	for key, issues := range issueMap {
		messages := make([]string, 0, len(issues))
		for _, issue := range issues {
			if issue == nil {
				continue
			}

			messages = append(messages, issue.Message)
		}

		// TODO: Zod performs its own name conversion (e.g., OldPassword => Old_password).
		// We need snake_case. See if there's a better way to handle this than converting it manually here.
		key = app.CamelCaseToSnakeCase(key)
		ie.Add(key, strings.Join(messages, "\n"))
	}

	return ie
}
