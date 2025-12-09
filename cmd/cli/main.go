// Package main ...
package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"

	"github.com/gopl-dev/server/app"

	"github.com/gopl-dev/server/cmd/cli/commands"
)

const usageText = `
%s CLI (Env: %s; Host: %s)
                                                                          
Enter help to list all available commands
Enter help [command] to show description of given command
`

func main() {
	conf := app.Config()

	ctx := context.Background()

	db, err := app.NewPool(ctx)
	if err != nil {
		log.Println(err)

		return
	}
	defer db.Close()

	if len(os.Args) > 1 {
		args := os.Args[2:]

		err := commands.Run(os.Args[1], args...)
		if err != nil {
			log.Println(err)
		}
		return
	}

	hostname, _ := os.Hostname()
	log.Printf(usageText, conf.App.Name, conf.App.Env, hostname)

	scanner := bufio.NewScanner(os.Stdin)

	var input string

	for {
		log.Print("> ")
		scanner.Scan()
		input = scanner.Text()
		input = strings.TrimSpace(input)

		args := strings.Split(input, " ")
		if args[0] == "" {
			continue
		}

		cleanedArgs := make([]string, 0)

		for _, arg := range args {
			arg = strings.TrimSpace(arg)
			if arg == "" {
				continue
			}

			cleanedArgs = append(cleanedArgs, arg)
		}

		args = cleanedArgs

		name := args[0]
		if name == "" {
			continue
		}

		if len(args) > 1 {
			args = args[1:]
		} else {
			args = []string{}
		}

		err := commands.Run(name, args...)
		if err != nil {
			log.Println(err)
		}
	}
}
