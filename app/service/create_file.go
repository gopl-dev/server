package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/file"
	"github.com/lithammer/shortuuid"
)

// UploadFileArgs ...
type UploadFileArgs struct {
	Name    string
	OwnerID ds.ID
	Purpose ds.FilePurpose
	Temp    bool
	Reader  io.Reader
}

// UploadFile ...
func (s *Service) UploadFile(ctx context.Context, args UploadFileArgs) (*ds.File, error) {
	ctx, span := s.tracer.Start(ctx, "UploadFile")
	defer span.End()

	buf := make([]byte, 512) //nolint:mnd
	n, err := io.ReadFull(args.Reader, buf)
	if errors.Is(err, io.ErrUnexpectedEOF) {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	mimeType := http.DetectContentType(buf[:n])
	reader := io.MultiReader(bytes.NewReader(buf[:n]), args.Reader)

	tmp, err := os.CreateTemp("", ".temp-upload-*")
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = os.Remove(tmp.Name())
	}()

	hasher := sha256.New()
	w := io.MultiWriter(tmp, hasher)

	size, err := io.Copy(w, reader)
	if err != nil {
		_ = tmp.Close()
		return nil, err
	}

	// Close the file now to flush all buffers and release the handle.
	// (This is required before reopening the file, especially on Windows)
	err = tmp.Close()
	if err != nil {
		return nil, err
	}

	f := &ds.File{
		ID:          ds.NewID(),
		OwnerID:     args.OwnerID,
		Path:        "",
		PreviewPath: "",
		Hash:        hex.EncodeToString(hasher.Sum(nil)),
		Type:        "",
		MimeType:    mimeType,
		Purpose:     args.Purpose,
		Size:        size,
		CreatedAt:   time.Now(),
		DeletedAt:   nil,
		Temp:        args.Temp,
	}

	existing, err := s.db.GetFileByHash(ctx, f.Hash)
	if errors.Is(err, repo.ErrFileNotFound) {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	// new file
	if existing == nil {
		f.Name = args.Name
		if f.Name == "" {
			f.Name = shortuuid.New()
		}
		f.Path = file.SafeName(f.Name)
		f.PreviewPath = ""
		f.Type = file.ResolveFileType(args.Name)

		err = handleFilePurpose(f)
		if err != nil {
			return nil, err
		}

		src, err := os.Open(tmp.Name())
		if err != nil {
			return nil, err
		}
		defer func() {
			closeErr := src.Close()
			if closeErr != nil {
				span.RecordError(closeErr)
			}
		}()

		_, err = file.Store(ctx, src, f.Path)
		if err != nil {
			return nil, err
		}

		if file.CanDoPreview(f.Path) {
			f.PreviewPath, err = file.CreatePreview(ctx, f.Path)
			if err != nil {
				return nil, err
			}
		}

		err = s.db.CreateFile(ctx, f)
		return f, err
	}

	// use existing file
	f.Name = existing.Name
	f.Path = existing.Path
	f.PreviewPath = existing.PreviewPath
	f.Type = existing.Type
	f.Purpose = existing.Purpose
	f.Size = existing.Size

	err = s.db.CreateFile(ctx, f)
	return f, err
}

// handleFilePurpose validates the given file purpose and applies
// purpose-specific validation and normalization.
//
// The current implementation is intentionally simple.
// When the number of supported purposes grows, this logic
// will be decomposed into separate handlers per purpose.
func handleFilePurpose(f *ds.File) error {
	var subDir string
	switch f.Purpose {
	case ds.FilePurposeBookCover:
		if f.Type != file.TypeImage {
			return app.InputError{"purpose": "invalid file type for book cover"}
		}
		if !file.CanDoPreview(f.Path) {
			return app.InputError{"purpose": fmt.Sprintf("invalid file type for book cover, only %v types is accepted", file.PreviewValidExt)}
		}
		subDir = "book-covers"
	default:
		return app.InputError{"purpose": "invalid purpose"}
	}

	if subDir != "" {
		f.Path = filepath.Join(subDir, f.Path)
	}

	return nil
}
