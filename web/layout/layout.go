package layout

import "github.com/a-h/templ"

type Data struct {
	Title string
	Head  templ.Component
	Body  templ.Component
	User  *User
}

type User struct {
	Username string
}
