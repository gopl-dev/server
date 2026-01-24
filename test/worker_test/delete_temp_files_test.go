package worker_test

import (
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory"
	deletetempfiles "github.com/gopl-dev/server/worker/delete_temp_files"
)

func TestDeleteTempFiles(t *testing.T) {
	files, err := factory.Ten(tt.Factory.CreateFile, ds.File{
		Temp:      true,
		CreatedAt: time.Now().AddDate(0, 0, -(ds.DeleteTempFilesAfterDays + 1)),
	})
	test.CheckErr(t, err)

	runJob(t, deletetempfiles.NewJob())

	for _, f := range files {
		test.AssertDeleted(t, tt.DB, "files", f.ID)
	}
}
