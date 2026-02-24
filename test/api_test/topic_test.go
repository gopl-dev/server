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

func TestFilterTopics(t *testing.T) {
	login(t)

	_, err := factory.Ten(tt.Factory.CreateTopic, ds.Topic{
		Type: ds.EntityTypeBook,
	})
	test.CheckErr(t, err)

	req := Query{
		Path: "topics",
		Params: request.FilterTopics{
			PerPage: 10,
			Type:    ds.EntityTypeBook,
		},
	}

	var resp response.FilterTopics
	GET(t, req, &resp)

	assert.Len(t, resp.Data, 10)

	t.Run("pagination", func(t *testing.T) {
		req.Params = request.FilterTopics{
			Page:    2,
			PerPage: 3,
			Type:    ds.EntityTypeBook,
		}

		GET(t, req, &resp)
		assert.Len(t, resp.Data, 3)
	})
}
