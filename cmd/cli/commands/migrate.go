package commands

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

func migrateCommand(_ []string, _ Flags) (err error) {
	// ctx := context.TODO()
	// return app.MigrateDB(ctx)
	return nil
}
