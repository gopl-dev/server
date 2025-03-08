// GitHub repository
// https://docs.github.com/en/rest/reference/repos#get-a-repository
// GET https://api.github.com/repos/{user}/{name}

// TODO: Consider refactoring this as a "data source" and making it one of multiple sources.
// (Since GitHub is the sole source for now, I'm leaving it as a top-level entity)

package ds

import (
	"time"
)

type ImportStatus int

const (
	ImportSuccess ImportStatus = iota + 1
	ImportFailed
)

const GithubAddr = "https://github.com/"

type GitHubRepo struct {
	ID int64 `json:"id"`

	// Path represents "{account}/{repo}" on GitHub.
	// Note that this is equivalent to "full_name" in the GitHub API.
	Path string `json:"path"`

	// DefaultBranch is needed to compose links (e.g., "view source", "edit", "commits").
	DefaultBranch string `json:"default_branch"`

	// HTMLURL is "https://github.com/{user}/{repo}"
	// and is populated from the html_url property of the GitHub API.
	HTMLURL string `json:"html_url" pg:"html_url"`

	CloneURL     string       `json:"-"`
	ImportStatus ImportStatus `json:"import_status"`
	ImportLog    string       `json:"import_log"`

	// Secret is used to authenticate the repository from GitHub.
	// It is a random string given to the user who creates the repository
	// and must be set in the repository's push webhook on GitHub.
	// https://github.com/{account}/{repo}/settings/hooks
	// The Secret is hashed using bcrypt and is available only once.
	Secret string `json:"-"`

	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
	ImportedAt *time.Time `json:"imported_at"`
}
