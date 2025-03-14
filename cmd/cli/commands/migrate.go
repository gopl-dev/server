package commands

import (
	"context"

	"github.com/gopl-dev/server/app"
)

func init() {
	Register(Command{
		Name:  "migrate",
		Alias: "mg",
		Help: []string{
			"migrate database to latest version (if new migrations available)",
		},
		Handler: migrateCommand,
	})
}

func migrateCommand(args []string, flags Flags) (err error) {
	ctx := context.TODO()
	return app.MigrateDB(ctx)
}
