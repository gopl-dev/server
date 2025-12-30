package service

import (
	"context"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

var getUserAndSessionFromJWTInputRules = z.Shape{
	"Token": z.String().Required(z.Message("Token is required")),
}

var (
	// ErrInvalidJWT is returned when an authentication token is malformed,
	// invalidly signed, or contains unexpected claims.
	ErrInvalidJWT = app.ErrForbidden("invalid token")

	// ErrSessionExpired is returned when a JWT is validly signed but the associated
	// database session has expired based on its timestamp.
	ErrSessionExpired = app.ErrForbidden("session expired")
)

// GetUserAndSessionFromJWT parses a JWT, validates it, checks the associated session's validity
// against the database, and retrieves the corresponding user record.
func (s *Service) GetUserAndSessionFromJWT(ctx context.Context, token string) (
	user *ds.User, session *ds.UserSession, err error) {
	ctx, span := s.tracer.Start(ctx, "GetUserAndSessionFromJWT")
	defer span.End()

	in := &GetUserAndSessionFromJWTInput{Token: token}
	err = Normalize(in)
	if err != nil {
		return
	}

	sessionID, userID, err := app.UnpackSessionFromJWT(in.Token)
	if err != nil {
		return
	}

	session, err = s.FindUserSessionByID(ctx, sessionID)
	if err != nil {
		return
	}

	if session.UserID != userID {
		return nil, nil, ErrInvalidJWT
	}

	if session.ExpiresAt.Before(time.Now()) {
		err = s.DeleteUserSession(ctx, session.ID)
		if err != nil {
			return
		}

		err = ErrSessionExpired
		return
	}

	user, err = s.FindUserByID(ctx, session.UserID)
	return
}

// GetUserAndSessionFromJWTInput ...
type GetUserAndSessionFromJWTInput struct {
	Token string
}

// Sanitize ...
func (in *GetUserAndSessionFromJWTInput) Sanitize() {
	in.Token = strings.TrimSpace(in.Token)
}

// Validate ...
func (in *GetUserAndSessionFromJWTInput) Validate() error {
	return validateInput(getUserAndSessionFromJWTInputRules, in)
}
