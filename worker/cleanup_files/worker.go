// Package cleanupfiles ...
package cleanupfiles

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
)

// Job implements the worker.Job interface for cleaning up deleted user accounts.
type Job struct{}

// NewJob ...
func NewJob() *Job {
	return &Job{}
}

// Name returns the unique name of the job.
func (w Job) Name() string {
	return "CLEANUP-FILES"
}

// Schedule defines when the job should run.
// This job is scheduled to run once daily at midnight.
func (w Job) Schedule() gocron.JobDefinition {
	return gocron.DailyJob(1,
		gocron.NewAtTimes(gocron.NewAtTime(2, 0, 0)), //nolint:mnd
	)
}

// Do mark temp files as deleted after ds.DeleteTempFilesAfterDays.
func (w Job) Do(ctx context.Context, s *service.Service, db *app.DB) (err error) {
	batchSize := 100

	// To delete a file, a user must be present in the context
	// and must be either an admin or the file owner.
	user := ds.User{IsAdmin: true}
	ctx = user.ToContext(ctx)

processBatch:
	files, count, err := s.FilterFiles(ctx, ds.FilesFilter{
		DeletedAt: ds.DtBefore(time.Now().AddDate(0, 0, -ds.CleanupDeletedFilesAfterDays)),
		PerPage:   batchSize,
		WithCount: true,
	})
	if err != nil {
		return
	}

	println("[CLEANUP-FILES]:", count, "files about to be removed from system")

	for _, f := range files {
		// detach from entities
		_, err = db.Exec(ctx, "UPDATE entities SET preview_file_id=NULL WHERE preview_file_id=$1", f.ID)
		if err != nil {
			err = fmt.Errorf("detach preview file from entities: %w", err)
			return
		}
		// detach book covers
		_, err = db.Exec(ctx, "UPDATE books SET cover_file_id=NULL WHERE cover_file_id=$1", f.ID)
		if err != nil {
			err = fmt.Errorf("detach cover file from books: %w", err)
			return
		}

		err := s.HardDeleteFileUnsafe(ctx, &f)
		if err != nil {
			return err
		}
	}

	if count > batchSize {
		goto processBatch
	}

	return nil
}
