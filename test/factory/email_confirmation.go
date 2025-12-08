package factory

import (
	"context"
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test"
	"github.com/gopl-dev/server/test/factory/random"
)

func (f *Factory) NewEmailConfirmation(overrideOpt ...ds.EmailConfirmation) (m *ds.EmailConfirmation) {
	m = &ds.EmailConfirmation{
		ID:          0,
		UserID:      0,
		Code:        random.String(16),
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(time.Hour),
		ConfirmedAt: nil,
	}

	if len(overrideOpt) == 1 {
		merge(m, overrideOpt[0])
	}

	return
}

func (f *Factory) CreateEmailConfirmation(t *testing.T, overrideOpt ...ds.EmailConfirmation) (m *ds.EmailConfirmation) {
	m = f.NewEmailConfirmation(overrideOpt...)

	if m.UserID == 0 {
		m.UserID = f.CreateUser(t).ID
	}

	err := f.repo.CreateEmailConfirmation(context.Background(), m)
	test.CheckErr(t, err)
	return
}
