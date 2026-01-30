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
		JOIN books b USING (id)
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
	const query = `SELECT * FROM entities e JOIN books b USING (id) WHERE e.public_id = $1 AND e.type = $2 AND e.deleted_at IS NULL LIMIT 1`

	err := pgxscan.Get(ctx, r.getDB(ctx), book, query, publicID, ds.EntityTypeBook)
	if noRows(err) {
		return nil, ErrBookNotFound
	}

	return book, err
}

// UpdateBook updates both entity and book tables.
func (r *Repo) UpdateBook(ctx context.Context, b *ds.Book) error {
	_, span := r.tracer.Start(ctx, "UpdateBook")
	defer span.End()

	err := r.update(ctx, b.ID, "books", data{
		"cover_file_id": b.CoverFileID,
		"author_name":   b.AuthorName,
		"author_link":   b.AuthorLink,
		"homepage":      b.Homepage,
		"release_date":  b.ReleaseDate,
	})
	if err != nil {
		return fmt.Errorf("update book: %w", err)
	}

	return nil
}

// FilterBooks ...
func (r *Repo) FilterBooks(ctx context.Context, f ds.BooksFilter) (books []ds.Book, count int, err error) {
	_, span := r.tracer.Start(ctx, "FilterBooks")
	defer span.End()

	count, err = r.filter("entities e", "e").
		columns(`
		  e.id            AS id,
		  e.public_id,
		  e.owner_id,
		  e.title,
		  e.description,
		  e.visibility,
		  e.status,
		  e.created_at,
		  e.updated_at,
		  e.deleted_at,
		
		  b.cover_file_id,
		  b.author_name,
		  b.author_link,
		  b.homepage,
		  b.release_date,
		  u.username AS "owner"`).
		join("LEFT JOIN books b USING (id)").
		join("LEFT JOIN users u ON e.owner_id = u.id").
		paginate(f.Page, f.PerPage).
		createdAt(f.CreatedAt).
		deletedAt(f.DeletedAt).
		deleted(f.Deleted).
		order(f.OrderBy, f.OrderDirection).
		apply(
			whereIn("e.status", f.Status),
			whereIn("e.visibility", f.Visibility),
		).
		withCount(f.WithCount).
		scan(ctx, &books)

	return
}
