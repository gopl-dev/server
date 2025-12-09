// Package layout defines base data structures and HTML code to render the layouts.
package layout

import (
	"github.com/a-h/templ"
	"github.com/gopl-dev/server/frontend"
)

// Data represents the data model passed to the base layout template.
type Data struct {
	Title           string
	MetaAuthor      string
	MetaDescription string
	MetaKeywords    string
	Body            templ.Component
	User            *frontend.User
}
