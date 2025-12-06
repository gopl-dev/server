package service

import "github.com/gopl-dev/server/app/ds"

func (s *Service) CreateBook(u *ds.Book) (err error) {
	//err = database.ORM().Insert(u)
	return
}

func (s *Service) UpdateBook(u *ds.Book) (err error) {
	//err = database.ORM().Update(u)
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
		{Title: "Hello World"},
		{Title: "Hello App"},
		{Title: "Hello Server"},
		{Title: "Hello Mobile"},
		{Title: "Hello Web"},
	}

	return
}
