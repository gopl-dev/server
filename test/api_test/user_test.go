package api_test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	useractivity "github.com/gopl-dev/server/app/ds/user_activity"
	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/server/handler"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory/random"
)

func TestUserSignUp(t *testing.T) {
	req := request.UserSignUp{
		Username: random.String(),
		Email:    random.Email(),
		Password: random.String(),
	}

	var resp response.Status
	POST(t, Request{
		path:         "/users/sign-up",
		body:         req,
		bindResponse: &resp,
		assertStatus: http.StatusOK,
	})

	test.AssertInDB(t, tt.DB, "users", test.Data{
		"username":        req.Username,
		"email":           req.Email,
		"email_confirmed": false,
	})

	vars := test.LoadEmailVars(t, req.Email)

	assert.Equal(t, req.Username, app.String(vars["username"]))
	assert.Equal(t, req.Email, app.String(vars["email"]))

	user, err := tt.Service.FindUserByEmail(context.Background(), req.Email)
	if err != nil {
		t.Error(err)
	}

	test.AssertInDB(t, tt.DB, "email_confirmations", test.Data{
		"user_id": user.ID,
		"code":    vars["code"],
	})

	test.AssertInDB(t, tt.DB, "user_activity_logs", test.Data{
		"user_id":     user.ID,
		"action_type": useractivity.UserRegistered,
		"is_public":   false, // "New user" event should not be public by default
	})

	t.Run("username already taken", func(t *testing.T) {
		req := request.UserSignUp{
			Username: user.Username,
			Email:    random.Email(),
			Password: random.String(),
		}

		var resp handler.Error
		POST(t, Request{
			path:         "/users/sign-up",
			body:         req,
			bindResponse: &resp,
			assertStatus: http.StatusUnprocessableEntity,
		})

		assert.Equal(t, resp.InputErrors["username"], service.UsernameAlreadyTaken)
	})

	t.Run("email already taken", func(t *testing.T) {
		req := request.UserSignUp{
			Username: random.String(),
			Email:    user.Email,
			Password: random.String(),
		}

		var resp handler.Error
		POST(t, Request{
			path:         "/users/sign-up",
			body:         req,
			bindResponse: &resp,
			assertStatus: http.StatusUnprocessableEntity,
		})

		assert.Equal(t, resp.InputErrors["email"], service.UserWithThisEmailAlreadyExists)
	})
}

func TestUserConfirmEmail(t *testing.T) {
	ec := tt.Factory.CreateEmailConfirmation(t)
	log := tt.Factory.CreateUserActivityLog(t, ds.UserActivityLog{
		UserID:     ec.UserID,
		ActionType: useractivity.UserRegistered,
	})

	req := request.ConfirmEmail{
		Code: ec.Code,
	}

	var resp response.Status
	POST(t, Request{
		path:         "/users/confirm-email",
		body:         req,
		bindResponse: &resp,
		assertStatus: http.StatusOK,
	})

	test.AssertInDB(t, tt.DB, "users", test.Data{
		"id":              ec.UserID,
		"email_confirmed": true,
	})

	test.AssertNotInDB(t, tt.DB, "email_confirmations", test.Data{
		"code": ec.Code,
	})

	test.AssertInDB(t, tt.DB, "user_activity_logs", test.Data{
		"id":          log.ID,
		"user_id":     ec.UserID,
		"action_type": useractivity.UserRegistered,
		"is_public":   true,
	})
}

func TestUserSignIn(t *testing.T) {
	password := random.String()
	user := tt.Factory.CreateUser(t, ds.User{
		Password: password,
	})

	req := request.UserSignIn{
		Email:    user.Email,
		Password: password,
	}

	var resp response.UserSignIn
	POST(t, Request{
		path:         "/users/sign-in",
		body:         req,
		bindResponse: &resp,
		assertStatus: http.StatusOK,
	})
}

func TestChangePassword(t *testing.T) {
	oldPassword := random.String(10)
	newPassword := random.String(10)

	user := tt.Factory.CreateUser(t, ds.User{Password: oldPassword})

	_, token, err := tt.Service.AuthenticateUser(context.Background(), user.Email, oldPassword)
	if err != nil {
		t.Fatal(err)
	}

	req := request.ChangePassword{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	}

	var resp response.Status
	PUT(t, Request{
		path:         "/users/password/",
		body:         req,
		authToken:    token,
		bindResponse: &resp,
		assertStatus: http.StatusOK,
	})

	// Login with the old password
	var signInResp handler.Error
	POST(t, Request{
		path: "/users/sign-in/",
		body: request.UserSignIn{
			Email:    user.Email,
			Password: oldPassword,
		},
		bindResponse: &signInResp,
		assertStatus: http.StatusUnprocessableEntity,
	})

	t.Run("login with new password", func(t *testing.T) {
		var signInResp response.UserSignIn
		POST(t, Request{
			path: "/users/sign-in/",
			body: request.UserSignIn{
				Email:    user.Email,
				Password: newPassword,
			},
			bindResponse: &signInResp,
			assertStatus: http.StatusOK,
		})
	})

	t.Run("incorrect old password", func(t *testing.T) {
		req := request.ChangePassword{
			OldPassword: "incorrect-password",
			NewPassword: newPassword,
		}

		var resp handler.Error
		PUT(t, Request{
			path:         "/users/password/",
			body:         req,
			authToken:    token,
			bindResponse: &resp,
			assertStatus: http.StatusUnprocessableEntity,
		})

		assert.Equal(t, resp.InputErrors["old_password"], service.ErrInvalidPassword.Error())
	})
}

