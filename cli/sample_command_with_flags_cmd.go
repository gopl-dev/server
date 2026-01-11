//nolint:mnd
package cli

import (
	"context"
	"fmt"
	"time"

	aur "github.com/logrusorgru/aurora"
)

// NewSampleCommandWithFlagsCmd creates a new command instance demonstrating flag usage.
func NewSampleCommandWithFlagsCmd() Command {
	return Command{
		Name: "how_to_flags",
		Help: []string{
			"Demonstrates a CLI command that uses flags",
			"env: Deployment environment",
			"-v: Verbose output",
			"-y: Force confirmation",
		},
		Handler: &SampleCommandWithFlagsCmd{},
	}
}

// SampleCommandWithFlagsCmd ...
type SampleCommandWithFlagsCmd struct {
	Env     string `arg:"env" default:"STAGING"`
	Verbose bool   `arg:"-v"`
	Confirm bool   `arg:"-y"`
}

// Handle ...
func (c *SampleCommandWithFlagsCmd) Handle(_ context.Context) (err error) {
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
			switch {
			case wk == "Friday":
				Err("Don't do it on Friday!")

			case now.Hour() > 15:
				Err("You should start earlier!")
				Info("Consider postponing until tomorrow")

			default:
				fmt.Println(aur.Green(
					fmt.Sprintf("%s %s is a great time for deployment", wk, now.Format("15:04")),
				).String())
			}
		}

		if !Confirm("Are you sure?") {
			println("Operation canceled.")
			return nil
		}
	}

	// Actual deployment logic
	println("🚀 Deploying...")

	println("Successfully deployed to " + c.Env)
	return nil
}
