package service

import (
	"context"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/oauth/provider"
)

var getOAuthUserAccount = z.Shape{
	"Provider":       provider.TypeInputRules,
	"ProviderUserID": z.String().Required(z.Message("provider_user_id is required")),
}

// GetOAuthUserAccount ...
func (s *Service) GetOAuthUserAccount(
	ctx context.Context, prov provider.Type, provUserID string) (m *ds.OAuthUserAccount, err error) {
	ctx, span := s.tracer.Start(ctx, "GetOAuthUserAccount")
	defer span.End()

	in := &GetOAuthUserAccountInput{
		Provider:       prov,
		ProviderUserID: provUserID,
	}
	err = Normalize(in)
	if err != nil {
		return
	}

	return s.db.GetOAuthUserAccount(ctx, in.Provider, in.ProviderUserID)
}

// GetOAuthUserAccountInput ...
type GetOAuthUserAccountInput struct {
	Provider       provider.Type
	ProviderUserID string
}

// Sanitize ...
func (in *GetOAuthUserAccountInput) Sanitize() {
	in.ProviderUserID = strings.TrimSpace(in.ProviderUserID)
}

// Validate ...
func (in *GetOAuthUserAccountInput) Validate() error {
	return validateInput(getOAuthUserAccount, in)
}
