package cli

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	aur "github.com/logrusorgru/aurora"
)

const usageText = `
%s CLI (Env: %s; Host: %s)

Enter help to list all available commands
Enter help [command] to show description of given command
`

type App struct {
	Name     string
	Env      string
	commands map[string]Command
	aliases  map[string]string
}

func NewApp(name, env string) *App {
	return &App{
		Name:     name,
		Env:      env,
		commands: make(map[string]Command),
		aliases:  make(map[string]string),
	}
}

// Register adds commands to the application.
func (a *App) Register(cs ...Command) error {
	for _, c := range cs {
		if _, ok := a.commands[c.Name]; ok {
			return fmt.Errorf("command [%s] already registered", c.Name)
		}

		if c.Alias != "" && c.Alias != c.Name {
			if _, ok := a.commands[c.Alias]; ok {
				return fmt.Errorf("alias [%s] registered as command name", c.Alias)
			}
			a.aliases[c.Alias] = c.Name
		}

		c.cacheReflection()

		if err := c.prepareHelp(); err != nil {
			return fmt.Errorf("command [%s]: %w", c.Name, err)
		}

		if err := validateCommand(c); err != nil {
			return err
		}

		a.commands[c.Name] = c
	}

	return nil
}

// Run executes a command.
func (a *App) Run(name string, args ...string) error {
	if name == "help" || name == "?" {
		return a.showHelp(args)
	}

	cmd, ok := a.commands[name]
	if !ok {
		alias, ok := a.aliases[name]
		if !ok {
			log.Println("Command not found:", name)
			a.printSimilarCommands(name)
			return nil
		}
		cmd = a.commands[alias]
	}

	runner, err := cmd.prepareRunner(args)
	if err != nil {
		return err
	}

	return runner.Run(context.Background())
}

// PromptOrRun executes the CLI either from args or interactively.
func (a *App) PromptOrRun(args []string) {
	if len(args) > 1 {
		tail, err := splitArgs(strings.Join(args[2:], " "))
		if err != nil {
			log.Println(err)
			return
		}
		err = a.Run(os.Args[1], tail...)
		if err != nil {
			log.Println(err)
		}
		return
	}

	a.WaitForCommand()
}

// WaitForCommand starts the interactive CLI loop.
func (a *App) WaitForCommand() {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = aur.Red("unknown").String()
	}
	log.Printf(usageText, a.Name, a.Env, hostname)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		log.Print("> ")
		if !scanner.Scan() {
			return
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts, err := splitArgs(input)
		if err != nil {
			log.Println(err)
			continue
		}
		if len(parts) == 0 {
			continue
		}

		name := parts[0]
		var tail []string
		if len(parts) > 1 {
			tail = parts[1:]
		} else {
			tail = []string{}
		}

		if err := a.Run(name, tail...); err != nil {
			log.Println(err)
		}
	}
}

// showHelp prints help for all commands or a specific one.
func (a *App) showHelp(args []string) error {
	if len(args) > 0 {
		name := args[0]
		cmd, ok := a.commands[name]
		if !ok {
			alias := a.aliases[name]
			cmd, ok = a.commands[alias]
		}

		if !ok {
			fmt.Printf("Command '%s' not found\n", name)
			a.printSimilarCommands(name)
			return nil
		}

		a.printCommandHelp(cmd, true)
		return nil
	}

	for _, c := range a.commands {
		a.printCommandHelp(c, false)
		println("")
	}

	return nil
}

// printCommandHelp prints a single command's help.
func (a *App) printCommandHelp(cmd Command, verbose bool) {
	name := cmd.Alias
	if name == "" {
		name = cmd.Name
	}

	sigParts := make([]string, len(cmd.args))
	for i, ar := range cmd.args {
		n := ar.name
		if !ar.required {
			n = "[" + n + "]"
			n = aur.Gray(12, n).String()
		} else {
			if !ar.isFlag {
				n = aur.Blue(n).String()
			}
		}

		sigParts[i] = n
	}

	name = aur.Green(name).Bold().String()
	println(name + ": " + cmd.description)
	println(" Usage: " + aur.Green(name).Bold().String() + " " + strings.Join(sigParts, " "))

	if verbose {
		for _, ar := range cmd.args {
			argStr := "   " + ar.name
			if !ar.required {
				argStr += " (optional)"
			}

			argStr += " - " + aur.Italic(ar.description).String()
			if ar.defaultVal != "" {
				argStr += " (default: " + ar.defaultVal + ")"
			}
			println(argStr)
			if len(ar.help) > 0 {
				for _, h := range ar.help {
					println("     " + aur.Italic(h).String())
				}
				println("")
			}
		}
	}
}

// printSimilarCommands prints commands with similar names.
func (a *App) printSimilarCommands(name string) {
	similar := make([]string, 0)
	for _, c := range a.commands {
		if strings.Contains(c.Name, name) || strings.Contains(c.Alias, name) {
			similar = append(similar, "- "+c.Alias)
		}
	}

	if len(similar) > 0 {
		fmt.Println("Maybe you are looking for:")
		for _, s := range similar {
			fmt.Println(s)
		}
	}
}
