package factory

import (
	"context"
	"testing"
	"time"

	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/oauth/provider"
	"github.com/gopl-dev/server/test/factory/random"
)

// NewOAuthUserAccount ...
func (f *Factory) NewOAuthUserAccount(overrideOpt ...ds.OAuthUserAccount) (m *ds.OAuthUserAccount) {
	m = &ds.OAuthUserAccount{
		ID:             ds.NilID,
		UserID:         ds.NilID,
		Provider:       random.Element(provider.Types),
		ProviderUserID: random.String(),
		CreatedAt:      time.Now(),
	}

	if len(overrideOpt) == 1 {
		merge(m, overrideOpt[0])
	}

	return
}

// CreateOAuthUserAccount ...
func (f *Factory) CreateOAuthUserAccount(t *testing.T, overrideOpt ...ds.OAuthUserAccount) (
	m *ds.OAuthUserAccount) {
	t.Helper()

	m = f.NewOAuthUserAccount(overrideOpt...)

	if m.UserID.IsNil() {
		m.UserID = f.CreateUser(t).ID
	}

	err := f.repo.CreateOAuthUserAccount(context.Background(), m)
	checkErr(t, err)

	return
}
