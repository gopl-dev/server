package service

import (
	"context"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/markbates/goth"
)

var authenticateOAuthUser = z.Shape{
	"Email":    emailInputRules,
	"Provider": z.String().Required(z.Message("provider is required")),
	"UserID":   z.String().Required(z.Message("user_id is required")),
}

// AuthenticateOAuthUser ...
func (s *Service) AuthenticateOAuthUser(ctx context.Context, authAcc goth.User) (token string, err error) {
	ctx, span := s.tracer.Start(ctx, "AuthenticateOAuthUser")
	defer span.End()

	in := &AuthenticateOAuthUserInput{&authAcc}
	user, err := s.ResolveUserFromOAuthAccount(ctx, *in.User)
	if err != nil {
		return
	}

	return s.newSignedSessionToken(ctx, user.ID)
}

// AuthenticateOAuthUserInput ...
type AuthenticateOAuthUserInput struct {
	*goth.User
}

// Sanitize ...
func (in *AuthenticateOAuthUserInput) Sanitize() {
	in.UserID = strings.TrimSpace(in.UserID)
	in.Provider = strings.TrimSpace(in.Provider)
	in.Email = strings.TrimSpace(in.Email)
}

// Validate ...
func (in *AuthenticateOAuthUserInput) Validate() error {
	return validateInput(authenticateOAuthUser, in)
}
