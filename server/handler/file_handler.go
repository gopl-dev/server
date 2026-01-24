package handler

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
	"github.com/gopl-dev/server/file"
)

// UploadFile is a handler for file upload.
//
//	@ID			UploadFile
//	@Summary	Upload file
//	@Tags		files
//	@Accept		mpfd
//	@Produce	json
//	@Param		file	body		[]byte	true	"File"
//	@Param		purpose	body		string	true	"File purpose"
//	@Success	201		{object}	response.Status
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/files/ [post]
//	@Security	ApiKeyAuth
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "UploadFile")
	defer span.End()

	user := ds.UserFromContext(ctx)
	if user == nil {
		Abort(w, r, app.ErrUnauthorized())
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 10<<20) //nolint:mnd

	maxSize := app.Config().Files.MaxUploadSizeMB
	err := r.ParseMultipartForm(maxSize << 20) //nolint:mnd
	if err != nil {
		if errors.Is(err, multipart.ErrMessageTooLarge) {
			Abort(w, r, app.ErrBadRequest("File too large. Max is %dMB", maxSize))
			return
		}
		Abort(w, r, app.ErrBadRequest("Invalid multipart form: %v", err))
		return
	}

	src, header, err := r.FormFile("file")
	if err != nil {
		Abort(w, r, app.ErrBadRequest("Invalid file"))
		return
	}
	defer func() {
		closeErr := src.Close()
		if closeErr != nil {
			span.RecordError(closeErr)
		}
	}()

	f, err := h.service.UploadFile(ctx, service.UploadFileArgs{
		Name:    header.Filename,
		OwnerID: user.ID,
		Purpose: ds.FilePurpose(r.FormValue("purpose")),
		Temp:    true,
		File:    src,
	})
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonCreated(w, f)
}

// DeleteFile handles file deletion.
//
//	@ID			DeleteFile
//	@Summary	Delete file
//	@Tags		files
//	@Accept		json
//	@Produce	json
//	@Param		id	path		string	true	"File ID"
//	@Success	200		{object}	response.Status
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/files/{id}/ [delete]
//	@Security	ApiKeyAuth
func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "DeleteFile")
	defer span.End()

	id, err := idFromPath(r)
	if err != nil {
		Abort(w, r, app.ErrBadRequest("Invalid file ID"))
		return
	}

	err = h.service.DeleteFile(ctx, id)
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonSuccess(w)
}

// GetFileMetadata returns file metadata.
//
//	@ID			GetFileMetadata
//	@Summary	Get file metadata
//	@Tags		files
//	@Accept		json
//	@Produce	json
//	@Param		id	path		string	true	"File ID"
//	@Success	200		{object}	ds.File
//	@Failure	422		{object}	Error
//	@Failure	500		{object}	Error
//	@Router		/files/{id}/ [post]
//	@Security	ApiKeyAuth
func (h *Handler) GetFileMetadata(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "GetFileMetadata")
	defer span.End()

	id, err := idFromPath(r)
	if err != nil {
		Abort(w, r, err)
		return
	}

	f, err := h.service.GetFileByID(ctx, id)
	if err != nil {
		Abort(w, r, err)
		return
	}

	jsonOK(w, f)
}

// DownloadFile serves the file content.
/*
func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "DownloadFile")
	defer span.End()

	id, err := idFromPath(r)
	if err != nil {
		Abort(w, err)
		return
	}

	f, err := h.service.GetFileByID(ctx, id)
	if err != nil {
		Abort(w, err)
		return
	}

	// Serve file
	http.ServeFile(w, r, f.Path)
}
*/

// RenderFile serves the file content.
func (h *Handler) RenderFile(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "RenderFile")
	defer span.End()

	id, err := idFromPath(r)
	if err != nil {
		Abort(w, r, err)
		return
	}

	f, err := h.service.GetFileByID(ctx, id)
	if err != nil {
		Abort(w, r, err)
		return
	}

	filePath := f.Path

	etag := f.Hash
	_, isPreview := r.URL.Query()["preview"]
	if isPreview {
		etag += "-preview"
	}

	if inm := r.Header.Get("If-None-Match"); inm != "" && inm == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if ims := r.Header.Get("If-Modified-Since"); ims != "" {
		t, err := time.Parse(http.TimeFormat, ims)
		if err == nil {
			if !f.CreatedAt.After(t) {
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	var fh file.ReadSeekCloser
	var size int64
	if isPreview {
		fh, size, err = h.service.GetFilePreview(ctx, f)
	} else {
		fh, size, err = file.Open(ctx, filePath)
	}
	if err != nil {
		Abort(w, r, err)
		return
	}
	defer func() {
		closeErr := fh.Close()
		if closeErr != nil {
			span.RecordError(closeErr)
		}
	}()

	w.Header().Set("ETag", `"`+etag+`"`)
	w.Header().Set("Last-Modified", f.CreatedAt.UTC().Format(http.TimeFormat))
	w.Header().Set("Content-Type", f.MimeType)
	w.Header().Set("Content-Length", strconv.FormatInt(size, 10))

	// w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, f.Name))

	http.ServeContent(w, r, f.Name, f.CreatedAt, fh)
}
