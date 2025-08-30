package app

import (
	"fmt"
	"path/filepath"
	"strings"

	z "github.com/Oudwins/zog"
)

const (
	DevEnv     = "DEV"
	TestEnv    = "TEST"
	StagingEnv = "STAGING"
	ReleaseEnv = "RELEASE"
)

func Pointer[T any](v T) *T {
	return &v
}

func RelativeFilePath(basePath, fullPath string) string {
	rel, err := filepath.Rel(basePath, fullPath)
	if err != nil {
		return fullPath
	}

	rel = filepath.ToSlash(rel)
	return rel
}

func Validate(schema z.Shape, data any) (err error) {
	// Zod panics if struct is missing schema key
	// we don't want that
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
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
