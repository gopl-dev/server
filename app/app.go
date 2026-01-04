// Package app ...
package app

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	z "github.com/Oudwins/zog"
	"github.com/gosimple/slug"
)

const (
	// DevEnv is used for the local development environment.
	DevEnv = "dev"

	// TestEnv is used for running automated tests and quality assurance (QA) checks.
	TestEnv = "test"

	// StagingEnv is a production-like environment used for final testing before a full public release.
	StagingEnv = "staging"

	// ReleaseEnv refers to the final production environment serving live users.
	ReleaseEnv = "release"
)

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")

	jwtSessionParam = "session"
	jwtUserParam    = "user"
)

var (
	// ErrInvalidJWT is returned when an authentication token is malformed,
	// invalidly signed, or contains unexpected claims.
	ErrInvalidJWT = ErrForbidden("invalid token")
)

// CamelCaseToSnakeCase converts a string from CamelCase to snake_case.
func CamelCaseToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")

	return strings.ToLower(snake)
}

// Pointer is a generic helper function that returns a pointer to the value provided.
func Pointer[T any](v T) *T {
	return &v
}

// RelativeFilePath computes the relative path of a full path with respect to a base path,
// and converts the path to use forward slashes.
func RelativeFilePath(basePath, fullPath string) string {
	rel, err := filepath.Rel(basePath, fullPath)
	if err != nil {
		return fullPath
	}

	rel = filepath.ToSlash(rel)

	return rel
}

// Validate executes the validation rules defined in the provided 'schema' against the 'data' struct.
// It converts any validation issues into an InputError type for structured error handling.
func Validate(schema z.Shape, data any) (err error) {
	// Zod panics if struct is missing schema key
	// we don't want that
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
				return
			}

			err = fmt.Errorf("%v", r) //nolint:err113
			return
		}
	}()

	issueMap := z.Struct(schema).Validate(data)
	if len(issueMap) == 0 {
		return nil
	}

	ie := NewInputError()
	for key, issues := range issueMap {
		messages := make([]string, 0, len(issues))
		for _, issue := range issues {
			if issue == nil {
				continue
			}

			messages = append(messages, issue.Message)
		}

		ie.Add(CamelCaseToSnakeCase(key), strings.Join(messages, "\n"))
	}

	return ie
}

// String converts the input value into a string representation.
func String(v any) string {
	if v == nil {
		return ""
	}

	if s, ok := v.(string); ok {
		return s
	}

	if err, ok := v.(error); ok {
		return err.Error()
	}

	if s, ok := v.(fmt.Stringer); ok {
		return s.String()
	}

	return fmt.Sprintf("%v", v)
}

// Token creates a cryptographically secure random token.
func Token(lengthOpt ...int) (string, error) {
	length := 32
	if len(lengthOpt) > 0 {
		length = lengthOpt[0]
	}

	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Slug converts any string into a URL-friendly format.
func Slug(s string) string {
	return slug.Make(s)
}
