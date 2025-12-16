package email

import (
	"fmt"

	"github.com/gopl-dev/server/app"
)

// PasswordResetRequest contains the data needed to render the password reset email.
type PasswordResetRequest struct {
	Username string
	Token    string
}

// Subject returns the subject line for the password reset email.
func (p PasswordResetRequest) Subject() string {
	return "Password Reset Request"
}

// TemplateName returns the filename of the HTML template for the password reset email.
func (p PasswordResetRequest) TemplateName() string {
	return "password_reset"
}

// Variables returns the data to be used in the email template.
func (p PasswordResetRequest) Variables() map[string]any {
	return map[string]any{
		"username": p.Username,
		"link":     fmt.Sprintf("%s/password/reset?token=%s", app.Config().Server.Addr, p.Token),
	}
}
