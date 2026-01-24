package worker_test

import (
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory"
	cleanupchangeemailrequests "github.com/gopl-dev/server/worker/cleanup_change_email_requests"
	cleanupexpiredusersessions "github.com/gopl-dev/server/worker/cleanup_expired_user_sessions"
)

func TestCleanupExpiredUserSessions(t *testing.T) {
	user := create[ds.User](t)

	_, err := factory.Ten(tt.Factory.CreateUserSession, ds.UserSession{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(-time.Hour),
	})
	test.CheckErr(t, err)

	runJob(t, cleanupexpiredusersessions.NewJob())

	test.AssertNotInDB(t, tt.DB, "user_sessions", test.Data{
		"user_id": user.ID,
	})

	// this one should not be deleted
	session := create[ds.UserSession](t)

	runJob(t, cleanupchangeemailrequests.NewJob())

	test.AssertInDB(t, tt.DB, "user_sessions", test.Data{"id": session.ID})
}
