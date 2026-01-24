package repo

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

var (
	// ErrFileNotFound is a sentinel error returned when file not found.
	ErrFileNotFound = app.ErrNotFound("file not found")
)

// CreateFile inserts a new file record into the database.
func (r *Repo) CreateFile(ctx context.Context, f *ds.File) error {
	ctx, span := r.tracer.Start(ctx, "CreateFile")
	defer span.End()

	err := r.insert(ctx, "files", data{
		"id":           f.ID,
		"owner_id":     f.OwnerID,
		"name":         f.Name,
		"path":         f.Path,
		"preview_path": f.PreviewPath,
		"hash":         f.Hash,
		"type":         f.Type,
		"mime_type":    f.MimeType,
		"purpose":      f.Purpose,
		"size":         f.Size,
		"created_at":   f.CreatedAt,
		"deleted_at":   f.DeletedAt,
		"temp":         f.Temp,
	})
	if err != nil {
		return fmt.Errorf("insert file: %w", err)
	}

	return nil
}

// GetFileByID retrieves a file by its ID.
func (r *Repo) GetFileByID(ctx context.Context, id ds.ID) (*ds.File, error) {
	ctx, span := r.tracer.Start(ctx, "GetFileByID")
	defer span.End()

	const query = `SELECT * FROM files WHERE id = $1 AND deleted_at IS NULL`

	file := new(ds.File)
	err := pgxscan.Get(ctx, r.getDB(ctx), file, query, id)
	if noRows(err) {
		return nil, ErrFileNotFound
	}

	return file, err
}

// GetFileByHash retrieves a file by its hash.
func (r *Repo) GetFileByHash(ctx context.Context, hash string) (*ds.File, error) {
	ctx, span := r.tracer.Start(ctx, "GetFileByHash")
	defer span.End()

	const query = `SELECT * FROM files WHERE hash = $1 AND deleted_at IS NULL LIMIT 1`

	file := new(ds.File)
	err := pgxscan.Get(ctx, r.getDB(ctx), file, query, hash)
	if noRows(err) {
		file = nil
		err = ErrFileNotFound
	}

	return file, err
}

// DeleteFile soft deletes a file.
func (r *Repo) DeleteFile(ctx context.Context, id ds.ID) error {
	ctx, span := r.tracer.Start(ctx, "DeleteFile")
	defer span.End()

	const query = `UPDATE files SET deleted_at = NOW() WHERE id = $1`

	err := r.exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	return nil
}

// UpdateFilePreviewByHash ...
func (r *Repo) UpdateFilePreviewByHash(ctx context.Context, preview, hash string) error {
	ctx, span := r.tracer.Start(ctx, "UpdateFilePreviewByHash")
	defer span.End()

	const query = `UPDATE files SET preview_path = $1 WHERE hash = $2`

	err := r.exec(ctx, query, preview, hash)
	if err != nil {
		return fmt.Errorf("update preview path: %w", err)
	}

	return nil
}

// CommitFile ...
func (r *Repo) CommitFile(ctx context.Context, fileID ds.ID) error {
	ctx, span := r.tracer.Start(ctx, "CommitFile")
	defer span.End()

	const query = `UPDATE files SET temp = FALSE WHERE id = $1`

	err := r.exec(ctx, query, fileID)
	if err != nil {
		return fmt.Errorf("commit file: %w", err)
	}

	return nil
}

// FilterFiles ...
func (r *Repo) FilterFiles(ctx context.Context, f ds.FilesFilter) (files []ds.File, count int, err error) {
	_, span := r.tracer.Start(ctx, "FilterUsers")
	defer span.End()

	count, err = r.filter("files").
		paginate(f.Page, f.PerPage).
		createdAt(f.CreatedAt).
		deletedAt(f.DeletedAt).
		deleted(f.Deleted).
		order(f.OrderBy, f.OrderDirection).
		withCount(f.WithCount).
		scan(ctx, &files)

	return
}

// HardDeleteFile permanently deletes a file record from the database.
// Epstein and his friends would be jealous.
func (r *Repo) HardDeleteFile(ctx context.Context, fileID ds.ID) (err error) {
	_, span := r.tracer.Start(ctx, "HardDeleteFile")
	defer span.End()

	return r.hardDelete(ctx, "files", fileID)
}
