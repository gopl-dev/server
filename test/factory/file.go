package factory

import (
	"context"
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/file"
	"github.com/gopl-dev/server/test/factory/random"
)

// NewFile ...
func (f *Factory) NewFile(overrideOpt ...ds.File) (m *ds.File) {
	m = &ds.File{
		ID:          ds.NilID,
		OwnerID:     ds.NilID,
		Name:        random.String(),
		Path:        random.String(),
		PreviewPath: random.String(),
		Hash:        random.String(),
		Type:        file.TypeImage,
		MimeType:    random.String(),
		Purpose:     ds.FilePurposeBookCover,
		Size:        1,
		CreatedAt:   time.Now(),
		DeletedAt:   nil,
		Temp:        false,
	}

	if len(overrideOpt) == 1 {
		merge(m, overrideOpt[0])
	}

	return
}

// CreateFile ...
func (f *Factory) CreateFile(t *testing.T, overrideOpt ...ds.File) (m *ds.File) {
	t.Helper()

	m = f.NewFile(overrideOpt...)

	if m.ID.IsNil() {
		m.ID = ds.NewID()
	}

	if m.OwnerID.IsNil() {
		m.OwnerID = f.CreateUser(t).ID
	}

	err := f.repo.CreateFile(context.Background(), m)
	checkErr(t, err)

	return
}
