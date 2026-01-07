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

// --- ORCHESTRATION ---

// prepareRunner creates a fresh instance of the command and fills it with data.
func (c Command) prepareRunner(rawArgs []string) (Runner, error) {
	posArgs, foundFlags, namedParams := extractArgs(rawArgs, c.Args)

	// Initialize fresh instance of the struct
	val, typ := c.createNewInstance()
	filled := make(map[string]bool)

	// Phase 1: Explicitly named params (-key=val)
	for k, v := range namedParams {
		err := fillField(val, typ, "arg", k, v)
		if err == nil {
			filled[k] = true
		}
	}

	// Phase 2: Positional arguments
	err := c.bindPositional(val, typ, posArgs, filled)
	if err != nil {
		return nil, err
	}

	// Phase 3: Boolean flags
	err = c.bindFlags(val, typ, foundFlags)
	if err != nil {
		return nil, err
	}

	// Phase 4: Defaults and validation
	err = c.applyDefaults(val, typ, filled)
	if err != nil {
		return nil, err
	}

	return val.Addr().Interface().(Runner), nil
}

// --- BINDING LOGIC ---

func (c Command) bindPositional(v reflect.Value, t reflect.Type, args []string, filled map[string]bool) error {
	curr := 0
	for _, argDef := range c.Args {
		if filled[argDef.Name] {
			continue
		}

		if curr < len(args) {
			err := fillField(v, t, "arg", argDef.Name, args[curr])
			if err != nil {
				return err
			}
			filled[argDef.Name] = true
			curr++
		}
	}

	if curr < len(args) {
		return fmt.Errorf("too many positional arguments provided")
	}

	return nil
}

// bindFlags set boolean fields based on existence in input
func (c Command) bindFlags(v reflect.Value, t reflect.Type, found map[string]bool) error {
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		name := f.Tag.Get("flag")
		if name == "" {
			continue
		}

		// We know 'found[name]' is a bool, we format it to string
		// so setFieldValue can parse it back to the struct field.
		isPresent := strconv.FormatBool(found[name])
		err := setFieldValue(v.Field(i), isPresent)
		if err != nil {
			// This will catch situations where a 'flag' tag is placed on a non-bool field
			return fmt.Errorf("field '%s' with tag flag:'%s': %w", f.Name, name, err)
		}
	}

	return nil
}

func (c Command) applyDefaults(v reflect.Value, t reflect.Type, filled map[string]bool) error {
	for _, arg := range c.Args {
		if filled[arg.Name] {
			continue
		}

		if arg.Required {
			return fmt.Errorf("argument '%s' is required", arg.Name)
		}

		if arg.Default != "" {
			err := fillField(v, t, "arg", arg.Name, arg.Default)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// --- REFLECTION HELPERS ---

func (c Command) createNewInstance() (reflect.Value, reflect.Type) {
	orig := reflect.ValueOf(c.Command)
	if orig.Kind() == reflect.Ptr {
		orig = orig.Elem()
	}
	newVal := reflect.New(orig.Type()).Elem()
	return newVal, newVal.Type()
}

func fillField(v reflect.Value, t reflect.Type, tag, tagVal, val string) error {
	for i := 0; i < t.NumField(); i++ {
		if t.Field(i).Tag.Get(tag) == tagVal {
			return setFieldValue(v.Field(i), val)
		}
	}
	return fmt.Errorf("tag %s:%s not found", tag, tagVal)
}

func setFieldValue(field reflect.Value, value string) error {
	fType := field.Type()

	// Handle pointers
	if fType.Kind() == reflect.Ptr {
		elemType := fType.Elem()
		newVal := reflect.New(elemType)
		err := setFieldValue(newVal.Elem(), value)
		if err != nil {
			return err
		}
		field.Set(newVal)
		return nil
	}

	// Handle slices
	if fType.Kind() == reflect.Slice {
		parts := strings.Split(value, ",")
		slice := reflect.MakeSlice(fType, 0, len(parts))
		for _, p := range parts {
			item := reflect.New(fType.Elem()).Elem()
			err := setFieldValue(item, strings.TrimSpace(p))
			if err != nil {
				return err
			}
			slice = reflect.Append(slice, item)
		}
		field.Set(slice)
		return nil
	}

	// Scalar types
	switch fType.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int64:
		iv, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(iv)
	case reflect.Bool:
		bv, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(bv)
	default:
		return fmt.Errorf("unsupported type %s", fType.Kind())
	}
	return nil
}

func extractArgs(raw []string, expectedArgs []Arg) (pos []string, flags map[string]bool, named map[string]string) {
	flags = make(map[string]bool)
	named = make(map[string]string)

	for _, a := range raw {
		foundAsNamed := false

		// Check against defined arguments for "name=" pattern
		// to avoid catching random strings containing "="
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

		// Handle flags and positional arguments
		if strings.HasPrefix(a, flagIdent) {
			// It's a flag (e.g., -v)
			name := strings.TrimPrefix(a, flagIdent)
			flags[name] = true
		} else {
			// It's a positional value
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
		fmt.Println("\n> " + aurora.Bold(aurora.Green(question)).String() + "\n")
		scanner.Scan()
		input := scanner.Text()
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
