//nolint:all
package service

import (
	"time"

	"github.com/gopl-dev/server/app/ds"
)

func (s *Service) CreateBook(u *ds.Book) (err error) {
	// err = database.ORM().Insert(u)
	return
}

func (s *Service) UpdateBook(u *ds.Book) (err error) {
	// err = database.ORM().Update(u)
	return
}

type FilterBooksParams struct {
	Limit  int
	Offset int
	Name   string
}

func (s *Service) FilterBooks(params FilterBooksParams) (data []ds.Book, count int, err error) {
	_ = params
	count = 5
	data = []ds.Book{
		{
			ID:          1,
			Title:       "Hello World",
			Description: "",
			CreatedAt:   time.Time{},
			UpdatedAt:   nil,
			DeletedAt:   nil,
		},
		{
			ID:          2,
			Title:       "Hello App",
			Description: "",
			CreatedAt:   time.Time{},
			UpdatedAt:   nil,
			DeletedAt:   nil,
		},
		{
			ID:          3,
			Title:       "Hello Server",
			Description: "",
			CreatedAt:   time.Time{},
			UpdatedAt:   nil,
			DeletedAt:   nil,
		},
		{
			ID:          4,
			Title:       "Hello Mobile",
			Description: "",
			CreatedAt:   time.Time{},
			UpdatedAt:   nil,
			DeletedAt:   nil,
		},
		{
			ID:          5,
			Title:       "Hello Web",
			Description: "",
			CreatedAt:   time.Time{},
			UpdatedAt:   nil,
			DeletedAt:   nil,
		},
	}

	return
}
