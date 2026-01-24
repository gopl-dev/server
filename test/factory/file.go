package factory

import (
	"context"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/file"
	"github.com/gopl-dev/server/test/factory/random"
)

// NewFile ...
func (f *Factory) NewFile(overrideOpt ...ds.File) (m *ds.File) {
	m = &ds.File{
		ID:          ds.NewID(),
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
func (f *Factory) CreateFile(overrideOpt ...ds.File) (m *ds.File, err error) {
	m = f.NewFile(overrideOpt...)

	if m.ID.IsNil() {
		m.ID = ds.NewID()
	}

	if m.OwnerID.IsNil() {
		u, err := f.CreateUser()
		if err != nil {
			return nil, err
		}

		m.OwnerID = u.ID
	}

	err = f.repo.CreateFile(context.Background(), m)
	return
}
