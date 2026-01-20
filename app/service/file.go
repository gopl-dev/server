package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
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

var (
	// ErrPreviewUnavailable ...
	ErrPreviewUnavailable = errors.New("preview unavailable")
)

// UploadFileArgs ...
type UploadFileArgs struct {
	Name    string
	OwnerID ds.ID
	Purpose ds.FilePurpose
	Temp    bool
	File    file.ReadSeekCloser
}

// UploadFile ...
func (s *Service) UploadFile(ctx context.Context, args UploadFileArgs) (*ds.File, error) {
	ctx, span := s.tracer.Start(ctx, "UploadFile")
	defer span.End()

	if file.IsResizableImage(args.Name) {
		err := file.CheckImageDimensions(args.File)
		if err != nil {
			return nil, app.ErrUnprocessable(err.Error())
		}
	}

	buf := make([]byte, 512) //nolint:mnd
	n, err := io.ReadFull(args.File, buf)
	if errors.Is(err, io.ErrUnexpectedEOF) {
		err = nil
	}
	if err != nil {
		return nil, err
	}

	mimeType := http.DetectContentType(buf[:n])
	reader := io.MultiReader(bytes.NewReader(buf[:n]), args.File)

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

		if file.IsResizableImage(f.Path) {
			f.PreviewPath, err = file.CreatePreview(ctx, f.Path)
			if err != nil {
				return nil, err
			}
		}

		err = s.CreateFile(ctx, f)
		return f, err
	}

	// use existing file
	f.Name = existing.Name
	f.Path = existing.Path
	f.PreviewPath = existing.PreviewPath
	f.Type = existing.Type
	f.Purpose = existing.Purpose
	f.Size = existing.Size

	err = s.CreateFile(ctx, f)
	return f, err
}

// CreateFile ...
func (s *Service) CreateFile(ctx context.Context, f *ds.File) error {
	ctx, span := s.tracer.Start(ctx, "CreateFile")
	defer span.End()

	err := ValidateCreate(f)
	if err != nil {
		return err
	}

	return s.db.CreateFile(ctx, f)
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
		if !file.IsResizableImage(f.Path) {
			return app.InputError{"purpose": fmt.Sprintf("invalid file type for book cover, only %v types is accepted", file.ResizableImages)}
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

// GetFileByID retrieves a file by its ID.
func (s *Service) GetFileByID(ctx context.Context, id ds.ID) (*ds.File, error) {
	ctx, span := s.tracer.Start(ctx, "GetFileByID")
	defer span.End()

	return s.db.GetFileByID(ctx, id)
}

// GetFileByHash retrieves a file by its ID.
func (s *Service) GetFileByHash(ctx context.Context, hash string) (*ds.File, error) {
	ctx, span := s.tracer.Start(ctx, "GetFileByHash")
	defer span.End()

	return s.db.GetFileByHash(ctx, hash)
}

// UpdateFilePreviewByHash ...
func (s *Service) UpdateFilePreviewByHash(ctx context.Context, preview, hash string) error {
	ctx, span := s.tracer.Start(ctx, "UpdateFilePreviewByHash")
	defer span.End()

	return s.db.UpdateFilePreviewByHash(ctx, preview, hash)
}

// FilterFiles ...
func (s *Service) FilterFiles(ctx context.Context, f ds.FilesFilter) (data []ds.File, count int, err error) {
	ctx, span := s.tracer.Start(ctx, "FilterFiles")
	defer span.End()

	return s.db.FilterFiles(ctx, f)
}

// GetFilePreview ...
func (s *Service) GetFilePreview(ctx context.Context, f *ds.File) (fh file.ReadSeekCloser, size int64, err error) {
	ctx, span := s.tracer.Start(ctx, "GetFilePreview")
	defer span.End()

	if f.PreviewPath == "" {
		if !file.IsResizableImage(f.Path) {
			err = ErrPreviewUnavailable
			return
		}

		f.PreviewPath, err = file.CreatePreview(ctx, f.Path)
		if err != nil {
			err = fmt.Errorf("create preview: %w", err)
			return
		}

		err = s.UpdateFilePreviewByHash(ctx, f.PreviewPath, f.Hash)
		if err != nil {
			return
		}
	}

	return file.Open(ctx, f.PreviewPath)
}

// DeleteFile deletes a file.
func (s *Service) DeleteFile(ctx context.Context, id ds.ID) error {
	ctx, span := s.tracer.Start(ctx, "DeleteFile")
	defer span.End()

	user := ds.UserFromContext(ctx)
	if user == nil {
		return app.ErrUnauthorized()
	}

	f, err := s.db.GetFileByID(ctx, id)
	if err != nil {
		return err
	}

	if !f.IsOwner(user.ID) && !user.IsAdmin {
		return app.ErrForbidden("not owner")
	}

	err = s.db.DeleteFile(ctx, id)
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	// actual file from disk/storage will be deleted by worker
	return nil
}

// HardDeleteFileUnsafe permanently deletes a file.
// Note: This method does not check deletion prerequisites (e.g., permissions,
// ownership, or references) and is intended for internal use only
// (for example, background cleanup jobs).
// It must not be called from public API handlers.
// Hence the "Unsafe" suffix.
func (s *Service) HardDeleteFileUnsafe(ctx context.Context, f *ds.File) error {
	ctx, span := s.tracer.Start(ctx, "HardDeleteFile")
	defer span.End()

	err := s.db.HardDeleteFile(ctx, f.ID)
	if err != nil {
		return err
	}

	// If no one is using this file, it should be deleted from the filesystem.
	// If, for any reason, the file cannot be deleted, the error is ignored.
	_, err = s.GetFileByHash(ctx, f.Hash)
	if errors.Is(err, repo.ErrFileNotFound) {
		err = file.Delete(ctx, f.Path)
		if err != nil {
			log.Println("[ERROR] DELETE FILE: " + err.Error())
		}
		if f.PreviewPath != "" {
			err = file.Delete(ctx, f.PreviewPath)
			if err != nil {
				log.Println("[ERROR] DELETE PREVIEW FILE: " + err.Error())
			}
		}
	}

	return nil
}
