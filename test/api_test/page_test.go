package api_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/server/handler"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory/random"
)

func TestCreatePage(t *testing.T) {
	// only admins can create pages now
	user := create[ds.User](t)
	token := loginAs(t, user)

	app.Config().Admins = []string{user.ID.String()}

	req := request.CreatePage{
		PublicID: random.String(),
		Title:    random.Title(),
		Content:  random.String(),
	}

	var resp ds.Page
	Request(t, RequestArgs{
		method:       http.MethodPost,
		path:         "/pages/",
		body:         req,
		authToken:    token,
		bindResponse: &resp,
		assertStatus: http.StatusCreated,
	})

	contentHTML, err := app.MarkdownToHTML(req.Content)
	test.CheckErr(t, err)
	assert.Equal(t, resp.Content, contentHTML)

	// check entity created
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":         resp.ID,
		"public_id":  req.PublicID,
		"title":      req.Title,
		"owner_id":   user.ID,
		"type":       ds.EntityTypePage,
		"status":     ds.EntityStatusApproved,
		"visibility": ds.EntityVisibilityPublic,
	})

	// check page created
	test.AssertInDB(t, tt.DB, "pages", test.Data{
		"id":          resp.ID,
		"content_raw": req.Content,
		"content":     contentHTML,
	})

	// check log created
	test.AssertInDB(t, tt.DB, "event_logs", test.Data{
		"entity_id": resp.ID,
		"user_id":   user.ID,
		"type":      ds.EventLogEntityAdded,
	})

	t.Run("public_id already taken", func(t *testing.T) {
		var errResp handler.Error
		Request(t, RequestArgs{
			method:       http.MethodPost,
			path:         "/pages/",
			body:         req,
			authToken:    token,
			bindResponse: &errResp,
			assertStatus: http.StatusUnprocessableEntity,
		})

		errMsg, ok := errResp.InputErrors["public_id"]
		assert.True(t, ok)
		assert.True(t, errMsg != "")
	})
}

func TestUpdatePage_WithReview(t *testing.T) {
	user := login(t)

	page := create[ds.Page](t)
	var resp service.EntityChange
	GET(t, pf("/pages/%s/edit/", page.PublicID), &resp)

	// first response's data should be same as page and with revision=0
	assert.Equal(t, page.ID, resp.ID)
	assert.Equal(t, 0, resp.Revision)
	assert.True(t, len(page.Data()) == len(resp.Data))

	for k, v := range page.Data() {
		assert.Equal(t, fmt.Sprintf("%v", v), fmt.Sprintf("%v", resp.Data[k]))
	}

	// do update (change only content)
	updateReq := request.UpdatePage{
		CreatePage: request.CreatePage{
			PublicID: page.PublicID,
			Title:    page.Title,
			Content:  random.String(),
		},
	}
	var updateResp ds.EntityChangeRequest
	UPDATE(t, pf("/pages/%s/", page.PublicID), updateReq, &updateResp)

	assert.Equal(t, 1, updateResp.Revision)
	assert.Equal(t, ds.EntityChangePending, updateResp.Status)

	// new change request should be created
	test.AssertInDB(t, tt.DB, "entity_change_requests", test.Data{
		"user_id":   user.ID,
		"entity_id": page.ID,
		"status":    ds.EntityChangePending,
		"revision":  1,
		"diff":      map[string]any{"content": updateReq.Content},
	})

	// page itself should not be changed
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":    page.ID,
		"title": resp.Data["title"],
	})

	// next edit should return "in-progress" data
	GET(t, pf("/pages/%s/edit/", page.PublicID), &resp)
	assert.Equal(t, page.ID, resp.ID)
	assert.Equal(t, 1, resp.Revision) // revision should increase
	assert.True(t, len(page.Data()) == len(resp.Data))
	assert.Equal(t, any(updateReq.Content), resp.Data["content"])

	// updating page that already have change request for review
	// should only update that request
	updateReq.Title = random.String()
	UPDATE(t, pf("/pages/%s/", page.PublicID), updateReq, &updateResp)
	test.AssertInDB(t, tt.DB, "entity_change_requests", test.Data{
		"user_id":    user.ID,
		"entity_id":  page.ID,
		"status":     ds.EntityChangePending,
		"revision":   2,            // revision should be incremented
		"updated_at": test.NotNull, // updated_at should be set
		"diff": map[string]any{
			"title":   updateReq.Title,
			"content": updateReq.Content,
		},
	})
}

func TestUpdatePage_WithoutReview(t *testing.T) {
	user := create[ds.User](t)
	token := loginAs(t, user)

	app.Config().Admins = []string{user.ID.String()}

	page := create[ds.Page](t)

	// do update (change only content)
	req := request.UpdatePage{
		CreatePage: request.CreatePage{
			PublicID: page.PublicID,
			Title:    page.Title,
			Content:  random.String(),
		},
	}
	var resp ds.EntityChangeRequest
	Request(t, RequestArgs{
		method:       http.MethodPut,
		path:         pf("/pages/%s/", page.PublicID),
		body:         req,
		authToken:    token,
		bindResponse: &resp,
		assertStatus: http.StatusOK,
	})

	assert.Equal(t, 1, resp.Revision)
	assert.Equal(t, ds.EntityChangeCommitted, resp.Status)

	// page should be updated
	test.AssertInDB(t, tt.DB, "pages", test.Data{
		"id":          page.ID,
		"content_raw": req.Content,
	})

	test.AssertInDB(t, tt.DB, "entity_change_requests", test.Data{
		"user_id":   user.ID,
		"entity_id": page.ID,
		"status":    ds.EntityChangeCommitted,
		"revision":  1,
		"diff":      map[string]any{"content": req.Content},
	})
}
