package cli

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
	aur "github.com/logrusorgru/aurora"
)

var (
	errCommandNameTaken = errors.New("command name already registered")
	errAliasIsCommand   = errors.New("alias is already used as command name")
	errAliasNameTaken   = errors.New("alias already taken")
)

const usageText = `
%s CLI (Env: %s; Host: %s)

Enter help to list all available commands
Enter help [command] to show description of given command
`

// App ...
type App struct {
	Name      string
	Env       string
	commands  map[string]Command
	aliases   map[string]string
	helpCache map[string]string
}

// NewApp creates a new CLI application instance.
func NewApp(name, env string) *App {
	return &App{
		Name:      name,
		Env:       env,
		commands:  make(map[string]Command),
		aliases:   make(map[string]string),
		helpCache: make(map[string]string),
	}
}

// Register adds commands to the application.
func (a *App) Register(cs ...Command) error {
	for _, c := range cs {
		if _, ok := a.commands[c.Name]; ok {
			return fmt.Errorf("%s: %w", c.Name, errCommandNameTaken)
		}

		if c.Alias != "" && c.Alias != c.Name {
			if _, ok := a.commands[c.Alias]; ok {
				return fmt.Errorf("%w (alias '%s' of '%s')", errAliasIsCommand, c.Alias, c.Name)
			}
			if conflictName, ok := a.aliases[c.Alias]; ok {
				conflicted := a.commands[conflictName]
				return fmt.Errorf("%w (alias '%s' of '%s' is taken by '%s')", errAliasNameTaken, c.Alias, c.Name, conflicted.Name)
			}
			a.aliases[c.Alias] = c.Name
		}

		c.cacheReflection()

		err := c.prepareHelp()
		if err != nil {
			return fmt.Errorf("command [%s]: %w", c.Name, err)
		}

		err = validateCommand(c)
		if err != nil {
			return err
		}

		a.commands[c.Name] = c
	}

	return nil
}

// Run executes a command by name with given arguments.
func (a *App) Run(name string, args ...string) error {
	if name == "help" || name == "?" {
		return a.showHelp(args)
	}

	cmd, ok := a.commands[name]
	if !ok {
		alias, ok := a.aliases[name]
		if !ok {
			Err("Handler '%s' not found", name)
			a.printSimilarCommands(name)
			return nil
		}
		cmd = a.commands[alias]
	}

	handler, err := cmd.prepareHandler(args)
	if err != nil {
		return err
	}

	return handler.Handle(context.Background())
}

// PromptOrRun executes the CLI either from provided args or starts interactive mode.
func (a *App) PromptOrRun(args []string) {
	if len(args) > 1 {
		tail := splitArgs(strings.Join(args[2:], " "))
		err := a.Run(os.Args[1], tail...)
		if err != nil {
			log.Println(err)
		}
		return
	}

	a.WaitForCommand()
}

