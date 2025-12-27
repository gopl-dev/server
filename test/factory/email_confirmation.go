package factory

import (
	"context"
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test/factory/random"
)

// NewEmailConfirmation ...
func (f *Factory) NewEmailConfirmation(overrideOpt ...ds.EmailConfirmation) (m *ds.EmailConfirmation) {
	m = &ds.EmailConfirmation{
		ID:          ds.NilID,
		UserID:      ds.NilID,
		Code:        random.String(16), //nolint:mnd
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Hour),
		ConfirmedAt: nil,
	}

	if len(overrideOpt) == 1 {
		merge(m, overrideOpt[0])
	}

	return
}

// CreateEmailConfirmation ...
func (f *Factory) CreateEmailConfirmation(t *testing.T, overrideOpt ...ds.EmailConfirmation) (m *ds.EmailConfirmation) {
	t.Helper()

	m = f.NewEmailConfirmation(overrideOpt...)

	if m.UserID.IsNil() {
		m.UserID = f.CreateUser(t).ID
	}

	err := f.repo.CreateEmailConfirmation(context.Background(), m)
	checkErr(t, err)

	return
}
