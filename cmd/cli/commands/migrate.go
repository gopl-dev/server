package commands

import (
	"context"
	"log"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/cli"
)

func NewMigrateCmd() cli.Command {
	return cli.Command{
		Name:  "migrate",
		Alias: "mg",
		Help: []string{
			"Migrate database to latest version (if new migrations available)",
		},
		Command: migrateCmd{},
	}
}

type migrateCmd struct{}

func (migrateCmd) Run(ctx context.Context) (err error) {
	db, err := app.NewDB(ctx)
	if err != nil {
		log.Println(err)
		return
	}
	defer db.Close()

	err = app.MigrateDB(ctx, db)
	if err != nil {
		log.Println(err)
		return
	}

	return nil
}
