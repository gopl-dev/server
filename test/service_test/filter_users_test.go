package service_test

import (
	"context"
	"testing"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test"
)

func TestFilterUsers(t *testing.T) {
	create[ds.User](t)

	_, _, err := tt.Service.FilterUsers(context.Background(), ds.UsersFilter{})
	test.CheckErr(t, err)
}
