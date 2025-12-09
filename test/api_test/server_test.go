package api_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/gopl-dev/server/server/response"
)

func TestGetServerStatus(t *testing.T) {
	var resp response.ServerStatus
	testGET(t, "status", &resp)

	assert.Equal(t, resp.Env, tt.Conf.App.Env)
	assert.Equal(t, resp.Version, tt.Conf.App.Version)
}
