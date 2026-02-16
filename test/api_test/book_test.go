package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/server/handler"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory"
	"github.com/gopl-dev/server/test/factory/random"
)

// TestCreateBook_Basic is a minimal effort to create valid book.
func TestCreateBook_Basic(t *testing.T) {
	user := login(t)

	topic := create(t, ds.Topic{Type: ds.EntityTypeBook})

	req := request.CreateBook{
		Title:       random.Title(),
		Summary:     random.String(),
		Description: random.String(),
		ReleaseDate: random.ReleaseDate(),
		Authors:     factory.NewBookAuthors(),
		Homepage:    random.URL(),
		Topics:      []ds.ID{topic.ID},
	}

	var resp ds.Book
	CREATE(t, "books", req, &resp)

	summaryHTML, err := app.MarkdownToHTML(req.Summary)
	test.CheckErr(t, err)
	assert.Equal(t, resp.Summary, summaryHTML)

	descriptionHTML, err := app.MarkdownToHTML(req.Description)
	test.CheckErr(t, err)
	assert.Equal(t, resp.Description, descriptionHTML)

	// check entity created
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":          resp.ID,
		"public_id":   resp.PublicID,
		"title":       req.Title,
		"summary_raw": req.Summary,
		"summary":     summaryHTML,
		"owner_id":    user.ID,
		"type":        ds.EntityTypeBook,
		"status":      ds.EntityStatusUnderReview,
	})

	// check book created
	test.AssertInDB(t, tt.DB, "books", test.Data{
		"authors":         req.Authors,
		"homepage":        req.Homepage,
		"description_raw": req.Description,
		"description":     descriptionHTML,
	})

	// check log created
	test.AssertInDB(t, tt.DB, "event_logs", test.Data{
		"entity_id": resp.ID,
		"user_id":   user.ID,
		"type":      ds.EventLogEntitySubmitted,
		"is_public": false,
	})

	// topic attached
	test.AssertInDB(t, tt.DB, "entity_topics", test.Data{
		"entity_id": resp.ID,
		"topic_id":  topic.ID,
	})

	t.Run("book without topic", func(t *testing.T) {
		req.Topics = []ds.ID{}
		var errResp handler.Error
		Request(t, RequestArgs{
			method:       http.MethodPost,
			path:         "/books/",
			body:         req,
			bindResponse: &errResp,
			assertStatus: http.StatusUnprocessableEntity,
		})

		errText, ok := errResp.InputErrors["topics"]
		assert.True(t, ok)
		assert.True(t, errText != "")
	})
}

// TestCreateBook_WithCover is a minimal effort to create valid book with cover.
func TestCreateBook_WithCover(t *testing.T) {
	login(t)

	topic := create(t, ds.Topic{Type: ds.EntityTypeBook})

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
		ReleaseDate: random.ReleaseDate(),
		Authors:     factory.NewBookAuthors(),
		Homepage:    random.URL(),
		CoverFileID: cover.ID,
		Topics:      []ds.ID{topic.ID},
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

	var data map[string]any
	jsonData, err := json.Marshal(book.Data())
	test.CheckErr(t, err)
	err = json.Unmarshal(jsonData, &data)
	test.CheckErr(t, err)

	for k, v := range data {
		assert.Equal(t, fmt.Sprintf("%v", v), fmt.Sprintf("%v", resp.Data[k]))
	}

	// do update (change only summary)
	updateReq := request.UpdateBook{
		CreateBook: request.CreateBook{
			Title:       book.Title,
			Summary:     random.Edit(book.Title),
			Description: book.Description,
			ReleaseDate: book.ReleaseDate,
			Authors:     book.Authors,
			Homepage:    book.Homepage,
			CoverFileID: book.CoverFileID,
		},
	}

	var updateResp ds.EntityChangeRequest
	UPDATE(t, pf("/books/%s/", book.ID), updateReq, &updateResp)

	assert.Equal(t, 1, updateResp.Revision)
	assert.Equal(t, ds.EntityChangePending, updateResp.Status)

	summaryPatch := app.MakePatch(book.Summary, updateReq.Summary)

	// new change request should be created
	test.AssertInDB(t, tt.DB, "entity_change_requests", test.Data{
		"user_id":   user.ID,
		"entity_id": book.ID,
		"status":    ds.EntityChangePending,
		"revision":  1,
		"diff": map[string]any{
			"summary": summaryPatch,
		},
	})

	// book itself should not be changed
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":    book.ID,
		"title": resp.Data["title"],
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
	updateReq.Description = random.Edit(book.Description)
	descriptionPatch := app.MakePatch(book.Description, updateReq.Description)
	UPDATE(t, pf("/books/%s/", book.ID), updateReq, &updateResp)
	test.AssertInDB(t, tt.DB, "entity_change_requests", test.Data{
		"user_id":    user.ID,
		"entity_id":  book.ID,
		"status":     ds.EntityChangePending,
		"revision":   2,            // revision should be incremented
		"updated_at": test.NotNull, // updated_at should be set
		"diff": map[string]any{
			"summary":     summaryPatch,
			"description": descriptionPatch,
		},
	})
}

