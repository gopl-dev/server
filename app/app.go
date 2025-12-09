// Package app ...
package app

import (
	"fmt"
	"path/filepath"
	"strings"

	z "github.com/Oudwins/zog"
)

const (
	// DevEnv is used for the local development environment.
	DevEnv = "DEV"

	// TestEnv is used for running automated tests and quality assurance (QA) checks.
	TestEnv = "TEST"

	// StagingEnv is a production-like environment used for final testing before a full public release.
	StagingEnv = "STAGING"

	// ReleaseEnv refers to the final production environment serving live users.
	ReleaseEnv = "RELEASE"
)

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

		ie.Add(key, strings.Join(messages, "\n"))
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
