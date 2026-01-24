//nolint:mnd
package cli

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// NewSampleCommandWithSignatureCmd creates a new command instance for demonstrating signature-based commands.
func NewSampleCommandWithSignatureCmd() Command {
	return Command{
		Name: "how_to_cli",
		Help: []string{
			"A command that judges you based on your age",
			"name: The victim's name",
			"mood: How you're feeling (as if we care)",
			"age: How many laps you've done around the sun",
		},
		Handler: &SampleCommandWithSignatureCmd{},
	}
}

// SampleCommandWithSignatureCmd demonstrates a command with positional arguments.
type SampleCommandWithSignatureCmd struct {
	Name string  `arg:"name"`
	Mood *string `arg:"mood" default:"Happy"`
	Age  *int    `arg:"age"`
}

// Handle executes the command logic.
func (cmd *SampleCommandWithSignatureCmd) Handle(_ context.Context) (err error) {
	var age int
	if cmd.Age != nil {
		age = *cmd.Age
	} else {
		age = rand.Intn(121) //nolint:gosec
	}

	var trait string

	switch {
	case age < 0:
		trait = "Edge-Case-Enthusiast"
	case age == 0:
		trait = "Still-Downloading"
	case age == 1:
		trait = "Initial-Commit"
	case age < 3:
		trait = "Unstable-Beta"
	case age < 6:
		trait = "Non-Stop-Notification"
	case age < 12:
		trait = "Feature-Request-Machine"
	case age < 19:
		trait = "Edgy-Front-End-Framework"
	case age < 30:
		trait = "Full-Stack-Dreamer"
	case age < 50:
		trait = "Pure-Source-Code"
	case age < 100:
		trait = "Legendary-Artifact"
	case age < 120:
		trait = "Legacy-System-Maintainer"
	case age == time.Now().Year():
		trait = "Professional-QA-Boundary-Tester"
	default:
		trait = "Vampire-In-Disguise"
	}

	fmt.Printf("Hello, %s %s %s!\n", *cmd.Mood, trait, cmd.Name)
	return nil
}
