// Package seed provides helpers for seeding test and development data.
//
// The package is intended for non-production use (tests, local development).
package seed

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"sync"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/file"
	"github.com/gopl-dev/server/test/factory"
	"github.com/gopl-dev/server/test/factory/random"
	"github.com/gopl-dev/server/tracing"
)

var (
	// ErrInvalidCount is returned when a count argument is less than or equal to zero.
	ErrInvalidCount = errors.New("count must be greater than zero")

	// ErrNoRows indicates that a query or operation returned no rows.
	ErrNoRows = errors.New("no rows")
)

// bucket returns an ID bucket for the given key, creating it if necessary.
//
// Buckets are used to lazily load and cache related entity IDs in a
// concurrency-safe manner.
type idBucket struct {
	once sync.Once
	ids  []ds.ID
	err  error
}

// Seed holds dependencies required by Seed methods.
type Seed struct {
	db      *app.DB
	repo    *repo.Repo
	factory *factory.Factory

	mu    sync.Mutex
	relID map[string]*idBucket
}

// New creates a Seed instance with all required dependencies initialized.
func New(db *app.DB) *Seed {
	t := tracing.NewNoOpTracer()
	r := repo.New(db, t)

	return &Seed{
		db:      db,
		repo:    r,
		factory: factory.New(db),
		relID:   make(map[string]*idBucket),
	}
}

// makeImageFile generates a random PNG image, stores it using the file storage,
// optionally creates a preview, and returns a ds.File model describing
// the stored file.
//
// The caller is responsible for setting all fields that are not directly
// related to the file blob (such as ownership and domain-specific metadata).
func makeImageFile(ctx context.Context, w, h int) (*ds.File, error) {
	img, err := random.ImageAbstractPNG(w, h)
	if err != nil {
		return nil, err
	}

	name := fake.UrlSlug(2) + ".png" //nolint:mnd

	hasher := sha256.New()
	r := bytes.NewReader(img)
	tee := io.TeeReader(r, hasher)

	path, err := file.Store(ctx, tee, file.SafeName(name))
	if err != nil {
		return nil, err
	}

	var previewPath string
	if file.IsResizableImage(path) {
		previewPath, err = file.CreatePreview(ctx, path)
		if err != nil {
			return nil, err
		}
	}

	f := &ds.File{
		ID:          ds.NewID(),
		Name:        name,
		Path:        path,
		PreviewPath: previewPath,
		Hash:        hex.EncodeToString(hasher.Sum(nil)),
		Size:        int64(len(img)),
		MimeType:    "image/png",

		// The caller is **responsible** for setting these fields.
		OwnerID:   ds.NilID,
		Type:      "",
		Purpose:   "",
		CreatedAt: time.Now(),
		DeletedAt: nil,
		Temp:      false,
	}

	return f, nil
}

// All seeds all supported entities in a predefined order.
func (s *Seed) All(ctx context.Context, count int) (err error) {
	if count < 1 {
		return fmt.Errorf("seed all: %w", ErrInvalidCount)
	}

	err = s.Users(ctx, count)
	if err != nil {
		return err
	}

	err = s.Books(ctx, count)
	if err != nil {
		return err
	}

	return nil
}

// RandomUserID returns a random user ID.
//
// On the first call, it loads up to 100 user IDs from the database and caches
// them for subsequent calls to avoid repeated queries.
func (s *Seed) RandomUserID(ctx context.Context) (ds.ID, error) {
	const q = `SELECT id FROM users ORDER BY RANDOM() LIMIT 100`
	return s.randomRelID(ctx, "user", q)
}

// randomRelID returns a random ID associated with the given relation key.
//
// On the first call for a specific key, it executes loadQuery to load and cache
// a set of IDs from the database. The query is executed exactly once per key,
// even under concurrent access.
//
// Subsequent calls return a random ID from the cached set.
func (s *Seed) randomRelID(ctx context.Context, key, loadQuery string) (ds.ID, error) {
	b := s.bucket(key)

	b.once.Do(func() {
		var ids []ds.ID
		b.err = pgxscan.Select(ctx, s.db, &ids, loadQuery)
		if b.err != nil {
			return
		}
		if len(ids) == 0 {
			b.err = fmt.Errorf("seed: %s: %w", key, ErrNoRows)
			return
		}
		b.ids = ids
	})

	if b.err != nil {
		return ds.NilID, b.err
	}

	return b.ids[rand.IntN(len(b.ids))], nil //nolint:gosec
}

// bucket returns an ID bucket associated with the given key.
func (s *Seed) bucket(key string) *idBucket {
	s.mu.Lock()
	defer s.mu.Unlock()

	b := s.relID[key]
	if b == nil {
		b = &idBucket{}
		s.relID[key] = b
	}

	return b
}
