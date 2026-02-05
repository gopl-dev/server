package email

// BookRejected represents the email payload used when a submitted book
// is rejected by moderation.
type BookRejected struct {
	Note     string
	BookName string
	Username string
}

// Subject returns the email subject for a book rejection notification.
func (BookRejected) Subject() string {
	return "Your book wasn’t approved"
}

// TemplateName returns the name of the email template used for this message.
func (BookRejected) TemplateName() string {
	return "book_rejected"
}

// TODO: At the end, add something like:
//  "If you think your book wasn't approved by error, please reach out to us: <contact options coming later>"

// Variables returns the template variables used to render the email body.
func (c BookRejected) Variables() map[string]any {
	return map[string]any{
		"username":  c.Username,
		"book_name": c.BookName,
		"note":      c.Note,
	}
}
