package request

import (
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gopl-dev/server/app/ds"
	"github.com/gopl-dev/server/app/service"
)

type RegisterUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *RegisterUser) Sanitize() {
	r.Username = strings.TrimSpace(r.Username)
	r.Email = strings.TrimSpace(r.Email)
	r.Password = strings.TrimSpace(r.Password)
}

func (r *RegisterUser) ValidationSchema() z.Shape {
	return ds.UserValidationRules
}

func (r RegisterUser) ToParams() service.RegisterUserArgs {
	return service.RegisterUserArgs{
		Username: r.Username,
		Email:    r.Email,
		Password: r.Password,
	}
}

type ConfirmEmail struct {
	Code string `json:"code"`
}

func (r *ConfirmEmail) Sanitize() {
	r.Code = strings.TrimSpace(r.Code)
}
