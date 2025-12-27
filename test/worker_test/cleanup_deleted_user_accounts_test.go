package worker_test

import (
	"testing"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory"
	"github.com/gopl-dev/server/worker/cleanup_deleted_users"
)

func TestCleanupDeletedUserAccounts(t *testing.T) {
	user := tt.Factory.CreateUser(t, ds.User{
		DeletedAt: app.Pointer(time.Now().Add(-(ds.CleanupDeletedUserAfter + time.Hour))),
	})

	factory.Five(t, tt.Factory.CreateUserSession, ds.UserSession{UserID: user.ID})
	factory.Five(t, tt.Factory.CreatePasswordResetToken, ds.PasswordResetToken{UserID: user.ID})
	factory.Five(t, tt.Factory.CreateEmailConfirmation, ds.EmailConfirmation{UserID: user.ID})
	factory.Five(t, tt.Factory.CreateChangeEmailRequest, ds.ChangeEmailRequest{UserID: user.ID})

	runJob(t, cleanupdeletedusers.Job{})

	test.AssertNotInDB(t, tt.DB, "user_sessions", test.Data{"user_id": user.ID})
	test.AssertNotInDB(t, tt.DB, "password_reset_tokens", test.Data{"user_id": user.ID})
	test.AssertNotInDB(t, tt.DB, "email_confirmations", test.Data{"user_id": user.ID})
	test.AssertNotInDB(t, tt.DB, "change_email_requests", test.Data{"user_id": user.ID})
	test.AssertNotInDB(t, tt.DB, "users", test.Data{"id": user.ID})
}
