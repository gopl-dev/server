package api_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
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

	imageBytes, err := random.ImagePNG(10)
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

func TestUpdateBook_WithReview(t *testing.T) {
	user := login(t)

	book := create[ds.Book](t)
	var resp service.EntityChange
	GET(t, pf("/books/%s/edit/", book.ID), &resp)

	// first response's data should be same as book and with revision=0
	assert.Equal(t, book.ID, resp.ID)
	assert.Equal(t, 0, resp.Revision)
	assert.True(t, len(book.Data()) == len(resp.Data))

	for k, v := range book.Data() {
		assert.Equal(t, fmt.Sprintf("%v", v), fmt.Sprintf("%v", resp.Data[k]))
	}

	// do update (change only title and description)
	updateReq := request.UpdateBook{
		CreateBook: request.CreateBook{
			Title:       random.Title(),
			Description: random.String(),
			ReleaseDate: book.ReleaseDate,
			AuthorName:  book.AuthorName,
			AuthorLink:  book.AuthorLink,
			Homepage:    book.Homepage,
			CoverFileID: book.CoverFileID,

			Visibility: book.Visibility,
		},
	}
	var updateResp response.UpdateBook
	UPDATE(t, pf("/books/%s/", book.ID), updateReq, &updateResp)

	assert.Equal(t, 1, updateResp.Revision)

	// new change request should be created
	test.AssertInDB(t, tt.DB, "entity_change_requests", test.Data{
		"user_id":   user.ID,
		"entity_id": book.ID,
		"status":    ds.EntityChangePending,
		"revision":  1,
		"diff":      map[string]any{"title": updateReq.Title, "description": updateReq.Description},
	})

	// book itself should not be changed
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":    book.ID,
		"title": resp.Data["title"],
	})
	test.AssertInDB(t, tt.DB, "books", test.Data{
		"id":          book.ID,
		"description": resp.Data["description"],
	})

	// next edit should return "in-progress" data
	GET(t, pf("/books/%s/edit/", book.ID), &resp)
	assert.Equal(t, book.ID, resp.ID)
	assert.Equal(t, 1, resp.Revision) // revision should increase
	assert.True(t, len(book.Data()) == len(resp.Data))

	assert.Equal(t, any(updateReq.Title), resp.Data["title"])
	assert.Equal(t, any(updateReq.Description), resp.Data["description"])

	// updating book that already have change request for review
	// should only update that request
	updateReq.AuthorName = random.String()
	UPDATE(t, pf("/books/%s/", book.ID), updateReq, &updateResp)
	test.AssertInDB(t, tt.DB, "entity_change_requests", test.Data{
		"user_id":    user.ID,
		"entity_id":  book.ID,
		"status":     ds.EntityChangePending,
		"revision":   2,            // revision should be incremented
		"updated_at": test.NotNull, // updated_at should be set
		"diff": map[string]any{
			"title":       updateReq.Title,
			"description": updateReq.Description,
			"author_name": updateReq.AuthorName,
		},
	})
}

func TestUpdateBook_WithoutReview(t *testing.T) {
	user := create[ds.User](t)
	token := loginAs(t, user)

	app.Config().Admins = []string{user.ID.String()}

	imageBytes1, err := random.ImagePNG(10)
	test.CheckErr(t, err)
	cover1 := UploadFile(t, fileForm{
		purpose:  ds.FilePurposeBookCover,
		filename: "cover.jpg",
		file:     bytes.NewReader(imageBytes1),
	})
	imageBytes2, err := random.ImagePNG(10)
	test.CheckErr(t, err)
	cover2 := UploadFile(t, fileForm{
		purpose:  ds.FilePurposeBookCover,
		filename: "cover.jpg",
		file:     bytes.NewReader(imageBytes2),
	})

	book := create(t, ds.Book{
		CoverFileID: cover1.ID,
	})

	// do update (change only title and description)
	req := request.UpdateBook{
		CreateBook: request.CreateBook{
			Title:       random.Title(),
			Description: random.String(),
			ReleaseDate: book.ReleaseDate,
			AuthorName:  book.AuthorName,
			AuthorLink:  book.AuthorLink,
			Homepage:    book.Homepage,
			CoverFileID: cover2.ID,

			Visibility: book.Visibility,
		},
	}
	var resp response.UpdateBook
	UPDATE(t, pf("/books/%s/", book.ID), req, &resp)
	Request(t, RequestArgs{
		method:       http.MethodPut,
		path:         pf("/books/%s/", book.ID),
		body:         req,
		authToken:    token,
		bindResponse: &resp,
		assertStatus: http.StatusOK,
	})

	assert.Equal(t, 0, resp.Revision)

	// book should be updated
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":              book.ID,
		"title":           req.Title,
		"preview_file_id": cover2.ID,
	})
	test.AssertInDB(t, tt.DB, "books", test.Data{
		"id":            book.ID,
		"description":   req.Description,
		"cover_file_id": cover2.ID,
	})
	test.AssertInDB(t, tt.DB, "files", test.Data{
		"id":   cover2.ID,
		"temp": false,
	})
	test.AssertInDB(t, tt.DB, "files", test.Data{
		"id":         cover1.ID,
		"deleted_at": test.NotNull,
	})
}
