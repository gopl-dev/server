package cli

import (
	"context"
	"fmt"
	"time"

	aur "github.com/logrusorgru/aurora"
)

func NewSampleCommandWithFlagsCmd() Command {
	return Command{
		Name:        "how_to_flags",
		Description: "Demonstrates a CLI command that uses flags",
		Args: []Arg{{
			Name:        "env",
			Description: "Deployment environment",
			Default:     "STAGING",
		}},
		Flags: []Flag{
			VerboseFlag,
			YesFlag,
		},
		Command: &SampleCommandWithFlagsCmd{},
	}
}

type SampleCommandWithFlagsCmd struct {
	Env     string `arg:"env"`
	Verbose bool   `flag:"v"`
	Confirm bool   `flag:"y"`
}

func (c *SampleCommandWithFlagsCmd) Run(ctx context.Context) (err error) {
	now := time.Now()
	defer func() {
		if c.Verbose {
			fmt.Printf("Duration: %s\n", time.Since(now))
		}
	}()

	if c.Verbose {
		println("Verbose output enabled")
	}

	if !c.Confirm {
		println("About to deploy to", c.Env)

		if c.Verbose {
			wk := time.Now().Weekday().String()
			if wk == "Friday" {
				println(aur.Red("Don't do it on Friday!").String())
			} else if time.Now().Hour() > 13 {
				println(aur.Red("You should start earlier!").String())
				println(aur.Red("Consider postponing until tomorrow").String()) // Fixed: Gerund "postponing"
			} else {
				fmt.Println(aur.Green(fmt.Sprintf("%s %s is a great time for deployment", wk, time.Now().Format("15:04"))).String())
			}
		}

		if !confirm("Are you sure?") {
			println("Operation canceled.")
			return nil
		}
	}

	// Actual deployment logic
	println("Deploying...")

	println("Successfully deployed to " + c.Env)
	return nil
}
