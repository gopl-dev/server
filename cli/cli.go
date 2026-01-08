// Package cli provides a lightweight CLI framework with support for
// positional arguments, named parameters (-key=val), and flags.
package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	aur "github.com/logrusorgru/aurora"
)

const usageText = `
%s CLI (Env: %s; Host: %s)

Enter help to list all available commands
Enter help [command] to show description of given command
`

type Runner interface {
	Run(ctx context.Context) error
}

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

type arg struct {
	name        string
	description string
	help        []string
	required    bool
	defaultVal  string
	isFlag      bool
	isParam     bool
}

// Command represents a CLI command.
type Command struct {
	Name        string
	Alias       string
	Help        []string
	description string
	args        []arg
	Command     Runner

	reflectType  reflect.Type
	structFields map[string]int // argName -> struct field index
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

// cacheReflection builds reflection metadata and auto-fills args from struct tags.
func (c *Command) cacheReflection() {
	c.structFields = make(map[string]int)
	c.args = nil

	val := reflect.ValueOf(c.Command)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	c.reflectType = val.Type()

	for i := 0; i < c.reflectType.NumField(); i++ {
		f := c.reflectType.Field(i)

		argTag := f.Tag.Get("arg")
		if argTag == "" {
			continue
		}

		a := arg{name: argTag}

		// 1) bool => flag
		if f.Type.Kind() == reflect.Bool {
			a.isFlag = true
		} else {
			// 2) non-flag + "-" prefix => named param
			if strings.HasPrefix(a.name, "-") {
				a.isParam = true
			}
		}

		// 3) required = not pointer + not flag
		if f.Type.Kind() != reflect.Ptr && !a.isFlag {
			a.required = true
		}

		// default from tag
		if def := f.Tag.Get("default"); def != "" {
			a.defaultVal = def
			a.required = false
		}

		c.args = append(c.args, a)
		c.structFields[a.name] = i
	}
}

func (c *Command) prepareHelp() error {
	if len(c.Help) == 0 {
		return errors.New("missing help text")
	}

	if len(c.args) == 0 {
		c.description = strings.Join(c.Help, "\n")
		return nil
	}

	argMap := make(map[string]*arg)
	for i, a := range c.args {
		argMap[a.name] = &c.args[i]
	}

	var commandLines []string
	var currentArg *arg

iterateHelp:
	for _, line := range c.Help {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		for name, arg := range argMap {
			prefix := name + ":"
			if strings.HasPrefix(line, prefix) {
				currentArg = arg
				currentArg.description = strings.TrimSpace(strings.TrimPrefix(line, prefix))
				currentArg.help = make([]string, 0)
				continue iterateHelp
			}
		}

		if currentArg == nil {
			commandLines = append(commandLines, line)
			continue
		}

		currentArg.help = append(currentArg.help, line)
	}

	for _, a := range c.args {
		if a.description == "" {
			return fmt.Errorf("description missing for '%s'", a.name)
		}
	}

	c.description = strings.Join(commandLines, "\n")
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
		tail := args[2:]
		err := a.Run(os.Args[1], tail...)
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

		err := a.Run(name, args...)
		if err != nil {
			log.Println(err)
		}
	}
}

// Confirm asks for y/n confirmation.
func Confirm(questionOpt ...string) (ok bool) {
	question := "Confirm?"
	yes := "y"
	yesAlt := "yes"

	if len(questionOpt) > 0 {
		question = questionOpt[0]
	}
	question += " y/n..."

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Println("\n> " + aur.Bold(aur.Green(question)).String() + "\n")
		scanner.Scan()
		input := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if input == yes || input == yesAlt {
			return true
		}

		return false
	}
}

