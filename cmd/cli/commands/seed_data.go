package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gopl-dev/server/app"
	"github.com/gopl-dev/server/cli"
	"github.com/gopl-dev/server/test/seed"
)

var (
	errInvalidDataName = errors.New("invalid seed data")
)

var seedAvailableData = []string{
	"all", "users", "books",
}

// NewSeedDataCmd ...
func NewSeedDataCmd() cli.Command {
	return cli.Command{
		Name:  "seed_data",
		Alias: "sd",
		Help: []string{
			"Seeds DB with test data",
			"data: Data to seed",
			fmt.Sprintf("Options: %s", strings.Join(seedAvailableData, ", ")),
			"count: Amount of data to seed",
		},
		Handler: &seedDataCmd{},
	}
}

type seedDataCmd struct {
	Data  *string `arg:"data" default:"all"`
	Count *int    `arg:"count" default:"100"`
}

func (cmd *seedDataCmd) Handle(ctx context.Context) (err error) {
	s := seed.New(db())

	if cmd.Data == nil {
		cmd.Data = app.Pointer("all") // TODO: new("all") soon
	}

	switch *cmd.Data {
	case "all":
		err = s.All(ctx, *cmd.Count)
	case "users":
		err = s.Users(ctx, *cmd.Count)
	case "books":
		err = s.Books(ctx, *cmd.Count)
	default:
		err = errInvalidDataName
	}
	if err != nil {
		return err
	}

	fmt.Println("Done!")
	return nil
}
