package email

import (
	"path"

	"github.com/gopl-dev/server/app"
)

// BookApproved represents the email payload sent when a book
// has been approved and published.
type BookApproved struct {
	BookName string
	Username string
	PublicID string
}

// Subject returns the email subject for a book approval notification.
func (BookApproved) Subject() string {
	return "Your book is online!"
}

// TemplateName returns the name of the email template used for this message.
func (BookApproved) TemplateName() string {
	return "book_approved"
}

// Variables returns the template variables used to render the email body.
func (c BookApproved) Variables() map[string]any {
	return map[string]any{
		"username":      c.Username,
		"book_name":     c.BookName,
		"view_book_url": path.Join(app.Config().Server.Addr, "/books/"+c.PublicID+"/"),
	}
}
