// Package main is the entry point for the CLI application.
package main

import (
	"log"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/cli"
	"github.com/gopl-dev/server/cmd/cli/commands"
)

func main() {
	conf := app.Config()
	cliApp := cli.NewApp(conf.App.Name, conf.App.Env)

	// Register core commands available in all environments
	err := cliApp.Register(
		commands.NewMigrateCmd(),

		// Uncomment to play with this demo commands
		// cli.NewSampleCommandWithSignatureCmd(),
		// cli.NewSampleCommandWithNamedParamsCmd(),
		// cli.NewSampleCommandWithFlagsCmd(),
		// cli.NewSampleCommandWithFeaturesCmd(),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register development/testing commands that should not be run in production
	if !conf.IsProductionEnv() {
		err = cliApp.Register(
			commands.NewResetDevEnvCmd(),
			commands.NewSeedDataCmd(),
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	cliApp.PromptOrRun()

	commands.CloseDB()
}
