package cli

import (
	"context"
	"fmt"
	"time"
)

// NewSampleCommandWithFeaturesCmd returns a sample command that demonstrates:
// - positional arguments
// - named options (-key=value)
// - boolean flags
//
// Example (without optional arguments):
//
//	htf my-proj PROD -token=do-not-hack-me -vars='foo="bar",bar="beer"'
func NewSampleCommandWithFeaturesCmd() Command {
	return Command{
		Name: "how_to_features",

		// Alias is a shorter alternative name for the command.
		Alias: "htf",

		// Help is a list of lines shown in the help output.
		//
		// Conventions used here:
		// - Lines before the first "{argName}:" entry describe the command itself.
		// - A line starting with "{argName}:" begins the description for that argument.
		// - Subsequent lines continue the same argument’s description until another "{argName}:" line appears.
		Help: []string{
			// Handler description (until the first argument entry).
			"Publishes a project to a specified environment.",
			"It allows control over the deployed version and the target environment.",
			"Optional flags enable dry runs, detailed logging, and safety overrides during deployment.",

			// Argument descriptions.
			"project: The project to deploy.",
			"This line is part of the project description.",
			"This too, and all other lines until the next argument entry.",

			"env: The target environment for the deployment.",

			// Option descriptions.
			"-token: Authentication token used to authorize the deployment request.",
			"-tmp-path: Path to the temporary directory used during the deployment process.",
			"Make sure the path exists; no fallback is implemented yet.",
			"-vars: One or more variables passed to the build.",
			"Example: -vars='foo=bar,bar=beer'",
			"-t: Operation timeout in seconds.",
			"Set to 0 to disable the timeout.",

			// Flag descriptions.
			"-v: Enables detailed logging during the deployment process.",
			"-y: Requires explicit confirmation before executing the deployment.",
		},
		Handler: &SampleCommandWithFeaturesCmd{},
	}
}

// SampleCommandWithFeaturesCmd demonstrates a complex command with various argument types.
type SampleCommandWithFeaturesCmd struct {
	// Positional arguments: must appear in the order they are declared here.
	// Their values are assigned to the corresponding fields.
	// To make a positional argument optional, use a pointer type.

	// Project is a required positional argument.
	Project string `arg:"project"`

	// Env is an optional positional argument (default: STAGING).
	Env *string `arg:"env" default:"STAGING"`

	// Named options: provided as {name}={value}.
	// Unlike positional arguments, they can appear in any order.
	AuthToken string   `arg:"-token"`
	TmpPath   *string  `arg:"-tmp-path" default:"/tmp/"`
	Variables []string `arg:"-vars"`
	Timeout   *int     `arg:"-t" default:"300"`

	// Flags: simple switches. If a flag is present, its value becomes true.
	// They can appear in any order.
	Verbose bool `arg:"-v"`
	Confirm bool `arg:"-y"`
}

// Handle executes the command logic.
func (cmd *SampleCommandWithFeaturesCmd) Handle(_ context.Context) (err error) {
	fmt.Println("Hello")
	fmt.Println("We about to begin 🚀")

	to := (time.Duration(*cmd.Timeout) * time.Second).String()
	if *cmd.Timeout == 0 {
		to = "without timeout"
	}

	fmt.Println("Project:", cmd.Project)
	fmt.Println("Env:", *cmd.Env)
	fmt.Println("AuthToken:", cmd.AuthToken)
	fmt.Println("TmpPath:", *cmd.TmpPath)
	fmt.Println("Variables:", fmt.Sprintf("%v", cmd.Variables))
	fmt.Println("Timeout:", to)
	fmt.Println("Verbose:", cmd.Verbose)
	fmt.Println("Confirm:", cmd.Confirm)

	return nil
}
