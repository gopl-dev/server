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

const flagIdent = "-"

const usageText = `
%s CLI (Env: %s; Host: %s)
                                                                          
Enter help to list all available commands
Enter help [command] to show description of given command
`

var (
	VerboseFlag = Flag{"v", "Verbose output"}
	YesFlag     = Flag{"y", "Force confirmation"}
)

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

type Arg struct {
	Name        string
	Description string
	Required    bool
	Default     string
}

type Flag struct {
	Name        string
	Description string
}

// Command represents a CLI command.
type Command struct {
	Name        string
	Alias       string
	Description string
	Args        []Arg
	Flags       []Flag
	Command     Runner

	argsCount         int
	requiredArgsCount int

	// Reflection cache
	reflectType  reflect.Type
	structFields map[string]int // arg/flag name -> field index
}

// Register adds a new Command definition to registry.
func (a *App) Register(cs ...Command) error {
	for _, c := range cs {
		if _, ok := a.commands[c.Name]; ok {
			return fmt.Errorf("command [%s] already registered", c.Name)
		}

		c.argsCount = len(c.Args)
		for _, arg := range c.Args {
			if arg.Required {
				c.requiredArgsCount++
			}
		}

		// Create reflection cache
		c.cacheReflection()

		a.commands[c.Name] = c
		if c.Alias != "" {
			if _, ok := a.commands[c.Alias]; ok {
				return fmt.Errorf("alias [%s] registered as command name", c.Alias)
			}
			a.aliases[c.Alias] = c.Name
		}

		err := validateCommand(c)
		if err != nil {
			return errors.New(aur.Bold(aur.Red(c.Name)).String() + ": " + err.Error())
		}
	}

	return nil
}

// cacheReflection prepares reflection metadata once.
func (c *Command) cacheReflection() {
	c.structFields = make(map[string]int)

	val := reflect.ValueOf(c.Command)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	c.reflectType = val.Type()

	for i := 0; i < c.reflectType.NumField(); i++ {
		f := c.reflectType.Field(i)
		if argTag := f.Tag.Get("arg"); argTag != "" {
			c.structFields[argTag] = i
		}
		if flagTag := f.Tag.Get("flag"); flagTag != "" {
			c.structFields[flagTag] = i
		}
	}
}

// Run executes a command by name with provided raw arguments.
func (a *App) Run(name string, args ...string) error {
	if name == "help" || name == "?" {
		return a.showHelp(args)
	}

	cmd, ok := a.commands[name]
	if !ok {
		aliasName, ok := a.aliases[name]
		if !ok {
			log.Println("Command " + name + " not found")
			a.printSimilarCommands(name)
			return nil
		}
		cmd = a.commands[aliasName]
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

// confirm asks for y/n confirmation.
func confirm(questionOpt ...string) (ok bool) {
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

// --- Reflection & argument helpers ---
func (c Command) prepareRunner(rawArgs []string) (Runner, error) {
	posArgs, foundFlags, namedParams := extractArgs(rawArgs, c.Args)

	val := reflect.New(c.reflectType).Elem()
	filled := make(map[string]bool)

	for k, v := range namedParams {
		if idx, ok := c.structFields[k]; ok {
			err := setFieldValue(val.Field(idx), v)
			if err != nil {
				return nil, err
			}
			filled[k] = true
		}
	}

	if err := bindPositionalArgs(c.Args, posArgs, val, c.structFields, filled); err != nil {
		return nil, err
	}
	if err := bindFlags(val, c.structFields, foundFlags); err != nil {
		return nil, err
	}
	if err := applyDefaults(c.Args, val, c.structFields, filled); err != nil {
		return nil, err
	}

	return val.Addr().Interface().(Runner), nil
}

func bindPositionalArgs(cmdArgs []Arg, args []string, val reflect.Value, fieldMap map[string]int, filled map[string]bool) error {
	curr := 0
	for _, argDef := range cmdArgs {
		if filled[argDef.Name] {
			continue
		}

		if curr < len(args) {
			if idx, ok := fieldMap[argDef.Name]; ok {
				err := setFieldValue(val.Field(idx), args[curr])
				if err != nil {
					return err
				}
				filled[argDef.Name] = true
			}
			curr++
		}
	}

	if curr < len(args) {
		return fmt.Errorf("too many positional arguments provided")
	}

	return nil
}

func bindFlags(val reflect.Value, fieldMap map[string]int, found map[string]bool) error {
	for name, present := range found {
		if idx, ok := fieldMap[name]; ok {
			err := setFieldValue(val.Field(idx), strconv.FormatBool(present))
			if err != nil {
				return fmt.Errorf("field for flag '%s': %w", name, err)
			}
		}
	}
	return nil
}

func applyDefaults(args []Arg, val reflect.Value, fieldMap map[string]int, filled map[string]bool) error {
	for _, arg := range args {
		if filled[arg.Name] {
			continue
		}
		if arg.Required {
			return fmt.Errorf("argument '%s' is required", arg.Name)
		}
		if arg.Default != "" {
			if idx, ok := fieldMap[arg.Name]; ok {
				err := setFieldValue(val.Field(idx), arg.Default)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func setFieldValue(field reflect.Value, value string) error {
	ft := field.Type()

	if ft.Kind() == reflect.Ptr {
		newVal := reflect.New(ft.Elem())
		if err := setFieldValue(newVal.Elem(), value); err != nil {
			return err
		}
		field.Set(newVal)
		return nil
	}

	if ft.Kind() == reflect.Slice {
		parts := strings.Split(value, ",")
		slice := reflect.MakeSlice(ft, 0, len(parts))
		for _, p := range parts {
			item := reflect.New(ft.Elem()).Elem()
			if err := setFieldValue(item, strings.TrimSpace(p)); err != nil {
				return err
			}
			slice = reflect.Append(slice, item)
		}
		field.Set(slice)
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
		bv, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(bv)
	default:
		return fmt.Errorf("unsupported type %s", ft.Kind())
	}
	return nil
}

func extractArgs(raw []string, expectedArgs []Arg) (pos []string, flags map[string]bool, named map[string]string) {
	flags = make(map[string]bool)
	named = make(map[string]string)

	for _, a := range raw {
		foundAsNamed := false

		for _, argDef := range expectedArgs {
			prefix := argDef.Name + "="
			if strings.HasPrefix(a, prefix) {
				parts := strings.SplitN(a, "=", 2)
				named[parts[0]] = parts[1]
				foundAsNamed = true
				break
			}
		}

		if foundAsNamed {
			continue
		}

		if strings.HasPrefix(a, flagIdent) {
			name := strings.TrimPrefix(a, flagIdent)
			flags[name] = true
		} else {
			pos = append(pos, a)
		}
	}
	return
}

func validateCommand(c Command) error {
	for _, a := range c.Args {
		if a.Required && a.Default != "" {
			return fmt.Errorf("arg '%s' cannot be required and have a default", a.Name)
		}
	}
	return nil
}
