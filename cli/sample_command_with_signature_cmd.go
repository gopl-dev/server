package cli

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func NewSampleCommandWithSignatureCmd() Command {
	return Command{
		Name: "how_to_cli",
		Help: []string{
			"A command that judges you based on your age",
			"name: The victim's name",
			"mood: How you're feeling (as if we care)",
			"age: How many laps you've done around the sun",
		},
		Command: &SampleCommandWithSignatureCmd{},
	}
}

type SampleCommandWithSignatureCmd struct {
	Name string  `arg:"name"`
	Mood *string `arg:"mood" default:"Happy"`
	Age  *int    `arg:"age"`
}

func (cmd *SampleCommandWithSignatureCmd) Run(ctx context.Context) (err error) {

	var a int
	if cmd.Age != nil {
		a = *cmd.Age
	} else {
		a = rand.Intn(121)
	}

	var trait string

	switch {
	case a < 0:
		trait = "Edge-Case-Enthusiast"
	case a == 0:
		trait = "Still-Downloading"
	case a == 1:
		trait = "Initial-Commit"
	case a < 3:
		trait = "Unstable-Beta"
	case a < 6:
		trait = "Non-Stop-Notification"
	case a < 12:
		trait = "Feature-Request-Machine"
	case a < 19:
		trait = "Edgy-Front-End-Framework"
	case a < 30:
		trait = "Full-Stack-Dreamer"
	case a < 50:
		trait = "Pure-Source-Code"
	case a < 100:
		trait = "Legendary-Artifact"
	case a < 120:
		trait = "Legacy-System-Maintainer"
	case a == time.Now().Year():
		trait = "Professional-QA-Boundary-Tester"
	default:
		trait = "Vampire-In-Disguise"
	}

	fmt.Printf("Hello, %s %s %s!\n", *cmd.Mood, trait, cmd.Name)
	return nil
}