// WaitForCommand starts the interactive CLI loop waiting for user input.
func (a *App) WaitForCommand() {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = aur.Red("unknown").String()
	}
	fmt.Printf(usageText, a.Name, a.Env, hostname)

	rl, err := readline.NewEx(&readline.Config{
		Prompt:            "> ",
		Stdin:             os.Stdin,
		Stdout:            os.Stdout,
		Stderr:            os.Stderr,
		HistorySearchFold: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = rl.Close()
	}()

	for {
		line, err := rl.Readline()
		if err != nil {
			if !errors.Is(err, readline.ErrInterrupt) {
				fmt.Println("[ERROR] read input: " + err.Error())
			}
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		err = readline.AddHistory(line)
		if err != nil {
			fmt.Println("[ERROR] saving history: " + err.Error())
		}

		args := splitArgs(line)
		if len(args) == 0 {
			continue
		}

		name := args[0]
		tail := []string{}
		if len(args) > 1 {
			tail = args[1:]
		}

		err = a.Run(name, tail...)
		if err != nil {
			fmt.Println(aur.Red(err).String())
		}
	}
}

// showHelp handles help command, showing all commands or specific command help.
func (a *App) showHelp(args []string) error {
	if len(args) > 0 {
		name := args[0]
		cmd, ok := a.commands[name]
		if !ok {
			alias := a.aliases[name]
			cmd, ok = a.commands[alias]
		}

		if !ok {
			fmt.Printf("Handler '%s' not found\n", name)
			a.printSimilarCommands(name)
			return nil
		}

		a.printCommandHelp(cmd, true)
		return nil
	}

	for _, c := range a.commands {
		a.printCommandHelp(c, false)
	}

	return nil
}

// printCommandHelp prints the help text for a single command.
func (a *App) printCommandHelp(cmd Command, verbose bool) {
	if cachedHelp, ok := a.helpCache[cmd.Name]; ok && verbose {
		fmt.Println(cachedHelp)
		return
	}

	var help strings.Builder

	// Build usage signature
	var posArgs []string
	var flags []string
	var params []string

	for _, ar := range cmd.args {
		n := ar.name
		if !ar.required {
			n = "[" + n + "]"
		}

		switch {
		case ar.isFlag:
			flags = append(flags, n)

		case ar.isParam:
			params = append(params, n)

		default:
			posArgs = append(posArgs, n)
		}
	}

	desc := cmd.description
	if !verbose {
		desc = strings.SplitN(desc, "\n", 2)[0] //nolint:mnd
	}

	// Build command header
	name := cmd.Name
	if cmd.Alias != "" {
		name = cmd.Name + " (" + cmd.Alias + ")"
	}
	help.WriteString(fmt.Sprintf("%s: %s\n", aur.Green(name).Bold(), desc))

	// Build usage line
	usageParts := make([]string, 0)
	usageParts = append(usageParts, posArgs...)
	usageParts = append(usageParts, params...)
	usageParts = append(usageParts, flags...)
	if cmd.Alias != "" {
		name = cmd.Alias
	}
	help.WriteString(fmt.Sprintf("Usage: %s %s\n\n", aur.Green(name).Bold(), strings.Join(usageParts, " ")))

	if !verbose {
		fmt.Print(help.String())
		return
	}

	// Group arguments by type
	var args []arg
	var flagArgs []arg

	for _, ar := range cmd.args {
		if ar.isFlag {
			flagArgs = append(flagArgs, ar)
		} else {
			args = append(args, ar)
		}
	}

	// Build Arguments section
	if len(args) > 0 {
		help.WriteString(fmt.Sprintf("%s\n", aur.Bold(aur.Cyan("Arguments:"))))
		for _, ar := range args {
			a.buildArgumentDetail(&help, ar, "  ")
		}
		help.WriteString("\n")
	}

	// Build Flags section
	if len(flagArgs) > 0 {
		help.WriteString(fmt.Sprintf("%s\n", aur.Bold(aur.Cyan("Flags:"))))
		for _, ar := range flagArgs {
			a.buildArgumentDetail(&help, ar, "  ")
		}
	}

	result := help.String()
	a.helpCache[cmd.Name] = result
	fmt.Print(result)
}

// buildArgumentDetail builds detailed help for a single argument into the output buffer.
func (a *App) buildArgumentDetail(help *strings.Builder, ar arg, indent string) {
	if ar.isFlag {
		var b strings.Builder

		// Flags: single line with description
		b.WriteString(indent)
		b.WriteString(ar.name)
		b.WriteString(" ")
		b.WriteString(aur.Italic(ar.description).String())

		// Add additional help lines if present
		if len(ar.help) > 0 {
			for _, h := range ar.help {
				b.WriteString("\n")
				b.WriteString(indent)
				b.WriteString("  ")
				b.WriteString(Gray("• " + h))
			}
		}

		argLine := b.String()

		help.WriteString(argLine + "\n")
		return
	}

	// Arguments: multi-line with type info
	argLine := indent + aur.Blue(ar.name).String() + " " + Gray("(type: %s)", ar.typ)

	// Add default value if present
	if ar.defaultVal != "" {
		argLine += " " + Gray("[default: %s]", ar.defaultVal)
	}

	// Add optional marker
	if !ar.required {
		argLine += " " + aur.Yellow("(optional)").String()
	}

	// Build argument line and description
	help.WriteString(argLine + "\n")
	_, _ = fmt.Fprintf(help, "%s%s\n", indent+"  ", aur.Italic(ar.description))

	// Build additional help lines
	if len(ar.help) > 0 {
		for _, h := range ar.help {
			_, _ = fmt.Fprintf(help, "%s• %s\n", indent+"    ", Gray(h))
		}
	}

	help.WriteString("\n")
}

// printSimilarCommands prints commands with names similar to the given name.
func (a *App) printSimilarCommands(name string) {
	similar := make([]string, 0)
	for _, c := range a.commands {
		if strings.Contains(c.Name, name) || strings.Contains(c.Alias, name) {
			cName := c.Name
			if c.Alias != "" {
				cName += " (" + c.Alias + ")"
			}
			desc := strings.SplitN(c.description, "\n", 2)[0] //nolint:mnd

			similar = append(similar, "- "+fmt.Sprintf("%s: %s", aur.Blue(cName), desc))
		}
	}

	if len(similar) > 0 {
		fmt.Println("Maybe you are looking for:")
		for _, s := range similar {
			fmt.Println(s)
		}
	}
}
