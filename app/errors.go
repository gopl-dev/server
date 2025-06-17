package app

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	CodeBadRequest    = http.StatusBadRequest
	CodeUnauthorized  = http.StatusUnauthorized
	CodeForbidden     = http.StatusForbidden
	CodeNotFound      = http.StatusNotFound
	CodeUnprocessable = http.StatusUnprocessableEntity
	CodeInternal      = http.StatusInternalServerError
)

type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string {
	return e.Message
}

func NewError(code int, message string, params ...interface{}) error {
	if len(params) > 0 {
		message = fmt.Sprintf(message, params...)
	}

	return Error{
		Code:    code,
		Message: message,
	}
}

func NewInputError() InputError {
	return InputError{}
}

type InputError map[string]string

func (i InputError) Add(k, v string) {
	i[k] = v
}

func (i InputError) AddIf(cond bool, k, v string) {
	if cond {
		i[k] = v
	}
}

func (i InputError) Has() bool {
	return len(i) > 0
}

func (i InputError) Error() string {
	s := strings.Builder{}
	for k, v := range i {
		s.WriteString(k + ": " + v + "\n")
	}
	return s.String()
}

func ErrUnprocessable(message string) error {
	return NewError(CodeUnprocessable, message)
}

func ErrNotFound(message string) error {
	return NewError(CodeNotFound, message)
}

func ErrInternal(message string) error {
	return NewError(CodeInternal, message)
}

func ErrBadRequest(message string) error {
	return NewError(CodeBadRequest, message)
}

func ErrUnauthorized() error {
	return NewError(CodeUnauthorized, "Unauthorized")
}

func ErrForbidden(message string) error {
	return NewError(CodeForbidden, message)
}