func TestUpdateBook_WithoutReview(t *testing.T) {
	user := login(t)
	makeAdmin(user)

	imageBytes1, err := random.ImagePNG(10)
	test.CheckErr(t, err)
	cover1 := UploadFile(t, fileForm{
		authToken: authToken,
		purpose:   ds.FilePurposeBookCover,
		filename:  "cover.jpg",
		file:      bytes.NewReader(imageBytes1),
	})
	imageBytes2, err := random.ImagePNG(10)
	test.CheckErr(t, err)
	cover2 := UploadFile(t, fileForm{
		authToken: authToken,
		purpose:   ds.FilePurposeBookCover,
		filename:  "cover.jpg",
		file:      bytes.NewReader(imageBytes2),
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
			Authors:     factory.NewBookAuthors(),
			Homepage:    book.Homepage,
			CoverFileID: cover2.ID,
		},
	}
	var resp ds.EntityChangeRequest
	Request(t, RequestArgs{
		method:       http.MethodPut,
		path:         pf("/books/%s/", book.ID),
		body:         req,
		authToken:    authToken,
		bindResponse: &resp,
		assertStatus: http.StatusOK,
	})

	assert.Equal(t, 1, resp.Revision)
	assert.Equal(t, ds.EntityChangeCommitted, resp.Status)

	descriptionHTML, err := app.MarkdownToHTML(req.Description)
	test.CheckErr(t, err)

	// book should be updated
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":              book.ID,
		"title":           req.Title,
		"preview_file_id": cover2.ID,
	})
	test.AssertInDB(t, tt.DB, "books", test.Data{
		"id":            book.ID,
		"cover_file_id": cover2.ID,
		"description":   descriptionHTML,
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

func TestApproveNewBook(t *testing.T) {
	user := login(t)
	makeAdmin(user)

	book := create(t, ds.Book{
		Entity: &ds.Entity{
			Type:   ds.EntityTypeBook,
			Status: ds.EntityStatusUnderReview,
		},
	})

	var resp response.Status
	UPDATE(t, pf("/books/%s/approve/", book.ID), struct{}{}, &resp)

	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":     book.ID,
		"status": ds.EntityStatusApproved,
	})

	test.AssertInDB(t, tt.DB, "event_logs", test.Data{
		"user_id":   user.ID,
		"type":      ds.EventLogEntityApproved,
		"entity_id": book.ID,
		"is_public": false,
	})

	test.AssertInDB(t, tt.DB, "event_logs", test.Data{
		"user_id":   book.OwnerID,
		"type":      ds.EventLogEntityAdded,
		"entity_id": book.ID,
		"is_public": true,
	})

	owner, err := tt.Service.GetUserByID(context.Background(), book.OwnerID)
	test.CheckErr(t, err)

	emailVars := test.LoadEmailVars(t, owner.Email)
	assert.Equal(t, emailVars, map[string]any{
		"username":      owner.Username,
		"book_name":     book.Title,
		"view_book_url": app.ServerURL("/books/" + book.PublicID + "/"),
	})
}

func TestRejectNewBook(t *testing.T) {
	user := login(t)
	makeAdmin(user)

	book := create(t, ds.Book{
		Entity: &ds.Entity{
			Type:   ds.EntityTypeBook,
			Status: ds.EntityStatusUnderReview,
		},
	})

	req := request.RejectBook{
		Note: random.String(),
	}
	var resp response.Status
	UPDATE(t, pf("/books/%s/reject/", book.ID), req, &resp)

	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":     book.ID,
		"status": ds.EntityStatusRejected,
	})

	test.AssertInDB(t, tt.DB, "event_logs", test.Data{
		"user_id":   user.ID,
		"type":      ds.EventLogEntityRejected,
		"entity_id": book.ID,
		"meta":      map[string]any{"note": req.Note},
		"is_public": false,
	})

	owner, err := tt.Service.GetUserByID(context.Background(), book.OwnerID)
	test.CheckErr(t, err)

	emailVars := test.LoadEmailVars(t, owner.Email)
	assert.Equal(t, emailVars, map[string]any{
		"username":  owner.Username,
		"book_name": book.Title,
		"note":      req.Note,
	})
}

func TestDeleteBook(t *testing.T) {
	user := login(t)
	makeAdmin(user)

	book := create[ds.Book](t)

	var resp response.Status
	DELETE(t, pf("/books/%s/", book.ID), &resp)

	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":         book.ID,
		"deleted_at": test.NotNull,
	})
}