// prepareRunner binds arguments and returns a Runner instance.
func (c Command) prepareRunner(rawArgs []string) (Runner, error) {
	posArgs, flags, named := extractArgs(rawArgs, c.args)

	val := reflect.New(c.reflectType).Elem()
	filled := make(map[string]bool)

	for k, v := range named {
		if idx, ok := c.structFields[k]; ok {
			if err := setFieldValue(val.Field(idx), v); err != nil {
				return nil, err
			}
			filled[k] = true
		}
	}

	for name, present := range flags {
		if idx, ok := c.structFields[name]; ok {
			if err := setFieldValue(val.Field(idx), strconv.FormatBool(present)); err != nil {
				return nil, err
			}
			filled[name] = true
		}
	}

	curr := 0
	for _, a := range c.args {
		if a.isFlag || a.isParam || filled[a.name] {
			continue
		}

		if curr < len(posArgs) {
			idx := c.structFields[a.name]
			if err := setFieldValue(val.Field(idx), posArgs[curr]); err != nil {
				return nil, err
			}
			filled[a.name] = true
			curr++
		}
	}

	for _, a := range c.args {
		if !filled[a.name] {
			if a.required {
				return nil, fmt.Errorf("argument '%s' is required", a.name)
			}
			if a.defaultVal != "" {
				idx := c.structFields[a.name]
				if err := setFieldValue(val.Field(idx), a.defaultVal); err != nil {
					return nil, err
				}
			}
		}
	}

	return val.Addr().Interface().(Runner), nil
}

// extractArgs splits raw args into positional, flags, and named parameters.
func extractArgs(raw []string, args []arg) (pos []string, flags map[string]bool, named map[string]string) {
	flags = make(map[string]bool)
	named = make(map[string]string)

	paramNames := make(map[string]struct{}) // from "-mood="
	flagNames := make(map[string]struct{})  // as "-y"

	for _, a := range args {
		if a.isParam {
			paramNames[a.name] = struct{}{}
		}
		if a.isFlag {
			flagNames[a.name] = struct{}{}
		}
	}

	for _, tok := range raw {
		// Named params: check "{argName}=" against known params
		if eq := strings.IndexByte(tok, '='); eq >= 0 {
			key := tok[:eq]
			if _, ok := paramNames[key]; ok {
				named[key] = tok[eq+1:]
				continue
			}
		}

		// Flags: token must match known flag name"
		if _, ok := flagNames[tok]; ok {
			flags[tok] = true
			continue
		}

		// 3) Positional as value
		pos = append(pos, tok)
	}

	return
}

// setFieldValue assigns string value to a reflected field.
func setFieldValue(field reflect.Value, value string) error {
	ft := field.Type()

	if ft.Kind() == reflect.Ptr {
		v := reflect.New(ft.Elem())
		if err := setFieldValue(v.Elem(), value); err != nil {
			return err
		}
		field.Set(v)
		return nil
	}

	switch ft.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int64:
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(v)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)
	default:
		return fmt.Errorf("unsupported type %s", ft.Kind())
	}

	return nil
}

func validateCommand(c Command) error {
	for _, a := range c.args {
		if a.required && a.defaultVal != "" {
			return fmt.Errorf("arg '%s' cannot be required and have a default", a.name)
		}

		if a.isFlag && !strings.HasPrefix(a.name, "-") {
			return fmt.Errorf("flag '%s' must start with '-'", a.name)
		}
	}

	return nil
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
	for i, arg := range cmd.args {
		n := arg.name
		if !arg.required {
			n = "[" + n + "]"
			n = aur.Gray(12, n).String()
		} else {
			if !arg.isFlag {
				n = aur.Blue(n).String()
			}
		}

		sigParts[i] = n
	}

	name = aur.Green(name).Bold().String()
	println(name + ": " + cmd.description)
	println(" Usage: " + aur.Green(name).Bold().String() + " " + strings.Join(sigParts, " "))

	if verbose {
		for _, arg := range cmd.args {
			argStr := "   " + arg.name
			if !arg.required {
				argStr += " (optional)"
			}

			argStr += " - " + aur.Italic(arg.description).String()
			if arg.defaultVal != "" {
				argStr += " (default: " + arg.defaultVal + ")"
			}
			println(argStr)
			if len(arg.help) > 0 {
				for _, h := range arg.help {
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
