package email

type ConfirmEmail struct {
	Username string
	Code     string
}

func (_ ConfirmEmail) Subject() string {
	return "Email confirmation"
}

func (_ ConfirmEmail) TemplateName() string {
	return "confirm_email"
}

func (c ConfirmEmail) Variables() map[string]any {
	return map[string]any{
		"username": c.Username,
		"code":     c.Code,
	}
}
