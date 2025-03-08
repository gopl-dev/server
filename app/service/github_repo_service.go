package service

import (
	"strings"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/app/ds"
)

var (
	ErrRepoNotFoundByPath = app.ErrNotFound("Repository is not found by path")
	ErrInvalidRepoPath    = func(given string) error {
		return app.ErrUnprocessable("Invalid repository path, must be in format of '{user}/{repo}', but '" + given + "' given")
	}
)

func FindGitHubRepoByPath(path string) (m ds.GitHubRepo, err error) {
	//err = database.ORM().
	//	Model(&m).
	//	Where("path = ?", path).
	//	First()
	//
	//if err == pg.ErrNoRows {
	//	err = ErrRepoNotFoundByPath
	//}

	return
}

// NormalizeRepoPath removes simple mistakes from potentially valid format
// (copying correct path from  GitHub's html page will result invalid path when pasting (whitespace around "/"))
// return error if path in invalid format
func NormalizeRepoPath(path string) (string, error) {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 {
		return "", ErrInvalidRepoPath(path)
	}

	path = strings.TrimSpace(parts[0]) + "/" + strings.TrimSpace(parts[1])
	return path, nil
}

func UpdateGitHubRepo(m *ds.GitHubRepo) error {

	return nil
}
