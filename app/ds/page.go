package ds

import (
	"context"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds/prop"
)

var pageCtxKey ctxKey = "page"

// Page defines the data structure for a page.
type Page struct {
	*Entity

	ContentRaw string `json:"-"`
	Content    string `json:"content"`
}

// Data returns the editable fields of the Page as a key-value map.
func (p *Page) Data() map[string]any {
	return p.WithEntityData(map[string]any{
		"content": p.ContentRaw,
	})
}

// PropertyType returns the property type for a given key.
func (p *Page) PropertyType(key string) prop.Type {
	// should rewrite switch statement to if statement (gocritic)
	// switch key {
	// case "content":
	// 		return prop.Markdown
	// }
	if key == "content" {
		return prop.Markdown
	}

	return p.Entity.PropertyType(key)
}

// CreateRules provides the validation map used when saving a new book.
func (p *Page) CreateRules() z.Shape {
	return z.Shape{
		"PublicID": z.String().Trim().Required(),
		"Title":    z.String().Trim().Required(),
		"Content":  z.String().Required(),
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
