package app

import (
	"path/filepath"
	"time"
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

func TimeNowPtr() *time.Time {
	return Pointer(time.Now())
}

func RelativeFilePath(basePath, fullPath string) string {
	rel, err := filepath.Rel(basePath, fullPath)
	if err != nil {
		return fullPath
	}

	rel = filepath.ToSlash(rel)
	return rel
}
