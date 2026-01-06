package commands

import (
	"context"
	"strings"

	aur "github.com/logrusorgru/aurora"
)

var helpCommand = Command{
	Name:        "help",
	Alias:       "?",
	Description: "Get help of given command, if command is not provided, a glimpse of all available commands will be shown",
	Args: []Arg{{
		Name:        "command",
		Description: "A command to get help with",
	}},

	Command: HelpCommandCmd{},
}

func init() {
	Register(helpCommand)
}

type HelpCommandCmd struct {
	Command *string `arg:"command"`
}

func (cmd HelpCommandCmd) Run(ctx context.Context) (err error) {
	// specific command
	if cmd.Command != nil {
		name := *cmd.Command

		c, ok := allCommands[name]
		if !ok {
			aliasName, ok := allAliases[name]
			if !ok {
				println("Command '" + name + "' not found")
				printSimilarCommands(name)

				return
			}

			c = allCommands[aliasName]
		}

		printCommandHelp(c, true)

		return
	}

	// all commands
	for _, c := range allCommands {
		if c.Name == helpCommand.Name {
			continue
		}
		printCommandHelp(c, false)
		println("")
	}

	return nil
}

func printCommandHelp(cmd Command, verbose bool) {
	name := cmd.Alias
	if name == "" {
		name = cmd.Name
	}

	sigParts := make([]string, len(cmd.Args)+len(cmd.Flags))
	for i, a := range cmd.Args {
		n := a.Name
		if !a.Required {
			n = "[" + n + "]"
			n = aur.Gray(12, n).String()
		} else {
			n = aur.Blue(n).String()
		}

		sigParts[i] = n
	}
	for i, a := range cmd.Flags {
		sigParts[i+len(cmd.Args)] = flagIdent + a.Name
	}

	name = aur.Green(name).Bold().String()
	println(name + ": " + cmd.Description)
	println(" Usage: " + aur.Green(name).Bold().String() + " " + strings.Join(sigParts, " "))

	if verbose {
		for _, a := range cmd.Args {
			argStr := "   " + a.Name
			if !a.Required {
				argStr += " (optional)"
			}

			argStr += " - " + aur.Italic(a.Description).String()
			if a.Default != "" {
				argStr += " (default: " + a.Default + ")"
			}
			println(argStr)
		}
		for _, f := range cmd.Flags {
			println("   " + flagIdent + f.Name + " - " + aur.Italic(f.Description).String())
		}
	}
}
