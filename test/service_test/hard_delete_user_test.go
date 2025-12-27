package service_test

import (
	"context"
	"testing"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory"
)

func TestHardDeleteUser(t *testing.T) {
	user := tt.Factory.CreateUser(t)
	// user sessions
	factory.Five(t, tt.Factory.CreateUserSession, ds.UserSession{UserID: user.ID})
	// email confirmations
	factory.Five(t, tt.Factory.CreateEmailConfirmation, ds.EmailConfirmation{UserID: user.ID})
	// password resets
	factory.Five(t, tt.Factory.CreatePasswordResetToken, ds.PasswordResetToken{UserID: user.ID})
	// change email requests
	factory.Five(t, tt.Factory.CreateChangeEmailRequest, ds.ChangeEmailRequest{UserID: user.ID})

	err := tt.Service.HardDeleteUser(context.Background(), user.ID)
	test.CheckErr(t, err)

	test.AssertNotInDB(t, tt.DB, "user_sessions", test.Data{"user_id": user.ID})
	test.AssertNotInDB(t, tt.DB, "email_confirmations", test.Data{"user_id": user.ID})
	test.AssertNotInDB(t, tt.DB, "password_reset_tokens", test.Data{"user_id": user.ID})
	test.AssertNotInDB(t, tt.DB, "change_email_requests", test.Data{"user_id": user.ID})
	test.AssertNotInDB(t, tt.DB, "users", test.Data{"id": user.ID})
}
