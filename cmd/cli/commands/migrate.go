package commands

import (
	"context"

	"github.com/gopl-dev/server/cli"
)

func NewMigrateCmd() cli.Command {
	return cli.Command{
		Name:        "migrate",
		Alias:       "mg",
		Description: "Migrate database to latest version (if new migrations available)",
		Command:     migrateCmd{},
	}
}

type migrateCmd struct{}

func (migrateCmd) Run(ctx context.Context) (err error) {
	// ctx := context.TODO()
	// return app.MigrateDB(ctx)
	return nil
}
