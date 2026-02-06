package email

import (
	"github.com/gopl-dev/server/app"
)

// ChangesRejected represents the email payload sent when an entity changes
// has been rejected.
type ChangesRejected struct {
	Username    string
	EntityTitle string
	Note        string
	ViewURL     string
}

// Subject returns the email subject for a book approval notification.
func (ChangesRejected) Subject() string {
	return "Your changes were not approved"
}

// TemplateName returns the name of the email template used for this message.
func (ChangesRejected) TemplateName() string {
	return "changes_rejected"
}

// Variables returns the template variables used to render the email body.
func (c ChangesRejected) Variables() map[string]any {
	return map[string]any{
		"username":     c.Username,
		"entity_title": c.EntityTitle,
		"note":         c.Note,
		"view_url":     app.ServerURL(c.ViewURL),
	}
}
