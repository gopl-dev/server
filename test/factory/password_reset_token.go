package factory

import (
	"context"
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test/factory/random"
)

// NewPasswordResetToken ...
func (f *Factory) NewPasswordResetToken(overrideOpt ...ds.PasswordResetToken) (m *ds.PasswordResetToken) {
	m = &ds.PasswordResetToken{
		ID:        0,
		UserID:    0,
		Token:     random.String(),
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	if len(overrideOpt) == 1 {
		merge(m, overrideOpt[0])
	}

	return
}

// CreatePasswordResetToken ...
func (f *Factory) CreatePasswordResetToken(t *testing.T, overrideOpt ...ds.PasswordResetToken) (
	m *ds.PasswordResetToken) {
	t.Helper()

	m = f.NewPasswordResetToken(overrideOpt...)

	if m.UserID == 0 {
		m.UserID = f.CreateUser(t).ID
	}

	err := f.repo.CreatePasswordResetToken(context.Background(), m)
	checkErr(t, err)

	return
}
