package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app/ds"
)

var (
	// ErrBookNotFound is a sentinel error returned when book not found.
	ErrBookNotFound = errors.New("book not found")
)

// CreateBook inserts a new book record into the database.
// The corresponding entity in the 'entities' table should be created separately.
func (r *Repo) CreateBook(ctx context.Context, b *ds.Book) error {
	_, span := r.tracer.Start(ctx, "CreateBook")
	defer span.End()

	return r.insert(ctx, "books", data{
		"id":           b.ID,
		"description":  b.Description,
		"author_name":  b.AuthorName,
		"author_link":  b.AuthorLink,
		"homepage":     b.Homepage,
		"release_date": b.ReleaseDate,
		"cover_image":  b.CoverImage,
	})
}

// FindBookByURLName retrieves a book by its URL-friendly name.
func (r *Repo) FindBookByURLName(ctx context.Context, urlName string) (*ds.Book, error) {
	_, span := r.tracer.Start(ctx, "FindBookByURLName")
	defer span.End()

	book := new(ds.Book)
	const query = `
		SELECT * FROM entities e
		JOIN books b ON e.id = b.id
		WHERE e.url_name = $1 AND e.deleted_at IS NULL`

	db := r.getDB(ctx)
	err := pgxscan.Get(ctx, db, book, query, urlName)
	if noRows(err) {
		return nil, ErrBookNotFound
	}

	return book, err
}

// UpdateBook updates both entity and book tables.
func (r *Repo) UpdateBook(ctx context.Context, b *ds.Book) error {
	_, span := r.tracer.Start(ctx, "UpdateBook")
	defer span.End()

	const updateEntitySQL = `
		UPDATE entities 
		SET url_name = $1, title = $2, visibility = $3, updated_at = NOW() 
		WHERE id = $4`

	const updateBookSQL = `
		UPDATE books 
		SET description = $1, author_name = $2 
		WHERE id = $3`

	err := r.exec(ctx, updateEntitySQL, b.URLName, b.Title, b.Visibility, b.ID)
	if err != nil {
		return fmt.Errorf("update entity: %w", err)
	}

	err = r.exec(ctx, updateBookSQL, b.Description, b.AuthorName)
	if err != nil {
		return fmt.Errorf("update book: %w", err)
	}

	return nil
}
