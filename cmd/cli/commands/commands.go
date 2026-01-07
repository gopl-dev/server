// Package commands provides a lightweight CLI framework with support for
// positional arguments, named parameters (-key=val), and flags.
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

	"github.com/logrusorgru/aurora"
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

// Register adds a new Command definition to the global registry.
func Register(c Command) {
	if _, ok := allCommands[c.Name]; ok {
		log.Fatalf("Command [%s] already registered", c.Name)
	}

	c.argsCount = len(c.Args)
	for _, arg := range c.Args {
		if arg.Required {
			c.requiredArgsCount++
		}
	}

	allCommands[c.Name] = c

	if c.Alias != "" {
		if _, ok := allCommands[c.Alias]; ok {
			log.Fatalf("Alias [%s] registered as command", c.Alias)
		}
		allAliases[c.Alias] = c.Name
	}

	err := validateCommand(c)
	if err != nil {
		log.Fatal(aurora.Bold(aurora.Red(c.Name)).String() + ": " + err.Error())
	}
}

// Run executes a command by name with provided raw arguments.
func Run(name string, args ...string) error {
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

	return runner.Run(context.Background())
}

// prepareRunner creates a fresh instance of the command and fills it with data.
func (c Command) prepareRunner(rawArgs []string) (Runner, error) {
	positionalArgs, foundFlags, namedParams := c.extractArgs(rawArgs)

	// Create new instance of the command structure using reflection
	origValue := reflect.ValueOf(c.Command)
	if origValue.Kind() == reflect.Ptr {
		origValue = origValue.Elem()
	}
	newCmdValue := reflect.New(origValue.Type())
	v := newCmdValue.Elem()
	t := v.Type()

	filledFields := make(map[string]bool)

	// 1. Fill named parameters (-key=value)
	for key, val := range namedParams {
		err := fillFieldByTag(v, t, "arg", key, val)
		if err == nil {
			filledFields[key] = true
		}
	}

	// 2. Fill positional arguments (skipping those already filled by name)
	posIdx := 0
	for _, argDef := range c.Args {
		if filledFields[argDef.Name] {
			continue
		}
		if posIdx < len(positionalArgs) {
			err := fillFieldByTag(v, t, "arg", argDef.Name, positionalArgs[posIdx])
			if err != nil {
				return nil, err
			}
			filledFields[argDef.Name] = true
			posIdx++
		}
	}

	// 3. Fill boolean flags
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		flagName := f.Tag.Get("flag")
		if flagName != "" {
			_, isPresent := foundFlags[flagName]
			err := setFieldValue(v.Field(i), strconv.FormatBool(isPresent))
			if err != nil {
				return nil, err
			}
		}
	}

	// 4. Fill defaults and check required status
	for _, argDef := range c.Args {
		if !filledFields[argDef.Name] {
			if argDef.Required {
				return nil, fmt.Errorf("argument '%s' is required", argDef.Name)
			}
			if argDef.Default != "" {
				err := fillFieldByTag(v, t, "arg", argDef.Name, argDef.Default)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return newCmdValue.Interface().(Runner), nil
}

// extractArgs parses the raw string slice into positional, flags, and named maps.
func (c Command) extractArgs(rawArgs []string) (pos []string, flags map[string]bool, named map[string]string) {
	flags = make(map[string]bool)
	named = make(map[string]string)

	for _, arg := range rawArgs {
		if strings.HasPrefix(arg, flagIdent) {
			if strings.Contains(arg, "=") {
				parts := strings.SplitN(arg, "=", 2)
				named[parts[0]] = parts[1]
			} else {
				flags[strings.TrimPrefix(arg, flagIdent)] = true
			}
		} else {
			pos = append(pos, arg)
		}
	}
	return
}

func fillFieldByTag(v reflect.Value, t reflect.Type, tagName, tagValue, value string) error {
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Tag.Get(tagName) == tagValue {
			return setFieldValue(v.Field(i), value)
		}
	}
	return fmt.Errorf("tag %s:%s not found", tagName, tagValue)
}

// setFieldValue handles conversion from string to native Go types, including pointers and slices.
func setFieldValue(field reflect.Value, value string) error {
	targetType := field.Type()
	isPtr := targetType.Kind() == reflect.Ptr
	if isPtr {
		targetType = targetType.Elem()
	}

	var valToSet reflect.Value

	switch targetType.Kind() {
	case reflect.Slice:
		// Split by comma and trim spaces
		rawElements := strings.Split(value, ",")
		slice := reflect.MakeSlice(targetType, 0, len(rawElements))
		elemType := targetType.Elem()

		for _, rawElem := range rawElements {
			trimmed := strings.TrimSpace(rawElem)
			newElem := reflect.New(elemType).Elem()
			// Recursively set value for slice element (supports []int, []string, etc.)
			err := setFieldValue(newElem, trimmed)
			if err != nil {
				return err
			}
			slice = reflect.Append(slice, newElem)
		}
		valToSet = slice

	case reflect.String:
		valToSet = reflect.ValueOf(value)

	case reflect.Int, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("'%s' is not an integer", value)
		}
		valToSet = reflect.ValueOf(i).Convert(targetType)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("'%s' is not a boolean", value)
		}
		valToSet = reflect.ValueOf(b)

	default:
		return fmt.Errorf("unsupported type %s", targetType.Kind())
	}

	if isPtr {
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
			return fmt.Errorf("arg '%s' cannot be required and have a default", a.Name)
		}
	}
	return nil
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
		fmt.Println("\n> " + aurora.Bold(aurora.Green(question)).String() + "\n")
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
		if match {
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
