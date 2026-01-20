package api_test

import (
	"bytes"
	"fmt"
	"net/http"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/file"
	"github.com/gopl-dev/server/server/handler"
	"github.com/gopl-dev/server/server/response"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory/random"
)

// TestUploadFile is a minimal effort to create valid file.
func TestUploadFile(t *testing.T) {
	user := login(t)

	imageBytes, err := random.ImagePNG()
	test.CheckErr(t, err)

	req := fileForm{
		purpose:  ds.FilePurposeBookCover,
		filename: "cover.jpg",
		file:     bytes.NewReader(imageBytes),
	}

	resp := UploadFile(t, req)

	assert.True(t, resp.Size > 0)
	assert.Equal(t, resp.MimeType, "image/png")

	test.AssertInDB(t, tt.DB, "files", test.Data{
		"id":        resp.ID,
		"owner_id":  user.ID,
		"purpose":   req.purpose,
		"size":      resp.Size,
		"type":      file.TypeImage,
		"mime_type": resp.MimeType,
		"temp":      true,
	})
}

func TestGetFileMetadata(t *testing.T) {
	f := tt.Factory.CreateFile(t)
	var resp ds.File
	path := fmt.Sprintf("/files/%s/", f.ID)
	GET(t, path, &resp)

	assert.Equal(t, resp.ID, f.ID)
	assert.Equal(t, resp.OwnerID, f.OwnerID)
	assert.Equal(t, resp.Name, f.Name)
	assert.Equal(t, resp.Type, f.Type)
	assert.Equal(t, resp.MimeType, f.MimeType)
	assert.Equal(t, resp.Purpose, f.Purpose)
	assert.Equal(t, resp.Size, f.Size)
}

func TestDeleteFile(t *testing.T) {
	_ = login(t)

	imageBytes, err := random.ImagePNG()
	test.CheckErr(t, err)

	fileUploadForm := fileForm{
		purpose:  ds.FilePurposeBookCover,
		filename: "cover.jpg",
		file:     bytes.NewReader(imageBytes),
	}

	f := UploadFile(t, fileUploadForm)
	var resp response.Status
	path := fmt.Sprintf("/files/%s/", f.ID)
	DELETE(t, path, &resp)

	test.AssertInDB(t, tt.DB, "files", test.Data{
		"id":         f.ID,
		"deleted_at": test.NotNull,
	})

	t.Run("ownership violation", func(t *testing.T) {
		f := tt.Factory.CreateFile(t) // file belongs to another (new) user
		path := fmt.Sprintf("/files/%s/", f.ID)
		var resp handler.Error
		Request(t, RequestArgs{
			path:         path,
			bindResponse: &resp,
			assertStatus: http.StatusForbidden,
			method:       http.MethodDelete,
		})
	})
}
