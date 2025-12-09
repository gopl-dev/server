package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/email"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidEmailOrPassword is returned when a user attempts to log in with credentials
	// that do not match any record.
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")

	// ErrInvalidJWT is returned when an authentication token is malformed,
	// invalidly signed, or contains unexpected claims.
	ErrInvalidJWT = errors.New("invalid token")

	// ErrSessionExpired is returned when a JWT is validly signed but the associated
	// database session has expired based on its timestamp.
	ErrSessionExpired = errors.New("session expired")
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
)

const (
	ctxUserKey contextKey = "user"
)

// RegisterUserArgs defines the expected input parameters for the user registration process.
type RegisterUserArgs struct {
	Username string
	Email    string
	Password string
}

// RegisterUser handles the complete user registration process.
func (s *Service) RegisterUser(ctx context.Context, p RegisterUserArgs) (user *ds.User, err error) {
	err = app.Validate(ds.UserValidationRules, &p)
	if err != nil {
		return
	}

	existing, err := s.db.FindUserByEmail(ctx, p.Email)
	if err != nil {
		return
	}

	if existing != nil {
		err = app.InputError{"email": UserWithThisEmailAlreadyExists}

		return
	}

	existing, err = s.db.FindUserByUsername(ctx, p.Username)
	if err != nil {
		return
	}

	if existing != nil {
		err = app.InputError{"username": UsernameAlreadyTaken}

		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(p.Password), app.DefaultBCryptCost)
	if err != nil {
		return
	}

	user = &ds.User{
		Username:       p.Username,
		Email:          p.Email,
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

	err = email.Send(user.Email, email.ConfirmEmail{
		Username: user.Username,
		Email:    p.Email,
		Code:     emailConfirmCode,
	})

	return
}

// LoginUser authenticates a user using their email and password.
func (s *Service) LoginUser(ctx context.Context, email, password string) (user *ds.User, token string, err error) {
	user, err = s.db.FindUserByEmail(ctx, email)
	if err != nil {
		return
	}

	if user == nil {
		err = ErrInvalidEmailOrPassword

		return
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

// FindUserByID retrieves a user record from the database by their ID.
func (s *Service) FindUserByID(ctx context.Context, id int64) (user *ds.User, err error) {
	return s.db.FindUserByID(ctx, id)
}

// FindUserByEmail retrieves a user record from the database by their email address.
func (s *Service) FindUserByEmail(ctx context.Context, email string) (user *ds.User, err error) {
	return s.db.FindUserByEmail(ctx, email)
}

// CreateUser is a placeholder method for creating a user, currently non-functional
// or delegated/deprecated in favor of RegisterUser.
func (s *Service) CreateUser(_ *ds.User) (err error) {
	// err = database.ORM().Insert(u)
	return
}

// SetUserEmailConfirmed sets the email_confirmed flag to true for a user in the database.
func (s *Service) SetUserEmailConfirmed(ctx context.Context, userID int64) (err error) {
	return s.db.SetUserEmailConfirmed(ctx, userID)
}

// GetUserAndSessionFromJWT parses a JWT, validates it, checks the associated session's validity
// against the database, and retrieves the corresponding user record.
func (s *Service) GetUserAndSessionFromJWT(ctx context.Context, jwt string) (
	user *ds.User, session *ds.UserSession, err error) {
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

// UserToContext adds the given user object to the provided context.
func (s *Service) UserToContext(ctx context.Context, user *ds.User) context.Context {
	return context.WithValue(ctx, ctxUserKey, user)
}

// UserFromContext attempts to retrieve the authenticated user object from the context.
func (s *Service) UserFromContext(ctx context.Context) *ds.User {
	if v := ctx.Value(ctxUserKey); v != nil {
		if user, ok := v.(*ds.User); ok {
			return user
		}
	}

	return nil
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
