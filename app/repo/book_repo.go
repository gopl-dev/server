package repo

import (
	"context"
	"fmt"

	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

var (
	// ErrBookNotFound is a sentinel error returned when book not found.
	ErrBookNotFound = app.ErrNotFound("book not found")
)

// CreateBook inserts a new book record into the database.
// The corresponding entity in the 'entities' table should be created separately.
func (r *Repo) CreateBook(ctx context.Context, b *ds.Book) error {
	_, span := r.tracer.Start(ctx, "CreateBook")
	defer span.End()

	return r.insert(ctx, "books", data{
		"id":            b.ID,
		"cover_file_id": b.CoverFileID,
		"description":   b.Description,
		"author_name":   b.AuthorName,
		"author_link":   b.AuthorLink,
		"homepage":      b.Homepage,
		"release_date":  b.ReleaseDate,
	})
}

// GetBookByID retrieves a book by its ID.
func (r *Repo) GetBookByID(ctx context.Context, id ds.ID) (*ds.Book, error) {
	_, span := r.tracer.Start(ctx, "GetBookByID")
	defer span.End()

	book := new(ds.Book)
	const query = `
		SELECT * FROM entities e
		JOIN books b ON e.id = b.id
		WHERE e.id = $1 AND e.deleted_at IS NULL`

	err := pgxscan.Get(ctx, r.getDB(ctx), book, query, id)
	if noRows(err) {
		return nil, ErrBookNotFound
	}

	return book, err
}

// GetBookByPublicID retrieves a book by its public ID.
func (r *Repo) GetBookByPublicID(ctx context.Context, publicID string) (*ds.Book, error) {
	_, span := r.tracer.Start(ctx, "GetBookByPublicID")
	defer span.End()

	book := new(ds.Book)
	const query = `SELECT * FROM entities e JOIN books b USING (id) WHERE e.public_id = $1 AND e.deleted_at IS NULL`

	err := pgxscan.Get(ctx, r.getDB(ctx), book, query, publicID)
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
		SET public_id = $1, title = $2, visibility = $3, updated_at = NOW() 
		WHERE id = $4`

	const updateBookSQL = `
		UPDATE books 
		SET description = $1, author_name = $2 
		WHERE id = $3`

	err := r.exec(ctx, updateEntitySQL, b.PublicID, b.Title, b.Visibility, b.ID)
	if err != nil {
		return fmt.Errorf("update entity: %w", err)
	}

	err = r.exec(ctx, updateBookSQL, b.Description, b.AuthorName)
	if err != nil {
		return fmt.Errorf("update book: %w", err)
	}

	return nil
}

// FilterBooks ...
func (r *Repo) FilterBooks(ctx context.Context, f ds.BooksFilter) (books []ds.Book, count int, err error) {
	_, span := r.tracer.Start(ctx, "FilterBooks")
	defer span.End()

	count, err = r.filter("entities e").
		join("JOIN books b USING (id)").
		paginate(f.Page, f.PerPage).
		createdAt(f.CreatedAt).
		deletedAt(f.DeletedAt).
		deleted(f.Deleted).
		order(f.OrderBy, f.OrderDirection).
		where("status", f.Status).
		where("visibility", f.Visibility).
		withCount(f.WithCount).
		scan(ctx, &books)

	return
}
