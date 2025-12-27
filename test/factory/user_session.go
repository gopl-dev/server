package factory

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gopl-dev/server/app/ds"
)

// NewUserSession ...
func (f *Factory) NewUserSession(overrideOpt ...ds.UserSession) (m *ds.UserSession) {
	m = &ds.UserSession{
		ID:        uuid.Nil,
		UserID:    0,
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
	if m.UserID == 0 {
		m.UserID = f.CreateUser(t).ID
	}

	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}

	err := f.repo.CreateUserSession(context.Background(), m)
	checkErr(t, err)

	return
}
