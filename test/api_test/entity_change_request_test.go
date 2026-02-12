package api_test

import (
	"bytes"
	"testing"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/diff"
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

	summaryPatch := random.Patch(book.Summary)
	cr := create(t, ds.EntityChangeRequest{
		EntityID: book.ID,
		UserID:   user.ID,
		Status:   ds.EntityChangePending,
		Diff: map[string]any{
			"summary": summaryPatch,
		},
	})

	var resp response.ChangeRequestDiff
	GET(t, pf("change-requests/%s/diff/", cr.ID), &resp)

	diffedSummary, err := diff.ComputeFromPatch(book.Summary, summaryPatch)
	test.CheckErr(t, err)

	assert.Len(t, resp.Diff, 1)
	assert.Equal(t, "summary", resp.Diff[0].Key)
	assert.Equal(t, diffedSummary.HTML(), resp.Diff[0].Diff)
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

	titlePatch := random.Patch(book.Title)
	summaryPatch := random.Patch(book.Summary)
	descriptionPatch := random.Patch(book.Description)
	homepagePatch := random.Patch(book.Homepage)
	releaseDatePatch := app.MakePatch(book.ReleaseDate, random.ReleaseDate())

	cr := create(t, ds.EntityChangeRequest{
		EntityID: book.ID,
		UserID:   user.ID,
		Status:   ds.EntityChangePending,
		Diff: map[string]any{
			"title":         titlePatch,
			"summary":       summaryPatch,
			"description":   descriptionPatch,
			"cover_file_id": newCover.ID,
			"homepage":      homepagePatch,
			"release_date":  releaseDatePatch,
			"topics":        []string{newTopic.PublicID},
			"authors":       []ds.BookAuthor{{Name: random.String(), Link: random.String()}},
		},
	})

	var resp response.Status
	UPDATE(t, pf("change-requests/%s/apply/", cr.ID), struct{}{}, &resp)

	patchedSummary, err := app.ApplyPatch(book.Summary, summaryPatch)
	test.CheckErr(t, err)
	newSummaryHTML, err := app.MarkdownToHTML(patchedSummary)
	test.CheckErr(t, err)

	patchedDescription, err := app.ApplyPatch(book.Description, descriptionPatch)
	test.CheckErr(t, err)
	newDescriptionHTML, err := app.MarkdownToHTML(patchedDescription)
	test.CheckErr(t, err)

	patchedTitle, err := app.ApplyPatch(book.Title, titlePatch)
	test.CheckErr(t, err)

	patchedHomepage, err := app.ApplyPatch(book.Homepage, homepagePatch)
	test.CheckErr(t, err)

	patchedReleaseDate, err := app.ApplyPatch(book.ReleaseDate, releaseDatePatch)
	test.CheckErr(t, err)

	// entity should be updated
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":          book.ID,
		"title":       patchedTitle,
		"summary":     newSummaryHTML,
		"summary_raw": patchedSummary,
	})

	// book should be updated
	test.AssertInDB(t, tt.DB, "books", test.Data{
		"id":              book.ID,
		"description":     newDescriptionHTML,
		"description_raw": patchedDescription,
		"cover_file_id":   cr.Diff["cover_file_id"],
		"homepage":        patchedHomepage,
		"release_date":    patchedReleaseDate,
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
		ContentRaw: contentMD,
		Content:    contentHTML,
	})

	// change request values
	titlePatch := random.Patch(page.Title)
	publicIDPatch := random.Patch(page.PublicID)
	contentPatch := random.Patch(page.ContentRaw)

	cr := create(t, ds.EntityChangeRequest{
		EntityID: page.ID,
		UserID:   user.ID,
		Status:   ds.EntityChangePending,
		Diff: map[string]any{
			"public_id": publicIDPatch,
			"title":     titlePatch,
			"content":   contentPatch,
		},
	})

	var resp response.Status
	UPDATE(t, pf("change-requests/%s/apply/", cr.ID), struct{}{}, &resp)

	patchedContent, err := app.ApplyPatch(page.ContentRaw, contentPatch)
	test.CheckErr(t, err)
	newContentHTML, err := app.MarkdownToHTML(patchedContent)
	test.CheckErr(t, err)

	patchedTitle, err := app.ApplyPatch(page.Title, titlePatch)
	test.CheckErr(t, err)
	patchedPublicID, err := app.ApplyPatch(page.PublicID, publicIDPatch)
	test.CheckErr(t, err)

	// entity should be updated
	test.AssertInDB(t, tt.DB, "entities", test.Data{
		"id":         page.ID,
		"title":      patchedTitle,
		"public_id":  patchedPublicID,
		"updated_at": test.NotNull,
	})

	// page should be updated
	test.AssertInDB(t, tt.DB, "pages", test.Data{
		"id":          page.ID,
		"content_raw": patchedContent,
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
	assert.Equal(t, app.ServerURL(patchedPublicID), emailVars["view_url"])
}
