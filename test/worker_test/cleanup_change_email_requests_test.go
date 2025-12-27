package worker_test

import (
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory"
	"github.com/gopl-dev/server/worker/cleanup_change_email_requests"
)

func TestCleanupChangeEmailRequests(t *testing.T) {
	user := tt.Factory.CreateUser(t)
	factory.Ten(t, tt.Factory.CreateChangeEmailRequest, ds.ChangeEmailRequest{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(-time.Hour),
	})

	runJob(t, cleanupchangeemailrequests.Job{})

	test.AssertNotInDB(t, tt.DB, "change_email_requests", test.Data{
		"user_id": user.ID,
	})

	req := tt.Factory.CreateChangeEmailRequest(t)

	runJob(t, cleanupchangeemailrequests.Job{})

	test.AssertInDB(t, tt.DB, "change_email_requests", test.Data{"id": req.ID})
}
