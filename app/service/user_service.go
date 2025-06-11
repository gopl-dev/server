package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/email"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrInvalidJWT             = errors.New("invalid token")
	ErrSessionExpired         = errors.New("session expired")
)

var (
	jwtSessionParam = "session"
	jwtUserParam    = "user"
)

const (
	ctxUserKey contextKey = "user"
)

type RegisterUserArgs struct {
	Username string
	Email    string
	Password string
}

func RegisterUser(ctx context.Context, p RegisterUserArgs) (user *ds.User, err error) {
	err = app.Validate(ds.UserValidationRules, &p)
	if err != nil {
		return
	}

	existing, err := repo.FindUserByEmail(ctx, p.Email)
	if err != nil {
		return
	}
	if existing != nil {
		err = app.InputError{"email": "User with this email already exists."}
		return
	}
	existing, err = repo.FindUserByUsername(ctx, p.Username)
	if err != nil {
		return
	}
	if existing != nil {
		err = app.InputError{"username": "Username already taken"}
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(p.Password), 11)
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

	err = repo.CreateUser(ctx, user)
	if err != nil {
		return
	}

	emailConfirmCode, err := CreateEmailConfirmation(ctx, user.ID)
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

func LoginUser(ctx context.Context, email, password string) (user *ds.User, token string, err error) {
	user, err = repo.FindUserByEmail(ctx, email)
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

	sess, err := CreateUserSession(ctx, user.ID)
	if err != nil {
		return
	}

	token, err = newSignedSessionJWT(sess.ID.String(), user.ID)
	if err != nil {
		return
	}

	return user, token, nil
}

func FindUserByID(ctx context.Context, id int64) (user *ds.User, err error) {
	return repo.FindUserByID(ctx, id)
}

func CreateUser(u *ds.User) (err error) {
	//err = database.ORM().Insert(u)
	return
}

// SetUserEmailConfirmed sets the email_confirmed flag for a user.
func SetUserEmailConfirmed(ctx context.Context, userID int64) (err error) {
	return repo.SetUserEmailConfirmed(ctx, userID)
}

func GetUserAndSessionFromJWT(ctx context.Context, jwt string) (user *ds.User, session *ds.UserSession, err error) {
	sessionID, userID, err := unpackSessionJWT(jwt)
	if err != nil {
		return
	}

	session, err = FindUserSessionByID(ctx, sessionID)
	if err != nil || session == nil {
		return
	}

	if session.UserID != userID {
		err = ErrInvalidJWT
		return
	}

	if session.ExpiresAt.Before(time.Now()) {
		err = DeleteUserSession(ctx, session.ID)
		if err != nil {
			return
		}
		err = ErrSessionExpired
		return
	}

	user, err = FindUserByID(ctx, session.UserID)
	if err != nil {
		return
	}

	return
}

func newSignedSessionJWT(sessionID string, userID int64) (token string, err error) {
	jt := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.MapClaims{
			jwtSessionParam: sessionID,
			jwtUserParam:    userID,
		})

	return jt.SignedString([]byte(app.Config().Session.Key))
}

func unpackSessionJWT(jt string) (sessionID string, userID int64, err error) {
	token, err := jwt.Parse(jt, func(token *jwt.Token) (any, error) {
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

func UserToContext(ctx context.Context, user *ds.User) context.Context {
	return context.WithValue(ctx, ctxUserKey, user)
}

func UserFromContext(ctx context.Context) *ds.User {
	if v := ctx.Value(ctxUserKey); v != nil {
		return v.(*ds.User)
	}

	return nil
}
