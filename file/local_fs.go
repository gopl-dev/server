package file

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gopl-dev/server/app"
)

// LocalFSName is the driver name used to select the local filesystem storage implementation.
const LocalFSName = "local-fs"

var (
	// ErrStoragePathNotSet indicates missing local storage base directory in config.
	ErrStoragePathNotSet = errors.New("[local-fs] storage_path is not set")

	// ErrInitStorage is a generic error prefix used for wrapping initialization failures.
	ErrInitStorage = errors.New("[local-fs] create driver")
)

// LocalFSStorage implements Driver using the local OS filesystem.
type LocalFSStorage struct {
	basePath string // absolute directory used as storage root
}

// NewLocalFSStorage constructs the local filesystem driver from config.
func NewLocalFSStorage() (Driver, error) {
	conf := app.Config().Files
	storagePath := conf.LocalFS.StoragePath

	if strings.TrimSpace(storagePath) == "" {
		return nil, ErrStoragePathNotSet
	}

	base, err := filepath.Abs(storagePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInitStorage, err)
	}
	base = filepath.Clean(base)

	err = os.MkdirAll(base, 0o750) //nolint:mnd
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInitStorage, err)
	}

	info, err := os.Stat(base)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInitStorage, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: not a directory: %s", ErrInitStorage, base)
	}

	return &LocalFSStorage{basePath: base}, nil
}

// Store writes content from r into the local storage under filename.
// For local FS, Store must be atomic internally (tmp + rename).
func (s *LocalFSStorage) Store(_ context.Context, r io.Reader, filename string) (string, error) {
	rel, full, err := s.fullpath(filename)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(full)
	err = os.MkdirAll(dir, 0o750) //nolint:mnd
	if err != nil {
		return "", err
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return "", err
	}
	tmpName := tmp.Name()

	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}()

	_, err = io.Copy(tmp, r)
	if err != nil {
		return "", err
	}

	// Ensure data is flushed to disk before moving into place.
	err = tmp.Sync()
	if err != nil {
		return "", err
	}
	err = tmp.Close()
	if err != nil {
		return "", err
	}

	err = renameReplace(tmpName, full)
	if err != nil {
		return "", err
	}

	return rel, nil
}

// Open opens a stored file for reading and seeking.
// It returns an os.File (implements ReadSeekCloser), the file size, and an error.
//
// Caller is responsible for closing the returned reader.
func (s *LocalFSStorage) Open(_ context.Context, filename string) (ReadSeekCloser, int64, error) {
	_, full, err := s.fullpath(filename)
	if err != nil {
		return nil, 0, err
	}

	f, err := os.Open(full) //nolint:gosec
	if err != nil {
		return nil, 0, err
	}

	fi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, 0, err
	}

	return f, fi.Size(), nil
}

// Load reads the entire stored file into memory and returns its bytes.
// Prefer Open for streaming large files.
func (s *LocalFSStorage) Load(_ context.Context, filename string) ([]byte, error) {
	_, full, err := s.fullpath(filename)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(full) //nolint:gosec
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Delete removes a stored file.
// It is not an error if the file does not exist.
func (s *LocalFSStorage) Delete(_ context.Context, filename string) error {
	_, full, err := s.fullpath(filename)
	if err != nil {
		return err
	}

	err = os.Remove(full)
	if os.IsNotExist(err) {
		return nil
	}

	return err
}

// fullpath converts a filename to:
//   - rel: normalized relative storage key
//   - full: absolute OS path under s.basePath
func (s *LocalFSStorage) fullpath(filename string) (string, string, error) {
	rel := normalizeFilepath(filename)
	if rel == "" || rel == "." {
		return "", "", fmt.Errorf("%w: %s", ErrInvalidFilename, filename)
	}

	full := filepath.Join(s.basePath, rel)
	full = filepath.Clean(full)

	return rel, full, nil
}

// normalizeFilepath normalizes a filename into a consistent, safe relative path.
func normalizeFilepath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	path = filepath.ToSlash(filepath.Clean(path))

	if path == ".." {
		return ""
	}
	if strings.HasPrefix(path, "../") {
		return ""
	}

	return path
}

// renameReplace renames file from -> to.
// On platforms where os.Rename does not replace an existing destination (notably Windows),
// it emulates "replace" semantics by removing the destination and retrying.
func renameReplace(from, to string) error { // TODO review
	err := os.Rename(from, to)
	if err == nil {
		return nil
	}

	// If destination doesn't exist -> real rename error.
	_, statErr := os.Stat(to)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return err
		}
		return fmt.Errorf("stat destination: %w", statErr)
	}

	// Destination exists -> remove and retry (Windows-friendly behavior).
	rmErr := os.Remove(to)
	if rmErr != nil {
		return fmt.Errorf("remove destination: %w", rmErr)
	}

	err = os.Rename(from, to)
	if err != nil {
		return err
	}

	return nil
}
