package factory

import (
	"context"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/gopl-dev/server/app/ds"
)

// NewBook ...
func (f *Factory) NewBook(overrideOpt ...ds.Book) (m *ds.Book) {
	m = &ds.Book{
		Entity:      f.NewEntity(),
		CoverFileID: ds.NilID,
		Description: fake.Paragraph(),
		AuthorName:  fake.BookAuthor(),
		AuthorLink:  fake.URL(),
		Homepage:    fake.URL(),
		ReleaseDate: fake.Date().Format("2006-01-02"),
	}

	if len(overrideOpt) == 1 {
		o := overrideOpt[0]
		merge(m, o)

		if o.Entity == nil {
			o.Entity = &ds.Entity{}
		} else {
			merge(m.Entity, o.Entity)
		}
	}

	return
}

// CreateBook ...
func (f *Factory) CreateBook(overrideOpt ...ds.Book) (m *ds.Book, err error) {
	m = f.NewBook(overrideOpt...)

	m.Entity, err = f.CreateEntity(*m.Entity)
	if err != nil {
		return
	}

	err = f.repo.CreateBook(context.Background(), m)
	return
}