func TestPasswordReset(t *testing.T) {
	user := tt.Factory.CreateUser(t)

	// 1. Request password reset
	var reqResetResp response.Status
	POST(t, Request{
		path:         "users/password-reset-request",
		body:         request.PasswordResetRequest{Email: user.Email},
		bindResponse: &reqResetResp,
		assertStatus: http.StatusOK,
	})

	test.AssertInDB(t, tt.DB, "password_reset_tokens", test.Data{"user_id": user.ID})

	emailVars := test.LoadEmailVars(t, user.Email)
	token := app.String(emailVars["token"])
	assert.NotZero(t, token)

	// 2. Successfully reset the password
	newPassword := random.String()
	var resetResp response.Status
	POST(t, Request{
		path: "users/password-reset",
		body: request.PasswordReset{
			Token:    token,
			Password: newPassword,
		},
		bindResponse: &resetResp,
		assertStatus: http.StatusOK,
	})

	// Assert the authToken was deleted
	test.AssertNotInDB(t, tt.DB, "password_reset_tokens", test.Data{"token": token})

	// 3. Verify login with the new password
	var signInResp response.UserSignIn
	POST(t, Request{
		path: "users/sign-in",
		body: request.UserSignIn{
			Email:    user.Email,
			Password: newPassword,
		},
		bindResponse: &signInResp,
		assertStatus: http.StatusOK,
	})

	// 4. Test failure cases
	t.Run("reset with invalid token", func(t *testing.T) {
		var errorResp handler.Error
		POST(t, Request{
			path: "users/password-reset",
			body: request.PasswordReset{
				Token:    "invalid-token",
				Password: newPassword,
			},
			bindResponse: &errorResp,
			assertStatus: http.StatusUnprocessableEntity,
		})
	})

	t.Run("reset with password too short", func(t *testing.T) {
		prt := tt.Factory.CreatePasswordResetToken(t, ds.PasswordResetToken{
			UserID: user.ID,
		})
		var errorResp handler.Error
		POST(t, Request{
			path: "users/password-reset",
			body: request.PasswordReset{
				Token:    prt.Token,
				Password: strings.Repeat("a", service.UserPasswordMinLen-1),
			},
			bindResponse: &errorResp,
			assertStatus: http.StatusUnprocessableEntity,
		})
		assert.NotZero(t, errorResp.InputErrors["password"])
	})
}

func TestChangeEmail(t *testing.T) {
	user := tt.Factory.CreateUser(t)
	token := loginAs(t, user)

	// Request email change
	newEmail := random.Email()
	var reqEmailChangeResp response.Status
	POST(t, Request{
		path:         "/users/email/",
		body:         request.EmailChangeRequest{Email: newEmail},
		authToken:    token,
		bindResponse: &reqEmailChangeResp,
		assertStatus: http.StatusOK,
	})

	test.AssertInDB(t, tt.DB, "change_email_requests", test.Data{
		"user_id":   user.ID,
		"new_email": newEmail,
	})

	emailVars := test.LoadEmailVars(t, newEmail)
	confirmToken := app.String(emailVars["token"])
	assert.NotZero(t, confirmToken)

	// Confirm the email change with the confirmToken
	var confirmResp response.Status
	PUT(t, Request{
		path: "/users/email/",
		body: request.EmailChangeConfirm{
			Token: confirmToken,
		},
		authToken:    token,
		bindResponse: &confirmResp,
		assertStatus: http.StatusOK,
	})

	test.AssertInDB(t, tt.DB, "users", test.Data{
		"id":    user.ID,
		"email": newEmail,
	})

	test.AssertNotInDB(t, tt.DB, "change_email_requests", test.Data{
		"token": confirmToken,
	})

	// Test failure case: using the same authToken again
	var errorResp handler.Error
	PUT(t, Request{
		path: "/users/email/",
		body: request.EmailChangeConfirm{
			Token: confirmToken,
		},
		authToken:    token,
		bindResponse: &errorResp,
		assertStatus: http.StatusUnprocessableEntity,
	})

	assert.Equal(t, service.ErrInvalidChangeEmailToken.Error(), errorResp.Error)
}

func TestChangeUsername(t *testing.T) {
	password := random.String(10)
	user := tt.Factory.CreateUser(t, ds.User{Password: password})
	token := loginAs(t, user)
	newUsername := random.String(10)

	t.Run("successful username change", func(t *testing.T) {
		req := request.ChangeUsername{
			Username: newUsername,
			Password: password,
		}

		var resp response.Status
		PUT(t, Request{
			path:         "/users/username/",
			body:         req,
			authToken:    token,
			bindResponse: &resp,
			assertStatus: http.StatusOK,
		})

		test.AssertInDB(t, tt.DB, "users", test.Data{
			"id":       user.ID,
			"username": newUsername,
		})
	})

	t.Run("incorrect password", func(t *testing.T) {
		req := request.ChangeUsername{
			Username: random.String(10),
			Password: "wrong-password",
		}

		var resp handler.Error
		PUT(t, Request{
			path:         "/users/username/",
			body:         req,
			authToken:    token,
			bindResponse: &resp,
			assertStatus: http.StatusUnprocessableEntity,
		})

		assert.Equal(t, resp.InputErrors["password"], "Incorrect password")
	})

	t.Run("username already taken", func(t *testing.T) {
		otherUser := tt.Factory.CreateUser(t)
		req := request.ChangeUsername{
			Username: otherUser.Username,
			Password: password,
		}

		var resp handler.Error
		PUT(t, Request{
			path:         "/users/username/",
			body:         req,
			authToken:    token,
			bindResponse: &resp,
			assertStatus: http.StatusUnprocessableEntity,
		})

		assert.Equal(t, resp.InputErrors["username"], service.UsernameAlreadyTaken)
	})
}
