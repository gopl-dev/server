package ds

import "time"

// PasswordResetToken represents a record in the password_reset_tokens table.
// It stores a single-use token for a user to reset their password.
type PasswordResetToken struct {
	ID        int64     `json:"-"`
	UserID    int64     `json:"-"`
	Token     string    `json:"-"`
	ExpiresAt time.Time `json:"-"`
	CreatedAt time.Time `json:"-"`
}

// Invalid returns true if the password reset token has expired.
func (t *PasswordResetToken) Invalid() bool {
	return t.ExpiresAt.Before(time.Now())
}
