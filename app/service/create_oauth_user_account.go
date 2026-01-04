package service

import (
	"context"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/oauth/provider"
)

var createOAuthUserAccountInputRules = z.Shape{
	"UserID":         ds.IDInputRules,
	"Provider":       provider.TypeInputRules,
	"ProviderUserID": z.String().Required(z.Message("provider_user_id is required")),
}

// CreateOAuthUserAccount creates a new user session object.
func (s *Service) CreateOAuthUserAccount(ctx context.Context, m *ds.OAuthUserAccount) (err error) {
	ctx, span := s.tracer.Start(ctx, "CreateOAuthUserAccount")
	defer span.End()

	in := &CreateOAuthUserAccountInput{m}
	err = Normalize(in)
	if err != nil {
		return
	}

	return s.db.CreateOAuthUserAccount(ctx, in.OAuthUserAccount)
}

// CreateOAuthUserAccountInput ...
type CreateOAuthUserAccountInput struct {
	*ds.OAuthUserAccount
}

// Sanitize ...
func (in *CreateOAuthUserAccountInput) Sanitize() {
	in.ProviderUserID = strings.TrimSpace(in.ProviderUserID)
}

// Validate ...
func (in *CreateOAuthUserAccountInput) Validate() error {
	return validateInput(createOAuthUserAccountInputRules, in)
}
