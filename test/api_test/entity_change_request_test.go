package api_test

import (
	"bytes"
	"testing"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/server/request"
	"github.com/gopl-dev/server/server/response"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory"
	"github.com/gopl-dev/server/test/factory/random"
	"github.com/stretchr/testify/assert"
)

func TestGetChangeRequestDiff(t *testing.T) {
	admin := create[ds.User](t)
	loginAs(t, admin)
	makeAdmin(admin)

	_, err := factory.Ten(tt.Factory.CreateEntityChangeRequest, ds.EntityChangeRequest{
		Status: ds.EntityChangePending,
	})
	test.CheckErr(t, err)

	var resp response.FilterChangeRequests
	GET(t, "change-requests/?status=pending", &resp)

	assert.NotEmpty(t, resp.Data)
	assert.Equal(t, ds.EntityChangePending, resp.Data[0].Status)
}

func TestFilterChangeRequest(t *testing.T) {
	user := create[ds.User](t)
	book := create[ds.Book](t)

	cr := create(t, ds.EntityChangeRequest{
		EntityID: book.ID,
		UserID:   user.ID,
		Status:   ds.EntityChangePending,
		Diff: map[string]any{
			"title": random.String(),
		},
	})

	var resp service.ChangeDiff
	GET(t, pf("change-requests/%s/diff/", cr.ID), &resp)

	assert.Equal(t, book.Title, resp.Current["title"])
	assert.Equal(t, cr.Diff["title"], resp.Proposed["title"])
}

func TestRejectChangeRequest(t *testing.T) {
	admin := create[ds.User](t)
	loginAs(t, admin)
	makeAdmin(admin)

	user := create[ds.User](t)
	book := create[ds.Book](t)
	cr := create(t, ds.EntityChangeRequest{
		EntityID: book.ID,
		UserID:   user.ID,
		Status:   ds.EntityChangePending,
		Diff: map[string]any{
			"title": random.String(),
		},
	})

	req := request.RejectChangeRequest{
		Note: random.String(),
	}
	var resp response.Status
	UPDATE(t, pf("change-requests/%s/reject/", cr.ID), req, &resp)

	// change request status should be changed
	test.AssertInDB(t, tt.DB, "entity_change_requests", test.Data{
		"id":          cr.ID,
		"status":      ds.EntityChangeRejected,
		"reviewer_id": admin.ID,
		"reviewed_at": test.NotNull,
		"review_note": req.Note,
		"updated_at":  test.NotNull,
	})

	// email should be sent
	emailVars := test.LoadEmailVars(t, user.Email)
	assert.Len(t, emailVars, 4)
	assert.Equal(t, user.Username, emailVars["username"])
	assert.Equal(t, book.Title, emailVars["entity_title"])
	assert.Equal(t, app.ServerURL(book.ViewURL()), emailVars["view_url"])
	assert.Equal(t, req.Note, emailVars["note"])
}

