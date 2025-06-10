package layout

import "github.com/a-h/templ"

type Data struct {
	Title           string
	MetaAuthor      string
	MetaDescription string
	MetaKeywords    string
	Body            templ.Component
	User            *User
}

type User struct {
	Username string
}
