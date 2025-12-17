package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/pkg/email"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidEmailOrPassword is returned when a user attempts to log in with credentials
	// that do not match any record.
	ErrInvalidEmailOrPassword = app.ErrUnprocessable("invalid email or password")

	// ErrInvalidPassword is returned when a user tries to change their password
	// but provides an incorrect old password.
	ErrInvalidPassword = app.ErrUnprocessable("invalid password")

	// ErrInvalidJWT is returned when an authentication token is malformed,
	// invalidly signed, or contains unexpected claims.
	ErrInvalidJWT = app.ErrForbidden("invalid token")

	// ErrSessionExpired is returned when a JWT is validly signed but the associated
	// database session has expired based on its timestamp.
	ErrSessionExpired = app.ErrForbidden("session expired")

	// ErrTokenExpired ...
	ErrTokenExpired = app.ErrUnprocessable("token expired")

	// ErrInvalidPasswordResetToken ...
	ErrInvalidPasswordResetToken = app.ErrUnprocessable("password reset request is either expired or invalid")
)

var (
	jwtSessionParam = "session"
	jwtUserParam    = "user"
)

const (
	// UserWithThisEmailAlreadyExists is the specific error message for email validation failure during registration.
	UserWithThisEmailAlreadyExists = "User with this email already exists."

	// UsernameAlreadyTaken is the specific error message for username validation failure during registration.
	UsernameAlreadyTaken = "Username already taken"

	passwordResetTokenLength = 32
)

// RegisterUserArgs defines the expected input parameters for the user registration process.
type RegisterUserArgs struct {
	Username string
	Email    string
	Password string
}

// RegisterUser handles the complete user registration process.
func (s *Service) RegisterUser(ctx context.Context, args RegisterUserArgs) (user *ds.User, err error) {
	ctx, span := s.tracer.Start(ctx, "RegisterUser")
	defer span.End()

	err = app.Validate(ds.UserValidationRules, &args)
	if err != nil {
		return
	}

	_, err = s.db.FindUserByEmail(ctx, args.Email)
	if err == nil {
		return nil, app.InputError{"email": UserWithThisEmailAlreadyExists}
	}
	if !errors.Is(err, repo.ErrUserNotFound) {
		return nil, err
	}

	_, err = s.db.FindUserByUsername(ctx, args.Username)
	if err == nil {
		return nil, app.InputError{"username": UsernameAlreadyTaken}
	}
	if !errors.Is(err, repo.ErrUserNotFound) {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(args.Password), app.DefaultBCryptCost)
	if err != nil {
		return
	}

	user = &ds.User{
		Username:       args.Username,
		Email:          args.Email,
		Password:       string(passwordHash),
		EmailConfirmed: false,
		CreatedAt:      time.Now(),
	}

	err = s.db.CreateUser(ctx, user)
	if err != nil {
		return
	}

	emailConfirmCode, err := s.CreateEmailConfirmation(ctx, user.ID)
	if err != nil {
		return
	}

	// todo send email async
	err = email.Send(user.Email, email.ConfirmEmail{
		Username: user.Username,
		Email:    args.Email,
		Code:     emailConfirmCode,
	})
	if err != nil {
		return
	}

	err = s.LogUserRegistered(ctx, user.ID)
	return
}

// LoginUser authenticates a user using their email and password.
func (s *Service) LoginUser(ctx context.Context, email, password string) (user *ds.User, token string, err error) {
	ctx, span := s.tracer.Start(ctx, "loginUser")
	defer span.End()

	user, err = s.db.FindUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repo.ErrUserNotFound) {
			return nil, "", ErrInvalidEmailOrPassword
		}
		return nil, "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		err = ErrInvalidEmailOrPassword
		return
	}

	sess, err := s.CreateUserSession(ctx, user.ID)
	if err != nil {
		return
	}

	token, err = newSignedSessionJWT(sess.ID.String(), user.ID)
	if err != nil {
		return
	}

	return user, token, nil
}

// ChangeUserPasswordArgs defines the input for changing a user's password.
type ChangeUserPasswordArgs struct {
	UserID      int64
	OldPassword string
	NewPassword string
}

// ChangeUserPassword handles the logic for an authenticated user to change their own password.
func (s *Service) ChangeUserPassword(ctx context.Context, args ChangeUserPasswordArgs) (err error) {
	ctx, span := s.tracer.Start(ctx, "ChangeUserPassword")
	defer span.End()

	err = app.Validate(ds.ChangePasswordValidationRules, &args)
	if err != nil {
		return
	}

	// The user is usually already resolved at this point,
	// but we do not know where this method will be used,
	// so we select the user again to keep things simple
	//
	// (TODO?) When/if performance becomes an issue, this could be split into two methods:
	// 1. ChangePasswordByUser(ctx, *ds.User?) - explicitly uses the user from context or argument.
	// 2. ChangePasswordByUserID(int64) - selects the user from the repo (current behavior).
	user, err := s.FindUserByID(ctx, args.UserID)
	if user == nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(args.OldPassword))
	if err != nil {
		return app.InputError{"old_password": ErrInvalidPassword.Error()}
	}

	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte(args.NewPassword), app.DefaultBCryptCost)
	if err != nil {
		return err
	}

	return s.db.UpdateUserPassword(ctx, user.ID, string(newPasswordHash))
}

