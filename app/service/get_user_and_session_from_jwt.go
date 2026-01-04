package service

import (
	"context"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/session"
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
	user *ds.User, sess *ds.UserSession, err error) {
	ctx, span := s.tracer.Start(ctx, "GetUserAndSessionFromJWT")
	defer span.End()

	in := &GetUserAndSessionFromJWTInput{Token: token}
	err = Normalize(in)
	if err != nil {
		return
	}

	sessionID, userID, err := session.UnpackFromJWT(in.Token)
	if err != nil {
		return
	}

	sess, err = s.FindUserSessionByID(ctx, sessionID)
	if err != nil {
		return
	}

	if sess.UserID != userID {
		return nil, nil, ErrInvalidJWT
	}

	if sess.ExpiresAt.Before(time.Now()) {
		err = s.DeleteUserSession(ctx, sess.ID)
		if err != nil {
			return
		}

		err = ErrSessionExpired
		return
	}

	user, err = s.FindUserByID(ctx, sess.UserID)
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
