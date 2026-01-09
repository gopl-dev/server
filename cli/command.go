package cli

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type arg struct {
	name        string
	description string
	help        []string
	required    bool
	defaultVal  string
	isFlag      bool
	isParam     bool
	typ         string
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
			a.typ = "bool"
		} else {
			// 2) non-flag + "-" prefix => named param
			if strings.HasPrefix(a.name, "-") {
				a.isParam = true
			}
			a.typ = strings.TrimPrefix(f.Type.String(), "*")
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

// prepareHelp processes the help text and extracts argument descriptions.
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

		for name, ar := range argMap {
			prefix := name + ":"
			if strings.HasPrefix(line, prefix) {
				currentArg = ar
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

// prepareRunner binds arguments to the command struct and returns a Runner instance.
func (c Command) prepareRunner(rawArgs []string) (Runner, error) {
	posArgs, flags, named := extractArgs(rawArgs, c.args)

	val := reflect.New(c.reflectType).Elem()
	filled := make(map[string]bool)

	for k, v := range named {
		if idx, ok := c.structFields[k]; ok {
			if err := setFieldValue(val.Field(idx), v); err != nil {
				return nil, fmt.Errorf("arg %s: %w", k, err)
			}
			filled[k] = true
		}
	}

	for name, present := range flags {
		if idx, ok := c.structFields[name]; ok {
			if err := setFieldValue(val.Field(idx), strconv.FormatBool(present)); err != nil {
				return nil, fmt.Errorf("arg %s: %w", name, err)
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
				return nil, fmt.Errorf("arg %s: %w", a.name, err)
			}
			filled[a.name] = true
			curr++
		}
	}

	for _, a := range c.args {
		if !filled[a.name] {
			if a.required {
				return nil, fmt.Errorf("argument '%s' is required. use '? [command]' for help", a.name)
			}
			if a.defaultVal != "" {
				idx := c.structFields[a.name]
				if err := setFieldValue(val.Field(idx), a.defaultVal); err != nil {
					return nil, fmt.Errorf("arg %s: %w", a.name, err)
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

	paramNames := make(map[string]struct{})
	flagNames := make(map[string]struct{})

	for _, a := range args {
		if a.isParam {
			paramNames[a.name] = struct{}{}
		}
		if a.isFlag {
			flagNames[a.name] = struct{}{}
		}
	}

	for _, tok := range raw {
		// Named params: "{argName}=" must match a registered param name.
		if eq := strings.IndexByte(tok, '='); eq >= 0 {
			key := tok[:eq]
			if _, ok := paramNames[key]; ok {
				named[key] = tok[eq+1:]
				continue
			}
		}

		// Flags: token must match a registered flag name exactly.
		if _, ok := flagNames[tok]; ok {
			flags[tok] = true
			continue
		}

		// Otherwise positional value
		pos = append(pos, tok)
	}

	return
}

// setFieldValue assigns a string value to a reflected field, handling type conversion.
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
			return fmt.Errorf("expecting integer, got '%s'", value)
		}
		field.SetInt(v)
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)
	case reflect.Slice:
		if ft.Elem().Kind() == reflect.String {
			parts := strings.Split(value, ",")
			slice := reflect.MakeSlice(ft, len(parts), len(parts))
			for i, part := range parts {
				slice.Index(i).SetString(strings.TrimSpace(part))
			}
			field.Set(slice)
		} else {
			return fmt.Errorf("unsupported slice type %s", ft.Elem().Kind())
		}
	default:
		return fmt.Errorf("unsupported type %s", ft.Kind())
	}

	return nil
}

// validateCommand checks the command configuration for errors.
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
