package api_test

import (
	"bytes"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory"
	"github.com/gopl-dev/server/test/factory/random"
)

// TestCreateBook_Basic is a minimal effort to create valid book.
func TestCreateBook_Basic(t *testing.T) {
	user := login(t)

	req := request.CreateBook{
		Title:       random.Title(),
		Description: random.String(),
		ReleaseDate: random.String(),
		AuthorName:  random.String(),
		AuthorLink:  random.URL(),
		Homepage:    random.URL(),
		Visibility:  random.Element(ds.EntityVisibilities),
	}

	var resp ds.Book
	CREATE(t, "books", req, &resp)

	// check entity created
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":         resp.ID,
		"public_id":  app.Slug(req.Title),
		"title":      req.Title,
		"owner_id":   user.ID,
		"type":       ds.EntityTypeBook,
		"visibility": req.Visibility,
	})

	// check book created
	test.AssertInDB(t, tt.DB, "books", test.Data{
		"description": req.Description,
		"author_name": req.AuthorName,
		"author_link": req.AuthorLink,
		"homepage":    req.Homepage,
	})

	// check log created
	test.AssertInDB(t, tt.DB, "entity_change_logs", test.Data{
		"entity_id": resp.ID,
		"user_id":   user.ID,
		"action":    ds.ActionCreate,
	})
}

// TestCreateBook_WithCover is a minimal effort to create valid book with cover.
func TestCreateBook_WithCover(t *testing.T) {
	login(t)

	imageBytes, err := random.ImagePNG()
	test.CheckErr(t, err)

	cover := UploadFile(t, fileForm{
		purpose:  ds.FilePurposeBookCover,
		filename: "cover.jpg",
		file:     bytes.NewReader(imageBytes),
	})

	// uploaded file without entity should be temporary
	test.AssertInDB(t, tt.DB, "files", test.Data{
		"id":   cover.ID,
		"temp": true,
	})

	req := request.CreateBook{
		Title:       random.Title(),
		Description: random.String(),
		ReleaseDate: random.String(),
		AuthorName:  random.String(),
		AuthorLink:  random.URL(),
		Homepage:    random.URL(),
		Visibility:  random.Element(ds.EntityVisibilities),
		CoverFileID: cover.ID,
	}

	var resp ds.Book
	CREATE(t, "books", req, &resp)

	// check entity created
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":              resp.ID,
		"preview_file_id": cover.ID,
	})

	test.AssertInDB(t, tt.DB, "books", test.Data{
		"id":            resp.ID,
		"cover_file_id": cover.ID,
	})

	// temp flag should be switched
	test.AssertInDB(t, tt.DB, "files", test.Data{
		"id":   cover.ID,
		"temp": false,
	})
}

func TestFilterBooks(t *testing.T) {
	login(t)

	_, err := factory.Ten(tt.Factory.CreateBook, ds.Book{
		Entity: &ds.Entity{
			Status:     ds.EntityStatusApproved,
			Visibility: ds.EntityVisibilityPublic,
			DeletedAt:  nil,
		},
	})
	test.CheckErr(t, err)

	req := Query{
		Path: "books",
		Params: request.FilterEntities{
			Page:    1,
			PerPage: 2,
		},
	}

	var resp response.FilterBooks
	GET(t, req, &resp)

	assert.Equal(t, 2, len(resp.Data))

	t.Run("pagination", func(t *testing.T) {
		req.Params = request.FilterEntities{
			Page:    2,
			PerPage: 3,
		}

		GET(t, req, &resp)
		assert.Equal(t, 3, len(resp.Data))
	})
}
