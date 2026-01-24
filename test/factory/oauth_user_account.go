package factory

import (
	"context"
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
func (f *Factory) CreateOAuthUserAccount(overrideOpt ...ds.OAuthUserAccount) (
	m *ds.OAuthUserAccount, err error) {
	m = f.NewOAuthUserAccount(overrideOpt...)

	if m.UserID.IsNil() {
		u, err := f.CreateUser()
		if err != nil {
			return nil, err
		}

		m.UserID = u.ID
	}

	err = f.repo.CreateOAuthUserAccount(context.Background(), m)
	return
}
