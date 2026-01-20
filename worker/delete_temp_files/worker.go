// Package deletetempfiles ...
package deletetempfiles

import (
	"context"

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
	return "DELETE-TEMP-FILES"
}

// Schedule defines when the job should run.
// This job is scheduled to run once daily at midnight.
func (w Job) Schedule() gocron.JobDefinition {
	return gocron.DailyJob(1,
		gocron.NewAtTimes(gocron.NewAtTime(1, 0, 0)),
	)
}

// Do mark temp files as deleted after ds.DeleteTempFilesAfterDays.
func (w Job) Do(ctx context.Context, _ *service.Service, db *app.DB) (err error) {
	_, err = db.Exec(ctx, "UPDATE files SET deleted_at = NOW() WHERE temp IS TRUE AND created_at < NOW() - ($1 * INTERVAL '1 day') AND deleted_at IS NULL", ds.DeleteTempFilesAfterDays)

	return err
}
