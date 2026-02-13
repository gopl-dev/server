// Package cleanupdeletedbooks provides a worker job for permanently removing soft-deleted books
// from the database after a retention period defined by ds.CleanupDeletedEntitiesAfterDays.
package cleanupdeletedbooks

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
)

// Job implements the worker.Job interface for cleaning up expired user sessions.
type Job struct{}

// NewJob ...
func NewJob() *Job {
	return &Job{}
}

// Name returns the unique name of the job.
func (w Job) Name() string {
	return "CLEANUP:DELETED_BOOKS"
}

// Schedule defines when the job should run.
// This job is scheduled to run once daily at midnight.
func (w Job) Schedule() gocron.JobDefinition {
	return gocron.DailyJob(1,
		gocron.NewAtTimes(gocron.NewAtTime(0, 0, 0)),
	)
}

// Do executes the job's task, which is to all soft-deleted ds.Books and relations.
func (w Job) Do(ctx context.Context, s *service.Service, db *app.DB) (err error) {
	batchSize := 100

processBatch:
	books, count, err := s.FilterBooks(ctx, ds.BooksFilter{
		EntitiesFilter: ds.EntitiesFilter{
			DeletedAt: ds.DtBefore(time.Now().AddDate(0, 0, -ds.CleanupDeletedEntitiesAfterDays)),
			PerPage:   batchSize,
			WithCount: true,
		},
	})
	if err != nil {
		return
	}
	if count == 0 {
		return nil
	}

	println("[CLEANUP-BOOKS]:", count, "books about to be permanently removed from system")

	ids := make([]ds.ID, len(books))
	fileIDs := make([]ds.ID, 0)

	for i, b := range books {
		ids[i] = b.ID

		if !b.CoverFileID.IsNil() {
			fileIDs = append(fileIDs, b.CoverFileID)
		}
		if !b.PreviewFileID.IsNil() {
			fileIDs = append(fileIDs, b.PreviewFileID)
		}
	}

	_, err = db.Exec(ctx, "DELETE FROM entity_topics WHERE entity_id =  ANY($1)", ids)
	if err != nil {
		err = fmt.Errorf("detach entity topics: %w", err)
		return
	}

	// Another worker will properly take of files, keep simple here
	_, err = db.Exec(ctx, "UPDATE files SET deleted_at=NOW() where id = ANY($1)", fileIDs)
	if err != nil {
		err = fmt.Errorf("set files deleted: %w", err)
		return
	}

	_, err = db.Exec(ctx, "DELETE FROM entity_change_requests WHERE entity_id = ANY($1)", ids)
	if err != nil {
		err = fmt.Errorf("delete books: %w", err)
		return
	}

	_, err = db.Exec(ctx, "DELETE FROM books WHERE id = ANY($1)", ids)
	if err != nil {
		err = fmt.Errorf("delete books: %w", err)
		return
	}

	_, err = db.Exec(ctx, "DELETE FROM entities WHERE id = ANY($1)", ids)
	if err != nil {
		err = fmt.Errorf("delete entities of books: %w", err)
		return
	}

	if count > batchSize {
		goto processBatch
	}

	return nil
}
