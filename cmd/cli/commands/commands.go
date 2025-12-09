// Package commands ...
package commands

import (
	"log"
	"regexp"
	"strings"
)

const flagIdent = "-"
const confirmFlag = "y"

var (
	allCommands = map[string]Command{}
	allAliases  = map[string]string{}
)

// Command ...
type Command struct {
	Name    string
	Alias   string
	Help    []string
	Flags   []string
	Handler func(args []string, flags Flags) error
}

// Register adds a new Command definition to the application's global command registry.
func Register(c Command) {
	_, ok := allCommands[c.Name]
	if ok {
		log.Fatalf("Command [%s] already registered", c.Name)
	}

	allCommands[c.Name] = c

	if c.Alias != "" {
		_, ok := allCommands[c.Alias]
		if ok {
			log.Fatalf("Alias [%s] already registered as command", c.Alias)
		}

		name, ok := allAliases[c.Alias]
		if ok {
			log.Fatalf("Alias [%s] already taken by command [%s]", c.Alias, name)
		}

		allAliases[c.Alias] = c.Name
	}
}

// Run is the main execution function for the CLI.
func Run(name string, args ...string) (err error) {
	cmd, ok := allCommands[name]
	if !ok {
		aliasName, ok := allAliases[name]
		if !ok {
			log.Println("Command " + name + " not found")
			printSimilarCommands(name)

			return nil
		}

		cmd = allCommands[aliasName]
	}

	return cmd.Handler(resolveFlags(cmd, args))
}

// Flags is a type alias for a map of strings to boolean values.
// It is used to store and manage a set of active command-line flags.
type Flags map[string]bool

// Has checks if a specific flag is present in the Flags map.
// It handles the removal of the flag prefix ("-")
// before performing the lookup.
func (f Flags) Has(name string) bool {
	name = strings.TrimPrefix(name, flagIdent)
	_, ok := f[name]

	return ok
}

// HasConfirm is a convenience method to check specifically for the presence of the
// confirmation flag ("-y").
func (f Flags) HasConfirm() bool {
	return f.Has(confirmFlag)
}

// Add sets a flag as present (true) in the Flags map.
// It removes flag prefix before storing the name.
func (f Flags) Add(name string) {
	name = strings.TrimPrefix(name, flagIdent)
	f[name] = true
}

// ToArgs converts the stored flags back into a slice of command-line arguments.
func (f Flags) ToArgs() []string {
	args := []string{}
	for k := range f {
		args = append(args, flagIdent+k)
	}

	return args
}

func resolveFlags(cmd Command, args []string) ([]string, Flags) {
	flags := Flags{}
	filteredArgs := make([]string, 0)

iterateArgs:
	for _, v := range args {
		for _, f := range cmd.Flags {
			f = strings.TrimSpace(strings.Split(f, ":")[0])

			f = strings.TrimPrefix(f, flagIdent)
			if v == flagIdent+f {
				flags[f] = true

				continue iterateArgs
			}
		}

		filteredArgs = append(filteredArgs, v)
	}

	return filteredArgs, flags
}

const helpVerboseFlag = "v"

var cmdParamRE = regexp.MustCompile(`\[([^\[]*)]`)

func init() {
	Register(Command{
		Name:  "help",
		Alias: "?",
		Help: []string{
			"Display description of command",
			"? [command]: show help of given command",
		},
		Flags: []string{
			"v: Display additional information if available",
		},
		Handler: helpCommand,
	})
}

func helpCommand(args []string, flags Flags) (err error) {
	// specific command
	if len(args) > 0 {
		name := args[0]

		cmd, ok := allCommands[name]
		if !ok {
			aliasName, ok := allAliases[name]
			if !ok {
				println("Command '" + name + "' not found")
				printSimilarCommands(name)

				return
			}

			cmd = allCommands[aliasName]
		}

		printCommandHelp(cmd, true)

		return
	}

	// all commands
	verbose := flags.Has(helpVerboseFlag)
	for _, cmd := range allCommands {
		printCommandHelp(cmd, verbose)
	}

	return nil
}

func printCommandHelp(cmd Command, verbose bool) {
	name := cmd.Alias
	if name == "" {
		name = cmd.Name
	}

	var title string

	usage := make([]string, 0)
	desc := make([]string, 0)

	for i, line := range cmd.Help {
		if i == 0 {
			title = line

			continue
		}

		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, cmd.Alias) {
			usage = append(usage, line)

			continue
		}

		desc = append(desc, line)
	}

	if title == "" {
		title = cmd.Name
	}

	println(name + ": " + title)

	if !verbose {
		return
	}

	if len(desc) > 0 {
		println(strings.Join(desc, "\n"))
	}

	if len(usage) > 0 {
		println("Usage:")
	}

	for _, c := range usage {
		parts := strings.Split(c, ":")
		c = strings.TrimPrefix(parts[0], cmd.Alias)
		c = cmdParamRE.ReplaceAllString(c, "[$1]")

		c = cmd.Alias + c
		if len(parts) > 1 {
			c += ":" + strings.Join(parts[1:], ":")
		}

		println("   " + c)
	}

	if len(cmd.Flags) > 0 {
		println("Flags:")
	}

	for _, f := range cmd.Flags {
		parts := strings.Split(f, ":")
		f = parts[0]
		f = strings.TrimPrefix(f, flagIdent)

		f = flagIdent + f
		if len(parts) > 1 {
			f += ":" + strings.Join(parts[1:], ":")
		}

		println("   " + f)
	}

	println("")
}

func printSimilarCommands(name string) {
	similar := make([]string, 0)

	for _, c := range allCommands {
		match := strings.Contains(c.Name, name) || strings.Contains(c.Alias, name)
		if !match {
			for _, h := range c.Help {
				match = strings.Contains(h, name)
				if match {
					break
				}
			}
		}

		if match {
			help := ""
			if len(c.Help) > 0 {
				help = ": " + c.Help[0]
			}

			similar = append(similar, "- "+c.Alias+help)
		}
	}

	if len(similar) > 0 {
		println("Maybe you are looking for:")

		for _, s := range similar {
			println(s)
		}
	}
}
