package api_test

import (
	"testing"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory/random"
)

func TestCreateBook(t *testing.T) {
	user := login(t)

	req := request.CreateBook{
		Title:       random.Title(),
		Description: random.String(),
		ReleaseDate: random.String(),
		AuthorName:  random.String(),
		AuthorLink:  random.URL(),
		Homepage:    random.URL(),
		CoverImage:  random.URL(),
		Visibility:  random.Element(ds.EntityVisibilities),
	}

	var resp ds.Book
	testCREATE(t, "books", req, &resp)

	// check entity created
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":         resp.ID,
		"title":      req.Title,
		"owner_id":   user.ID,
		"type":       ds.EntityTypeBook,
		"visibility": req.Visibility,
		"url_name":   app.Slug(req.Title),
	})

	// check book created
	test.AssertInDB(t, tt.DB, "books", test.Data{
		"description": req.Description,
		"author_name": req.AuthorName,
		"author_link": req.AuthorLink,
		"homepage":    req.Homepage,
		"cover_image": req.CoverImage,
	})

	// check log created
	test.AssertInDB(t, tt.DB, "entity_change_logs", test.Data{
		"entity_id": resp.ID,
		"user_id":   user.ID,
		"action":    ds.ActionCreate,
	})
}
