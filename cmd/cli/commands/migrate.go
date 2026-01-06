package commands

import "context"

func init() {
	Register(Command{
		Name:        "migrate",
		Alias:       "mg",
		Description: "Migrate database to latest version (if new migrations available)",
		Command:     MigrateCmd{},
	})
}

type MigrateCmd struct{}

func (MigrateCmd) Run(ctx context.Context) (err error) {
	// ctx := context.TODO()
	// return app.MigrateDB(ctx)
	return nil
}
