package service

import (
	"context"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

func CreateContentImportLog(ctx context.Context, log *ds.ContentImportLog) (err error) {
	_, err = app.DB().Exec(ctx,
		"INSERT INTO content_import_logs (status, log, created_at) VALUES ($1, $2, $3)",
		log.Status,
		log.Log,
		log.CreatedAt,
	)

	return err
}
