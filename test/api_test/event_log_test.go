package api_test

import (
	"testing"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory"
	"github.com/stretchr/testify/assert"
)

func TestFilterEventLogs(t *testing.T) {
	_, err := factory.Ten(tt.Factory.CreateEventLog, ds.EventLog{
		IsPublic: true,
	})
	test.CheckErr(t, err)

	req := Query{
		Path: "event-logs",
		Params: request.FilterEntities{
			Page:    1,
			PerPage: 10,
		},
	}

	var resp response.FilterEventLogs
	GET(t, req, &resp)

	assert.Len(t, resp.Data, 10)

	t.Run("pagination", func(t *testing.T) {
		req.Params = request.FilterEntities{
			Page:    2,
			PerPage: 3,
		}

		GET(t, req, &resp)
		assert.Len(t, resp.Data, 3)
	})
}

func TestGetEventLogChanges(t *testing.T) {
	log := create[ds.EventLog](t)

	var resp response.EventLogChanges
	GET(t, pf("event-logs/%s/changes/", log.ID), &resp)
}