func TestApplyChangeRequestToBook(t *testing.T) {
	admin := create[ds.User](t)
	loginAs(t, admin)
	makeAdmin(admin)

	imageBytes, err := random.ImagePNG(10)
	test.CheckErr(t, err)
	cover := UploadFile(t, fileForm{
		purpose:  ds.FilePurposeBookCover,
		filename: "cover.jpg",
		file:     bytes.NewReader(imageBytes),
	})

	newImageBytes, err := random.ImagePNG(10)
	test.CheckErr(t, err)
	newCover := UploadFile(t, fileForm{
		purpose:  ds.FilePurposeBookCover,
		filename: "cover.jpg",
		file:     bytes.NewReader(newImageBytes),
	})

	topic := create[ds.Topic](t)
	user := create[ds.User](t)
	book := create(t, ds.Book{
		Entity: &ds.Entity{
			Topics: []ds.Topic{*topic},
		},
		CoverFileID: cover.ID,
	})
	newTopic := create(t, ds.Topic{
		Type: ds.EntityTypeBook,
	})

	summaryMD := random.String()
	summaryHTML, err := app.MarkdownToHTML(summaryMD)
	test.CheckErr(t, err)

	descriptionMD := random.String()
	descriptionHTML, err := app.MarkdownToHTML(descriptionMD)
	test.CheckErr(t, err)

	cr := create(t, ds.EntityChangeRequest{
		EntityID: book.ID,
		UserID:   user.ID,
		Status:   ds.EntityChangePending,
		Diff: map[string]any{
			"title":         random.String(),
			"summary":       summaryMD,
			"description":   descriptionMD,
			"cover_file_id": newCover.ID,
			"homepage":      random.String(),
			"release_date":  random.ReleaseDate(),
			"topics":        []string{newTopic.PublicID},
			"authors":       []ds.BookAuthor{{Name: random.String(), Link: random.String()}},
		},
	})

	var resp response.Status
	UPDATE(t, pf("change-requests/%s/apply/", cr.ID), struct{}{}, &resp)

	// entity should be updated
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":          book.ID,
		"title":       cr.Diff["title"],
		"summary":     summaryHTML,
		"summary_raw": summaryMD,
	})

	// book should be updated
	test.AssertInDB(t, tt.DB, "books", test.Data{
		"id":              book.ID,
		"description":     descriptionHTML,
		"description_raw": descriptionMD,
		"cover_file_id":   cr.Diff["cover_file_id"],
		"homepage":        cr.Diff["homepage"],
		"release_date":    cr.Diff["release_date"],
		"authors":         cr.Diff["authors"],
	})

	// old cover should be deleted
	test.AssertInDB(t, tt.DB, "files", test.Data{
		"id":         cover.ID,
		"deleted_at": test.NotNull,
	})

	// new cover should not be temporary
	test.AssertInDB(t, tt.DB, "files", test.Data{
		"id":   newCover.ID,
		"temp": false,
	})

	// old topic should be detached
	test.AssertNotInDB(t, tt.DB, "entity_topics", test.Data{
		"entity_id": book.ID,
		"topic_id":  topic.ID,
	})
	// new topic should be attached
	test.AssertInDB(t, tt.DB, "entity_topics", test.Data{
		"entity_id": book.ID,
		"topic_id":  newTopic.ID,
	})

	// change request status should be changed
	test.AssertInDB(t, tt.DB, "entity_change_requests", test.Data{
		"id":          cr.ID,
		"status":      ds.EntityChangeCommitted,
		"reviewer_id": admin.ID,
		"reviewed_at": test.NotNull,
		"updated_at":  test.NotNull,
	})

	// log should be added
	test.AssertInDB(t, tt.DB, "event_logs", test.Data{
		"user_id":   user.ID,
		"type":      ds.EventLogEntityUpdated,
		"entity_id": book.ID,
		"is_public": true,
	})

	// email should be sent
	emailVars := test.LoadEmailVars(t, user.Email)
	assert.Len(t, emailVars, 3)
	assert.Equal(t, user.Username, emailVars["username"])
	assert.Equal(t, book.Title, emailVars["entity_title"])
	assert.Equal(t, app.ServerURL("/books/"+book.PublicID), emailVars["view_url"])
}

func TestApplyChangeRequestToPage(t *testing.T) {
	admin := create[ds.User](t)
	loginAs(t, admin)
	makeAdmin(admin)

	user := create[ds.User](t)

	contentMD := random.String()
	contentHTML, err := app.MarkdownToHTML(contentMD)
	test.CheckErr(t, err)

	page := create(t, ds.Page{
		Entity: &ds.Entity{
			Title:    random.String(),
			PublicID: random.String(),
		},
		ContentRaw: contentMD,
		Content:    contentHTML,
	})

	// change request values
	newPublicID := random.String()
	newTitle := random.String()
	newContentMD := random.String()
	newContentHTML, err := app.MarkdownToHTML(newContentMD)
	test.CheckErr(t, err)

	cr := create(t, ds.EntityChangeRequest{
		EntityID: page.ID,
		UserID:   user.ID,
		Status:   ds.EntityChangePending,
		Diff: map[string]any{
			"public_id": newPublicID,
			"title":     newTitle,
			"content":   newContentMD,
		},
	})

	var resp response.Status
	UPDATE(t, pf("change-requests/%s/apply/", cr.ID), struct{}{}, &resp)

	// entity should be updated
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":         page.ID,
		"title":      newTitle,
		"public_id":  newPublicID,
		"updated_at": test.NotNull,
	})

	// page should be updated
	test.AssertInDB(t, tt.DB, "pages", test.Data{
		"id":          page.ID,
		"content_raw": newContentMD,
		"content":     newContentHTML,
	})

	// change request status should be changed
	test.AssertInDB(t, tt.DB, "entity_change_requests", test.Data{
		"id":          cr.ID,
		"status":      ds.EntityChangeCommitted,
		"reviewer_id": admin.ID,
		"reviewed_at": test.NotNull,
		"updated_at":  test.NotNull,
	})

	// log should be added
	test.AssertInDB(t, tt.DB, "event_logs", test.Data{
		"user_id":   user.ID,
		"type":      ds.EventLogEntityUpdated,
		"entity_id": page.ID,
		"is_public": true,
	})

	// email should be sent
	emailVars := test.LoadEmailVars(t, user.Email)
	assert.Len(t, emailVars, 3)
	assert.Equal(t, user.Username, emailVars["username"])
	assert.Equal(t, page.Title, emailVars["entity_title"])
	assert.Equal(t, app.ServerURL(newPublicID), emailVars["view_url"])
}
