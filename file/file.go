package file

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/lithammer/shortuuid"
)

// ReadSeekCloser is a convenience interface combining io.Reader, io.Seeker,
// and io.Closer.
type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

// ByteSize represents a size in bytes with helper conversion and formatting methods.
type ByteSize uint64

const (
	// B base byte units.
	B ByteSize = 1
	// KB ...
	KB = B << 10
	// MB ...
	MB = KB << 10
	// GB ...
	GB = MB << 10
)

var (
	// storage holds the initialized storage driver instance.
	storage Driver

	// initOnce guarantees storage driver initialization happens only once.
	initOnce sync.Once
)

// Storage returns the initialized storage driver.
// The driver is selected based on application configuration and
// initialized lazily on first access.
func Storage() Driver {
	initOnce.Do(func() {
		var err error

		name := app.Config().Files.StorageDriver
		initStorage, ok := drivers[name]
		if !ok {
			panic("unknown storage driver: " + name)
		}

		storage, err = initStorage()
		if err != nil {
			panic("init storage: " + err.Error())
		}
	})

	return storage
}

// Store writes data from r into the configured storage backend under filename.
func Store(ctx context.Context, r io.Reader, filename string) (string, error) {
	return Storage().Store(ctx, r, filename)
}

// Open opens a stored file for reading and seeking.
func Open(ctx context.Context, filename string) (ReadSeekCloser, int64, error) {
	return Storage().Open(ctx, filename)
}

// Load reads the entire stored file into memory and returns its contents.
func Load(ctx context.Context, filename string) ([]byte, error) {
	return Storage().Load(ctx, filename)
}

// Delete removes a file from the configured storage backend.
func Delete(ctx context.Context, filename string) error {
	return Storage().Delete(ctx, filename)
}

// SafeName generates a collision-resistant filename using the current timestamp
// and a short UUID, preserving the extension from provided filename.
func SafeName(filename string) string {
	return time.Now().Format("20060102_150405_") + shortuuid.New() + filepath.Ext(filename)
}

// Bytes returns the raw size in bytes.
func (b ByteSize) Bytes() uint64 {
	return uint64(b)
}

// KBytes returns the size in kilobytes (base 1024) as a float.
func (b ByteSize) KBytes() float64 {
	v := b / KB
	r := b % KB
	return float64(v) + float64(r)/float64(KB)
}

// MBytes returns the size in megabytes (base 1024) as a float.
func (b ByteSize) MBytes() float64 {
	v := b / MB
	r := b % MB
	return float64(v) + float64(r)/float64(MB)
}

// GBytes returns the size in gigabytes (base 1024) as a float.
func (b ByteSize) GBytes() float64 {
	v := b / GB
	r := b % GB
	return float64(v) + float64(r)/float64(GB)
}

// String returns a human-readable representation of the byte size
// using binary (IEC-style) units.
func (b ByteSize) String() string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}
