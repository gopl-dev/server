package factory

import (
	"context"
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
)

// NewUserSession ...
func (f *Factory) NewUserSession(overrideOpt ...ds.UserSession) (m *ds.UserSession) {
	m = &ds.UserSession{
		ID:        ds.NilID,
		UserID:    ds.NilID,
		CreatedAt: time.Now(),
		UpdatedAt: nil,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	if len(overrideOpt) == 1 {
		merge(m, overrideOpt[0])
	}

	return
}

// CreateUserSession ...
func (f *Factory) CreateUserSession(t *testing.T, overrideOpt ...ds.UserSession) (m *ds.UserSession) {
	t.Helper()

	m = f.NewUserSession(overrideOpt...)
	if m.UserID.IsNil() {
		m.UserID = f.CreateUser(t).ID
	}

	err := f.repo.CreateUserSession(context.Background(), m)
	checkErr(t, err)

	return
}
