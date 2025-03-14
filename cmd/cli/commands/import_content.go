package commands

import (
	"context"

	"github.com/gopl-dev/server/content"
)

func init() {
	Register(Command{
		Name:  "import-content-from-gh",
		Alias: "imp",
		Help: []string{
			"Import content from GitHub repository",
			"imp",
		},
		Handler: importContentFromGitHub,
	})
}

func importContentFromGitHub(args []string, flags Flags) (err error) {
	ctx := context.TODO()
	err = content.ImportFromGitHub(ctx)
	if err != nil {
		return
	}

	println("Done!")
	return
}
