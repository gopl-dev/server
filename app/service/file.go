package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/file"
)

var (
	// ErrPreviewUnavailable ...
	ErrPreviewUnavailable = errors.New("preview unavailable")
)

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
		if !file.CanDoPreview(f.Path) {
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
