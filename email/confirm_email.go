package email

import (
	"path"

	"github.com/gopl-dev/server/app"
)

type ConfirmEmail struct {
	Username   string
	Email      string
	Code       string
	ConfirmUrl string
}

func (ConfirmEmail) Subject() string {
	return "Email confirmation"
}

func (ConfirmEmail) TemplateName() string {
	return "confirm_email"
}

func (c ConfirmEmail) Variables() map[string]any {
	return map[string]any{
		"username":    c.Username,
		"email":       c.Email,
		"code":        c.Code,
		"confirm_url": path.Join(app.Config().Server.Addr, "/users/confirm-email/"),
	}
}
