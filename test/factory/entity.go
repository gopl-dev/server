//nolint:mnd
package factory

import (
	"context"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test/factory/random"
)

// NewEntity ...
func (f *Factory) NewEntity(overrideOpt ...ds.Entity) (m *ds.Entity) {
	createdAt := fake.DateRange(time.Now().AddDate(0, -12, 0), time.Now())
	var publishedAt, updatedAt, deletedAt *time.Time

	status := random.Element(ds.EntityStatuses)
	if status == ds.EntityStatusApproved {
		publishedAt = &createdAt
		updatedAt = random.ValOrNil(fake.DateRange(createdAt.AddDate(0, -12, -25), createdAt), 50)
		deletedAt = random.ValOrNil(fake.DateRange(createdAt.AddDate(0, -12, -25), createdAt), 25)
	}

	m = &ds.Entity{
		ID:            ds.NewID(),
		PublicID:      fake.UrlSlug(random.Int(3, 5)),
		OwnerID:       ds.NilID,
		PreviewFileID: ds.NilID,
		Type:          random.Element(ds.EntityTypes),
		Title:         fake.BookTitle(),
		Visibility:    random.Element(ds.EntityVisibilities),
		Status:        status,
		PublishedAt:   publishedAt,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		DeletedAt:     deletedAt,
	}

	if len(overrideOpt) == 1 {
		merge(m, overrideOpt[0])
	}

	return
}

// CreateEntity ...
func (f *Factory) CreateEntity(overrideOpt ...ds.Entity) (m *ds.Entity, err error) {
	m = f.NewEntity(overrideOpt...)

	err = f.repo.CreateEntity(context.Background(), m)
	return
}
