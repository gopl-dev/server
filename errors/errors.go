package errors

import (
	"errors"
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

func New(code int, message string, params ...interface{}) error {
	if len(params) > 0 {
		message = fmt.Sprintf(message, params...)
	}

	return Error{
		Code:    code,
		Message: message,
	}
}

func NewInput() Input {
	return Input{}
}

type Input map[string]string

func (i Input) Add(k, v string) {
	i[k] = v
}

func (i Input) AddIf(cond bool, k, v string) {
	if cond {
		i[k] = v
	}
}

func (i Input) Has() bool {
	return len(i) > 0
}

func (i Input) Error() string {
	s := strings.Builder{}
	for k, v := range i {
		s.WriteString(k + ": " + v + "\n")
	}
	return s.String()
}

func Std(message string) error {
	return errors.New(message)
}

func Unprocessable(message string) error {
	return New(CodeUnprocessable, message)
}

func NotFound(message string) error {
	return New(CodeNotFound, message)
}

func Internal(message string) error {
	return New(CodeInternal, message)
}

func BadRequest(message string) error {
	return New(CodeBadRequest, message)
}

func Unauthorized(message string) error {
	return New(CodeUnauthorized, message)
}

func Forbidden(message string) error {
	return New(CodeForbidden, message)
}
