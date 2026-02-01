package factory

import (
	"context"
	"time"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test/factory/random"
)

// NewTopic creates a new Topic model instance populated with default
// randomly generated data.
func (f *Factory) NewTopic(overrideOpt ...ds.Topic) (m *ds.Topic) {
	m = &ds.Topic{
		ID:          ds.NewID(),
		Type:        random.Element(ds.EntityTypes),
		PublicID:    fake.UrlSlug(3), //nolint:mnd
		Name:        fake.MovieGenre(),
		Description: fake.MovieName(),
		CreatedAt:   time.Now(),
		UpdatedAt:   nil,
		DeletedAt:   nil,
	}

	if len(overrideOpt) == 1 {
		merge(m, overrideOpt[0])
	}

	return
}

// CreateTopic creates and persists a new Topic record in the repository.
func (f *Factory) CreateTopic(overrideOpt ...ds.Topic) (m *ds.Topic, err error) {
	m = f.NewTopic(overrideOpt...)

	err = f.repo.CreateTopic(context.Background(), m)
	return
}
