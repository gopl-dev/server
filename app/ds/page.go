package ds

import (
	"context"

	z "github.com/Oudwins/zog"
)

var pageCtxKey ctxKey = "page"

// Page defines the data structure for a page.
type Page struct {
	*Entity
}

// Data returns the editable fields of the Page as a key-value map.
func (p *Page) Data() map[string]any {
	return map[string]any{
		"public_id":   p.PublicID,
		"title":       p.Title,
		"description": p.Description,
	}
}

// CreateRules provides the validation map used when saving a new book.
func (p *Page) CreateRules() z.Shape {
	return z.Shape{
		"PublicID":    z.String().Trim().Required(),
		"Title":       z.String().Trim().Required(),
		"Description": z.String().Trim().Required(),
	}
}

// UpdateRules provides the validation map used when editing an existing book.
func (p *Page) UpdateRules() z.Shape {
	return p.CreateRules()
}

// ToContext adds the given book object to the provided context.
func (p *Page) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, pageCtxKey, p)
}

// PageFromContext attempts to retrieve page object from the context.
func PageFromContext(ctx context.Context) *Page {
	if v := ctx.Value(pageCtxKey); v != nil {
		if page, ok := v.(*Page); ok {
			return page
		}
	}

	return nil
}
