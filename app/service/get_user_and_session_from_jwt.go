package service

import (
	"context"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

var getUserAndSessionFromJWTInputRules = z.Shape{
	"token": z.String().Required(z.Message("Token is required")),
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

	err = ValidateGetUserAndSessionFromJWTInput(token)
	if err != nil {
		return
	}

	sessionID, userID, err := app.UnpackSessionJWT(token)
	if err != nil {
		return
	}

	session, err = s.FindUserSessionByID(ctx, sessionID)
	if err != nil || session == nil {
		return
	}

	if session.UserID != userID {
		err = ErrInvalidJWT
		return
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

// ValidateGetUserAndSessionFromJWTInput ...
func ValidateGetUserAndSessionFromJWTInput(token string) (err error) {
	in := &GetUserAndSessionFromJWTInput{
		Token: token,
	}

	return validateInput(getUserAndSessionFromJWTInputRules, in)
}

// GetUserAndSessionFromJWTInput ...
type GetUserAndSessionFromJWTInput struct {
	Token string
}
