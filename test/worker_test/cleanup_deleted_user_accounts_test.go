package worker_test

import (
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory"
	"github.com/gopl-dev/server/worker/cleanup_deleted_users"
)

func TestCleanupDeletedUserAccounts(t *testing.T) {
	user := create(t, ds.User{
		DeletedAt: new(time.Now().Add(-(ds.CleanupDeletedUserAfter + time.Hour))),
	})

	_, err := factory.Five(tt.Factory.CreateUserSession, ds.UserSession{UserID: user.ID})
	test.CheckErr(t, err)

	_, err = factory.Five(tt.Factory.CreatePasswordResetToken, ds.PasswordResetToken{UserID: user.ID})
	test.CheckErr(t, err)

	_, err = factory.Five(tt.Factory.CreateEmailConfirmation, ds.EmailConfirmation{UserID: user.ID})
	test.CheckErr(t, err)

	_, err = factory.Five(tt.Factory.CreateChangeEmailRequest, ds.ChangeEmailRequest{UserID: user.ID})
	test.CheckErr(t, err)

	runJob(t, cleanupdeletedusers.NewJob())

	test.AssertNotInDB(t, tt.DB, "user_sessions", test.Data{"user_id": user.ID})
	test.AssertNotInDB(t, tt.DB, "password_reset_tokens", test.Data{"user_id": user.ID})
	test.AssertNotInDB(t, tt.DB, "email_confirmations", test.Data{"user_id": user.ID})
	test.AssertNotInDB(t, tt.DB, "change_email_requests", test.Data{"user_id": user.ID})
	test.AssertNotInDB(t, tt.DB, "users", test.Data{"id": user.ID})
}
