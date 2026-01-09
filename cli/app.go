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

const usageText = `
%s CLI (Env: %s; Host: %s)

Enter help to list all available commands
Enter help [command] to show description of given command
`

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

// Run executes a command by name with given arguments.
func (a *App) Run(name string, args ...string) error {
	if name == "help" || name == "?" {
		return a.showHelp(args)
	}

	cmd, ok := a.commands[name]
	if !ok {
		alias, ok := a.aliases[name]
		if !ok {
			fmt.Println(fmt.Sprintf("%s", aur.Red("Command not found: "+name).String()+
				"\nType 'help' to get list of all available commands"))
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

// PromptOrRun executes the CLI either from provided args or starts interactive mode.
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
	defer rl.Close()

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

		args, err := splitArgs(line)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if len(args) == 0 {
			continue
		}

		name := args[0]
		tail := []string{}
		if len(args) > 1 {
			tail = args[1:]
		}

		if err := a.Run(name, tail...); err != nil {
			fmt.Println(err)
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
			fmt.Printf("Command '%s' not found\n", name)
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

		if ar.isFlag {
			flags = append(flags, n)
		} else if ar.isParam {
			params = append(params, n)
		} else {
			posArgs = append(posArgs, n)
		}
	}

	desc := cmd.description
	if !verbose {
		desc = strings.SplitN(desc, "\n", 2)[0]
	}

	// Build command header
	name := cmd.Name
	if cmd.Alias != "" {
		name = cmd.Name + " (" + cmd.Alias + ")"
	}
	help.WriteString(fmt.Sprintf("%s: %s\n", aur.Green(name).Bold(), desc))

	// Build usage line
	usageParts := append(posArgs, params...)
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
		// Flags: single line with description
		argLine := indent + ar.name + " " + aur.Italic(ar.description).String()

		// Add additional help lines if present
		if len(ar.help) > 0 {
			for _, h := range ar.help {
				argLine += "\n" + indent + "  " + aur.Gray(14, "• "+h).String()
			}
		}

		help.WriteString(argLine + "\n")
		return
	}

	// Arguments: multi-line with type info
	argLine := indent + aur.Blue(ar.name).String() + " " + aur.Gray(12, fmt.Sprintf("(type: %s)", ar.typ)).String()

	// Add default value if present
	if ar.defaultVal != "" {
		argLine += " " + aur.Gray(12, fmt.Sprintf("[default: %s]", ar.defaultVal)).String()
	}

	// Add optional marker
	if !ar.required {
		argLine += " " + aur.Yellow("(optional)").String()
	}

	// Build argument line and description
	help.WriteString(argLine + "\n")
	help.WriteString(fmt.Sprintf("%s%s\n", indent+"  ", aur.Italic(ar.description)))

	// Build additional help lines
	if len(ar.help) > 0 {
		for _, h := range ar.help {
			help.WriteString(fmt.Sprintf("%s• %s\n", indent+"    ", aur.Gray(14, h)))
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
			desc := strings.SplitN(c.description, "\n", 2)[0]

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
