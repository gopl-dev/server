package service

import (
	"context"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/repo"
	"github.com/gopl-dev/server/email"
	"golang.org/x/crypto/bcrypt"
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
		err = app.InputError{"email": "User with this email already exists"}
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

func CreateUser(u *ds.User) (err error) {
	//err = database.ORM().Insert(u)
	return
}

func UpdateUser(u *ds.User) (err error) {
	//err = database.ORM().Update(u)
	return
}
