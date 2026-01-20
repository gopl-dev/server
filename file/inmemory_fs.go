package file

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
)

// InMemoryFSName is the driver name for the in-memory storage implementation.
const InMemoryFSName = "in-memory-fs"

var (
	// ErrInvalidFilename is returned when an empty or invalid filename is provided.
	ErrInvalidFilename = errors.New("invalid filename")

	// ErrFileNotFound is returned when the requested file does not exist in storage.
	ErrFileNotFound = errors.New("file not found")
)

// memReadSeekCloser wraps bytes.Reader to satisfy ReadSeekCloser.
// Close is a no-op because the underlying data is memory-backed.
type memReadSeekCloser struct{ *bytes.Reader }

// Close implements io.Closer. It is a no-op for in-memory readers.
func (m memReadSeekCloser) Close() error { return nil }

// InMemoryFSStorage implements Driver using an in-memory map.
// It is intended for tests and ephemeral storage, not for persistence.
type InMemoryFSStorage struct {
	storage sync.Map // map[string][]byte
}

// NewInMemoryFSStorage constructs a new in-memory storage driver.
func NewInMemoryFSStorage() (Driver, error) { return &InMemoryFSStorage{}, nil }

// Store reads all data from r and stores it in memory under the given filename.
func (s *InMemoryFSStorage) Store(_ context.Context, r io.Reader, filename string) (string, error) {
	if filename == "" {
		return "", ErrInvalidFilename
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}

	s.storage.Store(filename, b)
	return filename, nil
}

// Load returns the full contents of the stored file as a byte slice.
func (s *InMemoryFSStorage) Load(_ context.Context, filename string) ([]byte, error) {
	if filename == "" {
		return nil, ErrInvalidFilename
	}

	v, ok := s.storage.Load(filename)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrFileNotFound, filename)
	}

	return v.([]byte), nil //nolint:forcetypeassert
}

// Open opens a stored file for reading and seeking.
func (s *InMemoryFSStorage) Open(ctx context.Context, filename string) (ReadSeekCloser, int64, error) {
	b, err := s.Load(ctx, filename)
	if err != nil {
		return nil, 0, err
	}

	r := memReadSeekCloser{bytes.NewReader(b)}
	return r, int64(len(b)), nil
}

// Delete removes a file from in-memory storage.
func (s *InMemoryFSStorage) Delete(_ context.Context, filename string) error {
	if filename == "" {
		return ErrInvalidFilename
	}

	s.storage.Delete(filename)
	return nil
}
