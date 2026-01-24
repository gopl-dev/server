//nolint:mnd
package seed

import (
	"context"
	"fmt"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test/factory"
	"github.com/gopl-dev/server/test/factory/random"
	"golang.org/x/sync/errgroup"
)

// Users seeds the database with `count` random users.
//
// It generates unique usernames and emails, assigns randomized timestamps
// (created/updated/deleted), and inserts users concurrently using the factory.
func (s *Seed) Users(ctx context.Context, count int) (err error) {
	if count < 1 {
		return fmt.Errorf("seed users: %w", ErrInvalidCount)
	}

	uniqueUsername := func() (string, error) {
		return factory.LookupIUnique(ctx, s.db, "users", "username", fake.Username(), func(s string) string {
			return s + "." + random.String(5)
		})
	}

	uniqueEmail := func() (string, error) {
		return factory.LookupIUnique(ctx, s.db, "users", "email", fake.Email(), func(s string) string {
			return random.String(5) + "." + s
		})
	}

	var eg errgroup.Group

	for range count {
		eg.Go(func() error {
			email, err := uniqueEmail()
			if err != nil {
				return err
			}

			username, err := uniqueUsername()
			if err != nil {
				return err
			}

			createdAt := fake.DateRange(time.Now().AddDate(0, -12, 0), time.Now())
			updatedAt := random.ValOrNil(fake.DateRange(createdAt.AddDate(0, -12, -25), createdAt), 75)
			deletedAt := random.ValOrNil(fake.DateRange(createdAt.AddDate(0, -12, -25), createdAt), 75)

			_, err = s.factory.CreateUser(ds.User{
				Username:       username,
				Email:          email,
				EmailConfirmed: random.Bool(),
				CreatedAt:      createdAt,
				UpdatedAt:      updatedAt,
				DeletedAt:      deletedAt,
			})

			return err
		})
	}

	return eg.Wait()
}
