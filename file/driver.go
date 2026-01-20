// Package file ...
package file

import (
	"context"
	"io"
)

// drivers maps storage driver names to their initialization functions.
// The selected driver is chosen at runtime based on configuration.
var drivers = map[string]func() (Driver, error){
	LocalFSName:    NewLocalFSStorage,
	InMemoryFSName: NewInMemoryFSStorage,
}

// Driver defines the common interface for file storage backends.
type Driver interface {
	// Store writes data from r into storage under the given filename.
	// It returns the normalized storage key used to reference the file.
	Store(ctx context.Context, r io.Reader, filename string) (string, error)

	// Open opens a stored file for reading and seeking.
	// It returns a ReadSeekCloser, the file size in bytes, and an error.
	Open(ctx context.Context, filename string) (fh ReadSeekCloser, size int64, err error)

	// Load reads the entire stored file into memory and returns its contents.
	// Prefer Open for large files.
	Load(ctx context.Context, filename string) ([]byte, error)

	// Delete removes a file from storage.
	// Deleting a non-existent file should not be considered an error.
	Delete(ctx context.Context, filename string) error
}
