package layout

import (
	"github.com/a-h/templ"
	"github.com/gopl-dev/server/frontend"
)

type Data struct {
	Title           string
	MetaAuthor      string
	MetaDescription string
	MetaKeywords    string
	Body            templ.Component
	User            *frontend.User
}
