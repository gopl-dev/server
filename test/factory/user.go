package factory

import (
	"context"
	"time"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/test/factory/random"
	"golang.org/x/crypto/bcrypt"
)

// NewUser ...
func (f *Factory) NewUser(overrideOpt ...ds.User) (m *ds.User) {
	m = &ds.User{
		ID:             ds.NewID(),
		Username:       random.String(),
		Email:          random.Email(),
		EmailConfirmed: false,
		Password:       "",
		CreatedAt:      time.Now(),
		UpdatedAt:      nil,
		DeletedAt:      nil,
	}

	if len(overrideOpt) == 1 {
		merge(m, overrideOpt[0])
	}

	return
}

// CreateUser ...
func (f *Factory) CreateUser(overrideOpt ...ds.User) (m *ds.User, err error) {
	m = f.NewUser(overrideOpt...)

	password := m.Password
	if password == "" {
		password = random.String()
	}

	passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(password), app.DefaultBCryptCost)
	if err != nil {
		return
	}

	m.Password = string(passwordHashBytes)

	err = f.repo.CreateUser(context.Background(), m)
	return
}
