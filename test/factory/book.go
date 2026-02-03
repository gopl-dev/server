package factory

import (
	"context"
	"strings"

	fake "github.com/brianvoe/gofakeit/v7"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test/factory/random"
)

// NewBook creates a new Book model populated with fake data.
func (f *Factory) NewBook(overrideOpt ...ds.Book) (m *ds.Book) {
	text := strings.Repeat(fake.Paragraph(), 5) //nolint:mnd

	m = &ds.Book{
		Entity:         f.NewEntity(),
		CoverFileID:    ds.NilID,
		Authors:        NewBookAuthors(),
		Homepage:       fake.URL(),
		ReleaseDate:    random.ReleaseDate(),
		Description:    text,
		DescriptionRaw: text,
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

// CreateBook creates and persists a new Book record in the repository.
func (f *Factory) CreateBook(overrideOpt ...ds.Book) (m *ds.Book, err error) {
	m = f.NewBook(overrideOpt...)

	m.Type = ds.EntityTypeBook
	m.Entity, err = f.CreateEntity(*m.Entity)
	if err != nil {
		return
	}

	err = f.repo.CreateBook(context.Background(), m)
	return
}

// NewBookAuthor creates a new BookAuthor model populated with fake data.
func NewBookAuthor() ds.BookAuthor {
	return ds.BookAuthor{
		Name: random.Element([]string{
			fake.BookAuthor(),
			fake.CelebrityActor(),
			fake.CelebrityBusiness(),
		}),
		Link: random.Maybe(fake.URL()),
	}
}

// NewBookAuthors generates and returns a slice of BookAuthor values.
func NewBookAuthors(countOpt ...int) []ds.BookAuthor {
	count := random.Int(1, 3) //nolint:mnd
	if len(countOpt) == 1 {
		count = countOpt[0]
	}

	authors := make([]ds.BookAuthor, count)
	for i := range authors {
		authors[i] = NewBookAuthor()
	}

	return authors
}
