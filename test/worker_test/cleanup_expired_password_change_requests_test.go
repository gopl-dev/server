package worker_test

import (
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory"
	cleanupexpiredpasswordchangerequests "github.com/gopl-dev/server/worker/cleanup_expired_password_change_requests"
)

func TestCleanupExpiredPasswordChangeRequests(t *testing.T) {
	user := tt.Factory.CreateUser(t)
	factory.Ten(t, tt.Factory.CreatePasswordResetToken, ds.PasswordResetToken{
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, -1),
	})

	runJob(t, cleanupexpiredpasswordchangerequests.Job{})

	test.AssertNotInDB(t, tt.DB, "password_reset_tokens", test.Data{
		"user_id": user.ID,
	})

	token := tt.Factory.CreatePasswordResetToken(t)
	runJob(t, cleanupexpiredpasswordchangerequests.Job{})

	test.AssertInDB(t, tt.DB, "password_reset_tokens", test.Data{
		"id": token.ID,
	})
}
