//nolint:mnd
package seed

import (
	"context"
	"fmt"
	"strings"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/file"
	"github.com/gopl-dev/server/test/factory"
	"github.com/gopl-dev/server/test/factory/random"
	"golang.org/x/sync/errgroup"
)

var topicList = []string{
	"Beginner", "Advanced", "Algorithms",
	"Concurrency", "Parallelism", "Microservices",
	"API Design", "High Load", "Databases",
	"DevOps", "Security", "Testing",
	"Refactoring", "Debugging", "Software Engineering",
	"Theory", "Case Studies", "Reference",
}

// Books seeds the database with `count` random books.
//
// It generates a unique public slug (entities.public_id) for each book title,
// creates and persists a cover image file, then creates an entity + book record.
// Inserts are executed concurrently; the first error stops the seeding.
func (s *Seed) Books(ctx context.Context, count int) (err error) {
	if count < 1 {
		return fmt.Errorf("seed books: %w", ErrInvalidCount)
	}

	uniqueSlug := func(from string) (string, error) {
		return factory.LookupIUnique(ctx, s.db, "entities", "public_id", from, func(s string) string {
			return s + "-" + fake.UrlSlug(1)
		})
	}

	topics, err := s.resolveBookTopics(ctx)
	if err != nil {
		return err
	}

	var eg errgroup.Group

	for range count {
		eg.Go(func() error {
			title := random.Element([]string{fake.BookTitle(), fake.MovieName()})

			ownerID, err := s.RandomUserID(ctx)
			if err != nil {
				return err
			}

			cover, err := makeImageFile(ctx, 500, 800)
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
				Type:          ds.EntityTypeBook,
				Title:         title,
				Description:   strings.Join([]string{fake.Paragraph(), fake.Paragraph(), fake.Paragraph()}, " "),
				PublicID:      app.Slug(title),
				OwnerID:       ownerID,
				PreviewFileID: cover.ID,
				DeletedAt:     random.ValOrNil(fake.DateRange(time.Now().AddDate(0, -12, -25), time.Now()), 10),
			})

		createBook:
			_, err = s.factory.CreateBook(ds.Book{
				Entity:      e,
				CoverFileID: cover.ID,
			})
			if _, ok := isUniqueViolation(err); ok {
				newID, err := uniqueSlug(e.PublicID)
				if err != nil {
					return err
				}
				e.PublicID = newID
				goto createBook
			}

			// topics
			for range random.Int(1, 3) {
				t := random.Element(topics)
				err := s.repo.CreateEntityTopic(ctx, e.ID, t.ID)
				if _, ok := isUniqueViolation(err); ok {
					continue
				}
				if err != nil {
					return fmt.Errorf("create entity topic: %w", err)
				}
			}

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

func (s *Seed) resolveBookTopics(ctx context.Context) ([]ds.Topic, error) {
	topics, _, err := s.repo.FilterTopics(ctx, ds.TopicsFilter{
		PerPage: 100,
		Type:    ds.EntityTypeBook,
	})
	if err != nil {
		return nil, err
	}

	if len(topics) == 0 {
		topics = make([]ds.Topic, len(topicList))
		for i, name := range topicList {
			t, err := s.factory.CreateTopic(ds.Topic{
				ID:          ds.NewID(),
				Type:        ds.EntityTypeBook,
				PublicID:    app.Slug(name),
				Name:        name,
				Description: fake.ProductDescription(),
				CreatedAt:   time.Now(),
			})
			if err != nil {
				return nil, err
			}

			topics[i] = *t
		}

		return topics, nil
	}

	return topics, nil
}
