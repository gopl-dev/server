// Package main ...
package main

import (
	"log"
	"os"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/cli"
	"github.com/gopl-dev/server/cmd/cli/commands"
)

func main() {
	conf := app.Config()

	cliApp := cli.NewApp(conf.App.Name, conf.App.Env)

	err := cliApp.Register(
		commands.NewMigrateCmd(),

		cli.NewSampleCommandWithSignatureCmd(),
		cli.NewSampleCommandWithNamedParamsCmd(),
		cli.NewSampleCommandWithFlagsCmd(),
		cli.NewSampleCommandWithFeaturesCmd(),
	)
	if err != nil {
		log.Fatal(err)
	}

	cliApp.PromptOrRun(os.Args)
}
