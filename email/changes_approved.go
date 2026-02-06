package email

import (
	"github.com/gopl-dev/server/app"
)

// ChangesApproved represents the email payload sent when an entity changes
// has been approved and applied.
type ChangesApproved struct {
	Username    string
	EntityTitle string
	ViewURL     string
}

// Subject returns the email subject for a book approval notification.
func (ChangesApproved) Subject() string {
	return "Your changes have been approved!"
}

// TemplateName returns the name of the email template used for this message.
func (ChangesApproved) TemplateName() string {
	return "changes_approved"
}

// Variables returns the template variables used to render the email body.
func (c ChangesApproved) Variables() map[string]any {
	return map[string]any{
		"username":     c.Username,
		"entity_title": c.EntityTitle,
		"view_url":     app.ServerURL(c.ViewURL),
	}
}
