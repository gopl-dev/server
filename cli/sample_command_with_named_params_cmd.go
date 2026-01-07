package cli

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

func NewSampleCommandWithNamedParamsCmd() Command {
	return Command{
		Name: "how_to_params",
		Help: []string{
			"A command that evaluates you based on your age",
			"name: The survivor's name",
			"-m: How you're feeling (We do care)",
			"-a: How many laps you've done around the sun",
		},
		Command: &SampleCommandWithNamedParamsCmd{},
	}
}

type SampleCommandWithNamedParamsCmd struct {
	Name string  `arg:"name"`
	Mood *string `arg:"-m"`
	Age  *int    `arg:"-a"`
}

func (cmd *SampleCommandWithNamedParamsCmd) Run(ctx context.Context) (err error) {
	mood := "Happy"
	if cmd.Mood != nil {
		mood = *cmd.Mood
	}

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

	fmt.Printf("Hello, %s %s %s!\n", mood, trait, cmd.Name)
	return nil
}
