package worker_test

import (
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test"
	deleteunconfirmedusers "github.com/gopl-dev/server/worker/delete_unconfirmed_users"
)

func TestDeleteUnconfirmedUsers(t *testing.T) {
	user := tt.Factory.CreateUser(t, ds.User{
		EmailConfirmed: false,
		CreatedAt:      time.Now().Add(-25 * time.Hour),
	})

	runJob(t, deleteunconfirmedusers.NewJob())

	test.AssertInDB(t, tt.DB, "users", test.Data{
		"id":              user.ID,
		"deleted_at":      test.NotNull,
		"email_confirmed": false,
	})
}
