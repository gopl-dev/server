package worker_test

import (
	"testing"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory"
	cleanupdeletedbooks "github.com/gopl-dev/server/worker/cleanup_deleted_books"
)

func TestCleanupDeletedBooks(t *testing.T) {
	topic := create[ds.Topic](t)
	file := create[ds.File](t)
	legitBook := create[ds.Book](t)
	bookWithFile := create[ds.Book](t, ds.Book{
		CoverFileID: file.ID,
		Entity: &ds.Entity{
			DeletedAt: app.Pointer(time.Now().AddDate(0, 0, -(ds.CleanupDeletedEntitiesAfterDays + 1))),
		},
	})
	books, err := factory.Ten(tt.Factory.CreateBook, ds.Book{
		Entity: &ds.Entity{
			DeletedAt: app.Pointer(time.Now().AddDate(0, 0, -(ds.CleanupDeletedEntitiesAfterDays + 1))),
			Topics:    []ds.Topic{*topic},
		},
	})
	test.CheckErr(t, err)

	for _, book := range books {
		create(t, ds.EntityChangeRequest{
			EntityID: book.ID,
		})
	}

	// run job
	runJob(t, cleanupdeletedbooks.NewJob())

	// legit book should not be deleted
	test.AssertInDB(t, tt.DB, "books", test.Data{
		"id": legitBook.ID,
	})

	// everything else should
	test.AssertInDB(t, tt.DB, "files", test.Data{
		"id":         file.ID,
		"deleted_at": test.NotNull,
	})
	test.AssertNotInDB(t, tt.DB, "books", test.Data{
		"id": bookWithFile.ID,
	})
	test.AssertNotInDB(t, tt.DB, "entities", test.Data{
		"id": bookWithFile.ID,
	})

	for _, b := range books {
		test.AssertNotInDB(t, tt.DB, "books", test.Data{
			"id": b.ID,
		})
		test.AssertNotInDB(t, tt.DB, "entities", test.Data{
			"id": b.ID,
		})
		test.AssertNotInDB(t, tt.DB, "entity_change_requests", test.Data{
			"entity_id": b.ID,
		})
		test.AssertNotInDB(t, tt.DB, "entity_topics", test.Data{
			"entity_id": b.ID,
		})
	}
}
