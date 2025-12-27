package factory

import (
	"context"
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test/factory/random"
)

// NewChangeEmailRequest ...
func (f *Factory) NewChangeEmailRequest(overrideOpt ...ds.ChangeEmailRequest) (m *ds.ChangeEmailRequest) {
	m = &ds.ChangeEmailRequest{
		ID:        0,
		UserID:    0,
		NewEmail:  random.Email(),
		Token:     random.String(),
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}

	if len(overrideOpt) == 1 {
		merge(m, overrideOpt[0])
	}

	return
}

// CreateChangeEmailRequest ...
func (f *Factory) CreateChangeEmailRequest(t *testing.T, overrideOpt ...ds.ChangeEmailRequest) (
	m *ds.ChangeEmailRequest) {
	t.Helper()

	m = f.NewChangeEmailRequest(overrideOpt...)

	if m.UserID == 0 {
		m.UserID = f.CreateUser(t).ID
	}

	err := f.repo.CreateChangeEmailRequest(context.Background(), m)
	checkErr(t, err)

	return
}
