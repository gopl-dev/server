package service

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/file"
)

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
