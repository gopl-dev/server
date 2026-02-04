// Package app ...
package app

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	z "github.com/Oudwins/zog"
	"github.com/gosimple/slug"
	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

const (
	// DevEnv is used for the local development environment.
	DevEnv = "dev"

	// TestEnv is used for running automated tests and quality assurance (QA) checks.
	TestEnv = "test"

	// StagingEnv is a production-like environment used for final testing before a full public release.
	StagingEnv = "staging"

	// ProductionEnv refers to the final production environment serving live users.
	ProductionEnv = "production"
)

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

var mdRenderer = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
	),
)

var htmlPolicy = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.RequireNoFollowOnLinks(true)
	p.RequireNoReferrerOnLinks(true)
	return p
}()

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

// MarkdownToHTML renders Markdown input into safe HTML.
func MarkdownToHTML(in string) (out string, err error) {
	var buf bytes.Buffer

	err = mdRenderer.Convert([]byte(in), &buf)
	if err != nil {
		return
	}

	out = htmlPolicy.Sanitize(buf.String())
	return
}

// HumanTime formats a timestamp into a human-readable relative time string.
//
// The function compares the given time with the current time and returns
// a concise, user-friendly representation (e.g. "just now", "yesterday, 15:04").
// An optional reference time may be provided for deterministic output.
func HumanTime(t time.Time, relOpt ...time.Time) string {
	now := time.Now()
	if len(relOpt) > 0 {
		now = relOpt[0]
	}
	d := now.Sub(t)

	switch {
	case d < 30*time.Second:
		return "just now"

	case d < 3*time.Minute:
		return "few minutes ago"

	case d < 10*time.Minute:
		return "five minutes ago"
	}

	ty, tm, td := t.Date()
	ny, nm, nd := now.Date()

	if ty == ny && tm == nm && td == nd {
		return t.Format("15:04")
	}

	yesterday := now.AddDate(0, 0, -1)
	yy, ym, yd := yesterday.Date()
	if ty == yy && tm == ym && td == yd {
		return "yesterday, " + t.Format("15:04")
	}

	if ty == ny {
		return t.Format("Jan 2, 15:04")
	}

	return t.Format("Jan 2, 2006; 15:04")
}
