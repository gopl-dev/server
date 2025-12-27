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
	user := tt.Factory.CreateUser(t)
	factory.Ten(t, tt.Factory.CreateUserSession, ds.UserSession{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(-time.Hour),
	})

	runJob(t, cleanupexpiredusersessions.Job{})

	test.AssertNotInDB(t, tt.DB, "user_sessions", test.Data{
		"user_id": user.ID,
	})

	// this one should not be deleted
	session := tt.Factory.CreateUserSession(t)

	runJob(t, cleanupchangeemailrequests.Job{})

	test.AssertInDB(t, tt.DB, "user_sessions", test.Data{"id": session.ID})
}
