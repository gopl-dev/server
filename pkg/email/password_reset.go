package email

import (
	"fmt"

	"github.com/gopl-dev/server/app"
)

// PasswordResetRequest ...
type PasswordResetRequest struct {
	Username string
	Token    string
}

// Subject ...
func (p PasswordResetRequest) Subject() string {
	return "Password Reset Request"
}

// TemplateName ...
func (p PasswordResetRequest) TemplateName() string {
	return "password_reset"
}

// Variables ...
func (p PasswordResetRequest) Variables() map[string]any {
	return map[string]any{
		"Username": p.Username,
		"Link":     fmt.Sprintf("%sreset-password/%s/", app.Config().Server.Addr, p.Token),

		// Token var is not used in email template, it is here to make testing easier.
		"token": p.Token,
	}
}
