// Package commands ...
package commands

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"

	aur "github.com/logrusorgru/aurora"
)

const flagIdent = "-"

var (
	verboseFlag = Flag{"v", "Verbose output"}
	yesFlag     = Flag{"y", "Force confirmation"}
)

var (
	allCommands = map[string]Command{}
	allAliases  = map[string]string{}
)

type Runner interface {
	Run(ctx context.Context) error
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

// Command ...
type Command struct {
	Name        string
	Alias       string
	Description string
	Args        []Arg
	Flags       []Flag
	Command     Runner

	argsCount         int
	requiredArgsCount int
}

// Register adds a new Command definition to the application's global command registry.
func Register(c Command) {
	_, ok := allCommands[c.Name]
	if ok {
		log.Fatalf("Command [%s] already registered", c.Name)
	}

	c.argsCount = len(c.Args)
	c.requiredArgsCount = 0
	for _, arg := range c.Args {
		if arg.Required {
			c.requiredArgsCount++
		}
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

	err := validateCommand(c)
	if err != nil {
		log.Fatal(aur.Bold(aur.Red(c.Name)).String() + ": " + err.Error())
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

	runner, err := cmd.prepareRunner(args)
	if err != nil {
		return err
	}

	ctx := context.Background()
	return runner.Run(ctx)
}

func confirm(questionOpt ...string) (ok bool) {
	question := "Confirm?"
	yes := "y"
	yesAlt := "yes"

	if len(questionOpt) > 0 {
		question = questionOpt[0]
	}
	question += " y/n..."

	scanner := bufio.NewScanner(os.Stdin)
	var input string
	for {
		println("\n> " + aur.Bold(aur.Green(question)).String() + "\n")
		scanner.Scan()
		input = scanner.Text()
		input = strings.ToLower(strings.TrimSpace(input))
		if input == yes || input == yesAlt {
			return true
		}

		return false
	}
}

func printSimilarCommands(name string) {
	similar := make([]string, 0)

	for _, c := range allCommands {
		match := strings.Contains(c.Name, name) || strings.Contains(c.Alias, name)
		if !match {
			//for _, h := range c.Help {
			//	match = strings.Contains(h, name)
			//	if match {
			//		break
			//	}
			//}
		}

		if match {
			help := ""
			//if len(c.Help) > 0 {
			//	help = ": " + c.Help[0]
			//}

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

func (c Command) prepareRunner(rawArgs []string) (Runner, error) {
	positionalArgs, foundFlags := c.extractFlags(rawArgs)

	if len(positionalArgs) > c.argsCount {
		return nil, fmt.Errorf("too many arguments: expected max %d, got %d", c.argsCount, len(positionalArgs))
	}
	if len(positionalArgs) < c.requiredArgsCount {
		return nil, fmt.Errorf("not enough arguments: expected at least %d, got %d", c.requiredArgsCount, len(positionalArgs))
	}

	origValue := reflect.ValueOf(c.Command)
	if origValue.Kind() == reflect.Ptr {
		origValue = origValue.Elem()
	}
	newCmdValue := reflect.New(origValue.Type())
	v := newCmdValue.Elem()
	t := v.Type()

	for i, argValue := range positionalArgs {
		argDef := c.Args[i]
		if err := fillFieldByTag(v, t, "arg", argDef.Name, argValue); err != nil {
			return nil, err
		}
	}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		flagName := f.Tag.Get("flag")
		if flagName != "" {
			_, isPresent := foundFlags[flagName]
			if err := setFieldValue(v.Field(i), strconv.FormatBool(isPresent)); err != nil {
				return nil, err
			}
		}
	}

	if len(positionalArgs) < len(c.Args) {
		for i := len(positionalArgs); i < len(c.Args); i++ {
			argDef := c.Args[i]
			if argDef.Default != "" {
				if err := fillFieldByTag(v, t, "arg", argDef.Name, argDef.Default); err != nil {
					return nil, fmt.Errorf("default value error: %w", err)
				}
			}
		}
	}

	return newCmdValue.Interface().(Runner), nil
}

func (c Command) extractFlags(rawArgs []string) ([]string, map[string]bool) {
	positional := make([]string, 0)
	foundFlags := make(map[string]bool)

	for _, arg := range rawArgs {
		if strings.HasPrefix(arg, flagIdent) {
			// Убираем "-" и сохраняем
			flagName := strings.TrimPrefix(arg, flagIdent)
			foundFlags[flagName] = true
		} else {
			positional = append(positional, arg)
		}
	}
	return positional, foundFlags
}

func fillFieldByTag(v reflect.Value, t reflect.Type, tagName, tagValue, value string) error {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Tag.Get(tagName) == tagValue {
			return setFieldValue(v.Field(i), value)
		}
	}

	return fmt.Errorf("field with %s:\"%s\" not found", tagName, tagValue)
}

func setFieldValue(field reflect.Value, value string) error {
	targetType := field.Type()
	if targetType.Kind() == reflect.Ptr {
		targetType = targetType.Elem()
	}

	var valToSet reflect.Value

	switch targetType.Kind() {
	case reflect.String:
		valToSet = reflect.ValueOf(value)

	case reflect.Int, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid integer", value)
		}
		valToSet = reflect.ValueOf(i).Convert(targetType)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("'%s' is not a valid boolean", value)
		}
		valToSet = reflect.ValueOf(b)

	default:
		return fmt.Errorf("unsupported type %s", targetType.Kind())
	}

	if field.Kind() == reflect.Ptr {
		ptr := reflect.New(targetType)
		ptr.Elem().Set(valToSet)
		field.Set(ptr)
	} else {
		field.Set(valToSet)
	}

	return nil
}

func validateCommand(c Command) error {
	for _, a := range c.Args {
		if a.Required && a.Default != "" {
			return fmt.Errorf("arg '%s' should not be required or not have a default value", a.Name)
		}
	}

	return nil
}
