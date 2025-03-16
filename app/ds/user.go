package ds

import (
	"time"

	z "github.com/Oudwins/zog"
)

var UserValidationRules = z.Schema{
	"username": z.String().Min(2).Max(30).Required(z.Message("Username is required")),
	"email":    z.String().Email().Required(z.Message("Email is required")),
	"password": z.String().Required(z.Message("Password is required")),
}

type User struct {
	ID             int64      `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	EmailConfirmed bool       `json:"-"`
	Password       string     `json:"-"`
	CreatedAt      time.Time  `json:"-"`
	UpdatedAt      *time.Time `json:"-"`
	DeletedAt      *time.Time `json:"-"`
}
