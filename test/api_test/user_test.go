package api_test

import (
	"context"
	"net/http"
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
