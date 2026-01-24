package worker_test

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/file"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory/random"
	cleanupfiles "github.com/gopl-dev/server/worker/cleanup_files"
)

func TestCleanupFiles(t *testing.T) {
	// make sample data of 10 deleted files
	deletedAt := time.Now().AddDate(0, 0, -(ds.CleanupDeletedFilesAfterDays + 1))
	ctx := context.Background()

	storedFiles := make([]*ds.File, 10)

	for i := range storedFiles {
		var previewPath string
		if random.Bool() {
			previewPath = random.String()
		}

		fileBytes, err := random.ImagePNG()
		test.CheckErr(t, err)

		filePath, err := file.Store(ctx, bytes.NewReader(fileBytes), random.String())
		test.CheckErr(t, err)

		if previewPath != "" {
			previewBytes, err := random.ImagePNG()
			test.CheckErr(t, err)

			previewPath, err = file.Store(ctx, bytes.NewReader(previewBytes), filepath.Join("preview", filePath))
			test.CheckErr(t, err)
		}

		f := create(t, ds.File{
			Path:        filePath,
			PreviewPath: previewPath,
			DeletedAt:   &deletedAt,
		})

		storedFiles[i] = f
	}

	// run job
	runJob(t, cleanupfiles.NewJob())

	// check if those files are gone
	for _, f := range storedFiles {
		test.AssertNotInDB(t, tt.DB, "files", test.Data{"id": f.ID})
		_, _, err := file.Open(ctx, f.Path)
		if err == nil {
			t.Fatal("expected error")
		}
		if !errors.Is(err, file.ErrFileNotFound) {
			t.Fatalf("expected ErrFileNotFound, got %v", err)
		}
	}
}
