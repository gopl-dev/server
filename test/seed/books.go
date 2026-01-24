//nolint:mnd
package seed

import (
	"context"
	"fmt"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/file"
	"github.com/gopl-dev/server/test/factory"
	"golang.org/x/sync/errgroup"
)

// Books seeds the database with `count` random books.
//
// It generates a unique public slug (entities.public_id) for each book title,
// creates and persists a cover image file, then creates an entity + book record.
// Inserts are executed concurrently; the first error stops the seeding.
func (s *Seed) Books(ctx context.Context, count int) (err error) {
	if count < 1 {
		return fmt.Errorf("seed books: %w", ErrInvalidCount)
	}

	uniqueSlug := func(name string) (string, error) {
		return factory.LookupIUnique(ctx, s.db, "entities", "public_id", app.Slug(name), func(s string) string {
			return s + "-" + fake.UrlSlug(1)
		})
	}

	var eg errgroup.Group

	for range count {
		eg.Go(func() error {
			title := fake.BookTitle()

			slug, err := uniqueSlug(title)
			if err != nil {
				return err
			}

			ownerID, err := s.RandomUserID(ctx)
			if err != nil {
				return err
			}

			cover, err := makeImageFile(ctx, 800, 600)
			if err != nil {
				return err
			}

			cover.OwnerID = ownerID
			cover.Type = file.TypeImage
			cover.Purpose = ds.FilePurposeBookCover
			cover.CreatedAt = time.Now()
			cover.DeletedAt = nil
			cover.Temp = false

			cover, err = s.factory.CreateFile(*cover)
			if err != nil {
				return err
			}

			e := s.factory.NewEntity(ds.Entity{
				Title:         title,
				PublicID:      slug,
				OwnerID:       ownerID,
				PreviewFileID: cover.ID,
			})

			_, err = s.factory.CreateBook(ds.Book{
				Entity:      e,
				CoverFileID: cover.ID,
			})
			return err
		})
	}

	err = eg.Wait()
	if err != nil {
		return
	}

	// make sure one Book.PublicID is just "test",
	// so we can easily do manual tests
	_, err = s.db.Exec(ctx, "UPDATE entities SET public_id = 'test', status=$1, visibility=$2, deleted_at=NULL WHERE id=(SELECT id FROM entities ORDER BY RANDOM() LIMIT 1)", ds.EntityStatusApproved, ds.EntityVisibilityPublic)
	if err != nil {
		return err
	}

	return nil
}