// FindUserByID retrieves a user record from the database by their ID.
func (s *Service) FindUserByID(ctx context.Context, id int64) (user *ds.User, err error) {
	ctx, span := s.tracer.Start(ctx, "FindUserByID")
	defer span.End()

	return s.db.FindUserByID(ctx, id)
}

// FindUserByEmail retrieves a user record from the database by their email address.
func (s *Service) FindUserByEmail(ctx context.Context, email string) (user *ds.User, err error) {
	ctx, span := s.tracer.Start(ctx, "FindUserByEmail")
	defer span.End()

	return s.db.FindUserByEmail(ctx, email)
}

// SetUserEmailConfirmed sets the email_confirmed flag to true for a user in the database.
func (s *Service) SetUserEmailConfirmed(ctx context.Context, userID int64) (err error) {
	ctx, span := s.tracer.Start(ctx, "SetUserEmailConfirmed")
	defer span.End()

	return s.db.SetUserEmailConfirmed(ctx, userID)
}

// GetUserAndSessionFromJWT parses a JWT, validates it, checks the associated session's validity
// against the database, and retrieves the corresponding user record.
func (s *Service) GetUserAndSessionFromJWT(ctx context.Context, jwt string) (
	user *ds.User, session *ds.UserSession, err error) {
	ctx, span := s.tracer.Start(ctx, "GetUserAndSessionFromJWT")
	defer span.End()

	sessionID, userID, err := unpackSessionJWT(jwt)
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
	if err != nil {
		return
	}

	return
}

// FindPasswordResetToken ...
func (s *Service) FindPasswordResetToken(ctx context.Context, token string) (t *ds.PasswordResetToken, err error) {
	ctx, span := s.tracer.Start(ctx, "FindPasswordResetToken")
	defer span.End()

	t, err = s.db.FindPasswordResetToken(ctx, token)
	if errors.Is(err, repo.ErrPasswordResetTokenNotFound) {
		return nil, ErrInvalidPasswordResetToken
	}
	if err != nil {
		return nil, err
	}

	if t.Invalid() {
		return nil, ErrInvalidPasswordResetToken
	}

	return t, nil
}

// PasswordResetRequest handles the logic for initiating a password reset.
// It finds the user by email, generates a unique token, and sends it to the user's email.
func (s *Service) PasswordResetRequest(ctx context.Context, emailAddr string) error {
	ctx, span := s.tracer.Start(ctx, "PasswordResetRequest")
	defer span.End()

	user, err := s.db.FindUserByEmail(ctx, emailAddr)
	if err != nil {
		// If the user is not found, we don't return an error to prevent email enumeration attacks.
		if errors.Is(err, repo.ErrUserNotFound) {
			return nil
		}

		return err
	}

	resetToken, err := app.Token(passwordResetTokenLength)
	if err != nil {
		return err
	}

	token := &ds.PasswordResetToken{
		UserID:    user.ID,
		Token:     resetToken,
		ExpiresAt: time.Now().Add(time.Hour * 1),
		CreatedAt: time.Now(),
	}

	err = s.db.CreatePasswordResetToken(ctx, token)
	if err != nil {
		return err
	}

	// TODO: Send email asynchronously
	return email.Send(user.Email, email.PasswordResetRequest{
		Username: user.Username,
		Token:    resetToken,
	})
}

// ResetPassword handles the logic for resetting a user's password using a token.
// It validates the token, checks for expiration, updates the user's password, and deletes the token.
func (s *Service) ResetPassword(ctx context.Context, token, newPassword string) error {
	ctx, span := s.tracer.Start(ctx, "ResetPassword")
	defer span.End()

	prt, err := s.db.FindPasswordResetToken(ctx, token)
	if errors.Is(err, repo.ErrPasswordResetTokenNotFound) {
		return ErrInvalidPasswordResetToken
	}
	if err != nil {
		return err
	}
	if prt.Invalid() {
		return ErrInvalidPasswordResetToken
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), app.DefaultBCryptCost)
	if err != nil {
		return err
	}

	err = s.db.UpdateUserPassword(ctx, prt.UserID, string(passwordHash))
	if err != nil {
		return err
	}

	return s.db.DeletePasswordResetToken(ctx, prt.ID)
}

// newSignedSessionJWT creates a new, signed JWT token containing the session ID and user ID claims.
// The token is signed using the secret key from the application configuration.
func newSignedSessionJWT(sessionID string, userID int64) (token string, err error) {
	jt := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			jwtSessionParam: sessionID,
			jwtUserParam:    userID,
		})

	return jt.SignedString([]byte(app.Config().Session.Key))
}

// unpackSessionJWT validates and parses a signed JWT string.
// It extracts the session ID (string) and user ID (int64) from the claims.
func unpackSessionJWT(jt string) (sessionID string, userID int64, err error) {
	token, err := jwt.Parse(jt, func(_ *jwt.Token) (any, error) {
		return []byte(app.Config().Session.Key), nil
	})
	if err != nil {
		return
	}

	if !token.Valid {
		err = ErrInvalidJWT
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = ErrInvalidJWT
		return
	}

	sessionID, ok = claims[jwtSessionParam].(string)
	if !ok {
		err = ErrInvalidJWT
		return
	}

	userIDFloat, ok := claims[jwtUserParam].(float64)
	if !ok {
		err = ErrInvalidJWT
		return
	}

	userID = int64(userIDFloat)

	return
}
